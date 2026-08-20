// Package mysql 基于 GORM + MySQL 实现 database.Database 接口。
package mysql

import (
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strconv"

	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"trade_game/internal/config"
	"trade_game/internal/database"
	"trade_game/internal/model"
)

// settings 表中的键
const (
	SettingInitialized          = "initialized"
	SettingDefaultMoney         = "default_money"
	SettingDefaultWarehouse     = "default_warehouse"
	SettingDefaultColdStorage   = "default_cold_storage"
	SettingWarehousePriceNormal = "warehouse_price_normal"
	SettingWarehousePriceCold   = "warehouse_price_cold"
)

const keyCharset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"

func init() {
	database.RegisterFactory("mysql", func(cfg config.DatabaseConfig) (database.Database, error) {
		return New(cfg)
	})
}

// MySQL 实现
type MySQL struct {
	db *gorm.DB
}

var _ database.Database = (*MySQL)(nil)

// New 连接 MySQL
func New(cfg config.DatabaseConfig) (*MySQL, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := gorm.Open(gmysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	return &MySQL{db: db}, nil
}

func (m *MySQL) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (m *MySQL) Ping() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// ---------- 初始化 ----------

// IsInitialized 判断是否已执行过初始化
func (m *MySQL) IsInitialized() (bool, error) {
	// settings 表不存在说明从未初始化（连接错误等则直接返回错误，避免误判后重建删库）
	if !m.db.Migrator().HasTable(&model.Setting{}) {
		return false, nil
	}
	var s model.Setting
	err := m.db.Where("k = ?", SettingInitialized).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return s.V == "true", nil
}

// Initialize 执行建表 SQL；若已初始化且未 force 则返回 ErrAlreadyInitialized
func (m *MySQL) Initialize(sqlFile string, force bool) (bool, error) {
	initialized, err := m.IsInitialized()
	if err != nil {
		return false, err
	}
	if initialized && !force {
		return false, database.ErrAlreadyInitialized
	}
	content, err := os.ReadFile(sqlFile)
	if err != nil {
		return false, fmt.Errorf("读取 SQL 文件失败: %w", err)
	}
	if err := m.db.Exec(string(content)).Error; err != nil {
		return false, fmt.Errorf("执行 SQL 失败: %w", err)
	}
	return true, nil
}

// GenerateKey 生成一个未被占用的 10 位随机密钥
func (m *MySQL) GenerateKey() (string, error) {
	return randomUniqueKeyTx(m.db)
}

func randomUniqueKeyTx(tx *gorm.DB) (string, error) {
	for i := 0; i < 10; i++ {
		k := randomString(10)
		var cnt int64
		if err := tx.Model(&model.RegKey{}).Where("`key` = ?", k).Count(&cnt).Error; err != nil {
			return "", err
		}
		if cnt == 0 {
			return k, nil
		}
	}
	return "", errors.New("生成密钥失败，请重试")
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(keyCharset))))
		if err != nil {
			b[i] = keyCharset[rand.Intn(len(keyCharset))]
			continue
		}
		b[i] = keyCharset[idx.Int64()]
	}
	return string(b)
}

// ImportInitData 导入初始化数据：管理员、初始密钥、商品、暴击事件、系统设置
func (m *MySQL) ImportInitData(data *database.InitData, adminUser, adminPassword string) (string, error) {
	if adminUser == "" || adminPassword == "" {
		return "", errors.New("管理员用户名或密码为空，请检查 config.yaml 中 platform 配置")
	}
	if len(data.Products) == 0 {
		return "", errors.New("初始化数据中没有商品，请检查数据 yaml 文件")
	}
	for i, p := range data.Products {
		if p.Name == "" {
			return "", fmt.Errorf("第 %d 个商品缺少 name", i+1)
		}
		if p.Category == "" {
			return "", fmt.Errorf("商品 %s 缺少 category", p.Name)
		}
		if parseStorageType(p.StorageType) == 0 {
			return "", fmt.Errorf("商品 %s 的 storage_type 必须为 normal/cold/both", p.Name)
		}
		if p.MinPrice <= 0 || p.MaxPrice < p.MinPrice {
			return "", fmt.Errorf("商品 %s 的价格区间不合法", p.Name)
		}
		if p.Size <= 0 {
			return "", fmt.Errorf("商品 %s 的 size 必须大于 0", p.Name)
		}
		if p.CritMin < 0 || p.CritMax < p.CritMin {
			return "", fmt.Errorf("商品 %s 的暴击区间不合法", p.Name)
		}
		if p.MinStock < 0 || p.MaxStock < p.MinStock {
			return "", fmt.Errorf("商品 %s 的存货量区间不合法（需满足 min_stock >= 0 且 max_stock >= min_stock）", p.Name)
		}
	}
	for i, e := range data.CritEvents {
		if e.Description == "" {
			return "", fmt.Errorf("第 %d 个暴击事件缺少 description", i+1)
		}
		if e.TargetType != "category" && e.TargetType != "product" {
			return "", fmt.Errorf("暴击事件 %q 的 target_type 必须为 category/product", e.Description)
		}
		if e.TargetValue == "" {
			return "", fmt.Errorf("暴击事件 %q 缺少 target_value", e.Description)
		}
		if e.Probability < 0 || e.Probability > 100 {
			return "", fmt.Errorf("暴击事件 %q 的 probability 必须在 0-100 之间", e.Description)
		}
	}

	d := data.Defaults
	if d.DefaultMoney <= 0 {
		d.DefaultMoney = 1000000
	}
	if d.DefaultWarehouse <= 0 {
		d.DefaultWarehouse = 1000
	}
	if d.DefaultColdStorage <= 0 {
		d.DefaultColdStorage = 1000
	}
	if d.WarehousePriceNormal <= 0 {
		d.WarehousePriceNormal = 10
	}
	if d.WarehousePriceCold <= 0 {
		d.WarehousePriceCold = 20
	}
	if d.InitialKeyMaxUses <= 0 {
		d.InitialKeyMaxUses = 10
	}

	var initialKey string
	err := m.db.Transaction(func(tx *gorm.DB) error {
		settings := []model.Setting{
			{K: SettingInitialized, V: "true"},
			{K: SettingDefaultMoney, V: fmt.Sprintf("%g", d.DefaultMoney)},
			{K: SettingDefaultWarehouse, V: strconv.FormatInt(d.DefaultWarehouse, 10)},
			{K: SettingDefaultColdStorage, V: strconv.FormatInt(d.DefaultColdStorage, 10)},
			{K: SettingWarehousePriceNormal, V: fmt.Sprintf("%g", d.WarehousePriceNormal)},
			{K: SettingWarehousePriceCold, V: fmt.Sprintf("%g", d.WarehousePriceCold)},
		}
		for _, s := range settings {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&s).Error; err != nil {
				return err
			}
		}

		admin := model.User{Username: adminUser, Password: model.MD5Hash(adminPassword), Money: d.DefaultMoney, Day: 1, IsAdmin: true}
		if err := tx.Create(&admin).Error; err != nil {
			return fmt.Errorf("创建管理员失败: %w", err)
		}
		// 为管理员生成第 1 天价格
		if err := ensureDayPricesTx(tx, admin.ID, 1); err != nil {
			return err
		}

		keyStr := d.InitialKey
		if keyStr == "" {
			k, err := randomUniqueKeyTx(tx)
			if err != nil {
				return err
			}
			keyStr = k
		}
		regKey := model.RegKey{Key: keyStr, MaxUses: d.InitialKeyMaxUses, CreatedBy: adminUser, Status: 1}
		if err := tx.Create(&regKey).Error; err != nil {
			return fmt.Errorf("创建初始密钥失败: %w", err)
		}
		initialKey = keyStr

		for _, p := range data.Products {
			np := model.Product{
				Name: p.Name, Category: p.Category, StorageType: parseStorageType(p.StorageType),
				MinPrice: p.MinPrice, MaxPrice: p.MaxPrice, Size: p.Size,
				NormalExpireDays: p.NormalExpireDays, ColdExpireDays: p.ColdExpireDays,
				CritMin: p.CritMin, CritMax: p.CritMax,
				MaxStock: p.MaxStock, MinStock: p.MinStock,
			}
			if err := tx.Create(&np).Error; err != nil {
				return fmt.Errorf("导入商品 %s 失败: %w", p.Name, err)
			}
		}

		for _, e := range data.CritEvents {
			ne := model.CritEvent{Description: e.Description, TargetType: e.TargetType, TargetValue: e.TargetValue, Probability: e.Probability}
			if err := tx.Create(&ne).Error; err != nil {
				return fmt.Errorf("导入暴击事件 %q 失败: %w", e.Description, err)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return initialKey, nil
}

// ---------- 用户与密钥 ----------

func (m *MySQL) LoginCheck(username, passwordMD5 string) (*model.User, error) {
	var user model.User
	err := m.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("用户名或密码错误")
	}
	if err != nil {
		return nil, err
	}
	if user.Password != passwordMD5 {
		return nil, errors.New("用户名或密码错误")
	}
	if user.Status == 0 {
		return nil, errors.New("账号已被禁用，请联系管理员")
	}
	return &user, nil
}

func (m *MySQL) GetUser(id uint) (*model.User, error) {
	var user model.User
	if err := m.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// Register 注册用户：校验用户名唯一、密钥有效且未达使用上限，创建用户并生成第 1 天价格
func (m *MySQL) Register(username, passwordMD5, regKey string) (*model.User, error) {
	if username == "" || passwordMD5 == "" || regKey == "" {
		return nil, errors.New("用户名、密码、密钥均不能为空")
	}
	var user *model.User
	err := m.db.Transaction(func(tx *gorm.DB) error {
		var cnt int64
		if err := tx.Model(&model.User{}).Where("username = ?", username).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return errors.New("用户名已存在")
		}

		var k model.RegKey
		err := tx.Where("`key` = ?", regKey).First(&k).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("注册密钥不存在")
		}
		if err != nil {
			return err
		}
		if k.Status != 1 {
			return errors.New("注册密钥已被禁用")
		}
		if k.UsedCount >= k.MaxUses {
			return errors.New("注册密钥使用次数已达上限")
		}

		money, _ := strconv.ParseFloat(getSetting(tx, SettingDefaultMoney, "1000000"), 64)
		user = &model.User{Username: username, Password: passwordMD5, Money: money, Day: 1, KeyID: k.ID}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.RegKey{}).Where("id = ?", k.ID).Update("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
			return err
		}

		// 生成第 1 天价格（不触发暴击）
		if err := ensureDayPricesTx(tx, user.ID, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *MySQL) CreateRegKey(maxUses int, createdBy string) (*model.RegKey, error) {
	if maxUses <= 0 {
		maxUses = 10
	}
	key, err := m.GenerateKey()
	if err != nil {
		return nil, err
	}
	k := &model.RegKey{Key: key, MaxUses: maxUses, CreatedBy: createdBy, Status: 1}
	if err := m.db.Create(k).Error; err != nil {
		return nil, err
	}
	return k, nil
}

func (m *MySQL) ListRegKeys() ([]model.RegKey, error) {
	var keys []model.RegKey
	if err := m.db.Order("id DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// DeleteRegKey 删除密钥：已被用户用于注册的密钥不允许删除（避免用户注册来源悬空）
func (m *MySQL) DeleteRegKey(id uint) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var k model.RegKey
		if err := tx.First(&k, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("密钥不存在")
			}
			return err
		}
		var cnt int64
		if err := tx.Model(&model.User{}).Where("key_id = ?", id).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return fmt.Errorf("该密钥已被 %d 个用户用于注册，无法删除", cnt)
		}
		return tx.Delete(&model.RegKey{}, id).Error
	})
}

// ListKeyUsers 查询使用某密钥注册的用户列表
func (m *MySQL) ListKeyUsers(keyID uint) ([]model.User, error) {
	var k model.RegKey
	if err := m.db.First(&k, keyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("密钥不存在")
		}
		return nil, err
	}
	var users []model.User
	if err := m.db.Where("key_id = ?", keyID).Order("id").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ---------- 用户管理（admin） ----------

// ListUsers 分页查询全部用户（单条 SQL 分页，无逐用户检查）
func (m *MySQL) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int64
	if err := m.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	if err := m.db.Order("id").Limit(pageSize).Offset((page - 1) * pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// checkUserTarget 校验管理员操作目标：不得操作自己、不得操作其他管理员
func checkUserTarget(tx *gorm.DB, operatorID, targetID uint) (*model.User, error) {
	if operatorID == targetID {
		return nil, errors.New("不能对自己执行该操作")
	}
	var target model.User
	if err := tx.First(&target, targetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	if target.IsAdmin {
		return nil, errors.New("不能对管理员账号执行该操作")
	}
	return &target, nil
}

// SetUserStatus 禁用/启用用户
func (m *MySQL) SetUserStatus(operatorID, targetID uint, status int) error {
	if status != 0 && status != 1 {
		return errors.New("status 仅支持 0=禁用 1=启用")
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		target, err := checkUserTarget(tx, operatorID, targetID)
		if err != nil {
			return err
		}
		if target.Status == status {
			if status == 0 {
				return errors.New("该用户已是禁用状态")
			}
			return errors.New("该用户已是启用状态")
		}
		return tx.Model(&model.User{}).Where("id = ?", targetID).Update("status", status).Error
	})
}

// DeleteUser 删除用户及其全部关联数据（仓库、每日价格、仓库购买、交易记录）
func (m *MySQL) DeleteUser(operatorID, targetID uint) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		if _, err := checkUserTarget(tx, operatorID, targetID); err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", targetID).Delete(&model.WarehouseItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", targetID).Delete(&model.DailyPrice{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", targetID).Delete(&model.WarehousePurchase{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", targetID).Delete(&model.Transaction{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, targetID).Error
	})
}

// ---------- 商品与交易 ----------

func (m *MySQL) ListProducts() ([]model.Product, error) {
	var products []model.Product
	if err := m.db.Order("id").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (m *MySQL) GetTodayPrices(userID uint) (*database.ProductPriceResult, error) {
	var user model.User
	if err := m.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	// 懒生成：若当日价格缺失（如初始化时创建的管理员）则补齐
	if err := ensureDayPricesTx(m.db, userID, user.Day); err != nil {
		return nil, err
	}
	products, err := m.ListProducts()
	if err != nil {
		return nil, err
	}
	var prices []model.DailyPrice
	if err := m.db.Where("user_id = ? AND day = ?", userID, user.Day).Find(&prices).Error; err != nil {
		return nil, err
	}
	pm := map[uint]model.DailyPrice{}
	for _, p := range prices {
		pm[p.ProductID] = p
	}
	// 当日已买入数量（用于计算剩余限购）
	bought, err := boughtTodayMap(m.db, userID, user.Day)
	if err != nil {
		return nil, err
	}
	res := &database.ProductPriceResult{Day: user.Day, Products: []database.ProductPrice{}}
	for _, p := range products {
		dp := pm[p.ID]
		pp := database.ProductPrice{Product: p, Price: dp.Price, CritApplied: dp.CritApplied, StockLimit: dp.StockLimit, StockBought: bought[p.ID]}
		if dp.StockLimit > 0 {
			pp.StockRemaining = dp.StockLimit - bought[p.ID]
			if pp.StockRemaining < 0 {
				pp.StockRemaining = 0
			}
		}
		res.Products = append(res.Products, pp)
	}
	return res, nil
}

// Buy 买入商品：校验仓库类型、余额、仓库容量，合并入库并记录交易
func (m *MySQL) Buy(userID, productID uint, quantity, storageType int) (*database.TradeResult, error) {
	if quantity <= 0 {
		return nil, errors.New("购买数量必须大于 0")
	}
	res := &database.TradeResult{}
	err := m.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := lockUser(tx, &user, userID); err != nil {
			return err
		}
		if err := ensureDayPricesTx(tx, userID, user.Day); err != nil {
			return err
		}
		var product model.Product
		if err := tx.First(&product, productID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("商品不存在")
			}
			return err
		}

		// 确定存储仓库
		st := storageType
		switch product.StorageType {
		case model.StorageNormal:
			if st != 0 && st != model.StorageNormal {
				return errors.New("该商品只能存放在普通仓库")
			}
			st = model.StorageNormal
		case model.StorageCold:
			if st != 0 && st != model.StorageCold {
				return errors.New("该商品只能存放在冷藏仓库")
			}
			st = model.StorageCold
		case model.StorageBoth:
			if st != model.StorageNormal && st != model.StorageCold {
				return errors.New("该商品支持普通/冷藏两种仓库，请指定存储仓库（storage_type: 1=普通 2=冷藏）")
			}
		}

		var dp model.DailyPrice
		if err := tx.Where("user_id = ? AND day = ? AND product_id = ?", userID, user.Day, productID).First(&dp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("今日价格尚未生成")
			}
			return err
		}

		// 当日限购校验：当日已购 + 本次数量不能超过当日限购上限（stock_limit 为 0 表示不限，兼容历史数据）
		if dp.StockLimit > 0 {
			bought, err := boughtTodayMap(tx, userID, user.Day)
			if err != nil {
				return err
			}
			if remaining := dp.StockLimit - bought[productID]; bought[productID]+quantity > dp.StockLimit {
				if remaining < 0 {
					remaining = 0
				}
				return fmt.Errorf("超出当日限购数量：今日限购 %d，已购 %d，剩余可购 %d，明天再来吧", dp.StockLimit, bought[productID], remaining)
			}
		}

		total := round2(dp.Price * float64(quantity))
		if user.Money < total {
			return fmt.Errorf("余额不足：本次需 %.2f，当前余额 %.2f", total, user.Money)
		}

		need := product.Size * int64(quantity)
		capacity, used, err := capacityAndUsed(tx, userID, user.Day, st)
		if err != nil {
			return err
		}
		if used+need > capacity {
			return fmt.Errorf("仓库空间不足：还需 %d 空间，当前剩余 %d，可前往特殊商城扩容", need, capacity-used)
		}

		newMoney := round2(user.Money - total)
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Update("money", newMoney).Error; err != nil {
			return err
		}

		// 过期天数
		expireDay := model.ExpirePermanent
		if st == model.StorageNormal {
			if product.NormalExpireDays != model.ExpirePermanent {
				expireDay = user.Day + product.NormalExpireDays
			}
		} else {
			if product.ColdExpireDays != model.ExpirePermanent {
				expireDay = user.Day + product.ColdExpireDays
			}
		}

		// 合并入库（同一商品累计总价值并计算均价）
		var item model.WarehouseItem
		err = tx.Where("user_id = ? AND product_id = ? AND storage_type = ?", userID, productID, st).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = model.WarehouseItem{
				UserID: userID, ProductID: productID, StorageType: st,
				Quantity: quantity, TotalCost: total, AvgPrice: round2(total / float64(quantity)),
				BuyDay: user.Day, ExpireDay: expireDay,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			item.Quantity += quantity
			item.TotalCost = round2(item.TotalCost + total)
			item.AvgPrice = round2(item.TotalCost / float64(item.Quantity))
			item.ExpireDay = minExpireDay(item.ExpireDay, expireDay)
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
		}

		tr := model.Transaction{
			UserID: userID, Day: user.Day, Type: model.TradeTypeBuy, Direction: model.DirectionExpense,
			ProductID: productID, ProductName: product.Name, Quantity: quantity,
			UnitPrice: dp.Price, Amount: total, BalanceAfter: newMoney,
		}
		if err := tx.Create(&tr).Error; err != nil {
			return err
		}

		res.ProductID = productID
		res.ProductName = product.Name
		res.StorageType = st
		res.Quantity = quantity
		res.UnitPrice = dp.Price
		res.Amount = total
		res.Balance = newMoney
		res.Day = user.Day
		res.ExpireDay = item.ExpireDay
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Sell 卖出商品：按今日价格卖出，可跨存储行扣减库存
func (m *MySQL) Sell(userID, productID uint, quantity, storageType int) (*database.TradeResult, error) {
	if quantity <= 0 {
		return nil, errors.New("卖出数量必须大于 0")
	}
	res := &database.TradeResult{}
	err := m.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := lockUser(tx, &user, userID); err != nil {
			return err
		}
		if err := ensureDayPricesTx(tx, userID, user.Day); err != nil {
			return err
		}
		var product model.Product
		if err := tx.First(&product, productID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("商品不存在")
			}
			return err
		}

		var dp model.DailyPrice
		if err := tx.Where("user_id = ? AND day = ? AND product_id = ?", userID, user.Day, productID).First(&dp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("今日价格尚未生成")
			}
			return err
		}

		cond := map[string]interface{}{"user_id": userID, "product_id": productID}
		if storageType != 0 {
			cond["storage_type"] = storageType
		}
		var items []model.WarehouseItem
		if err := tx.Where(cond).Order("storage_type").Find(&items).Error; err != nil {
			return err
		}

		remaining := quantity
		for i := range items {
			if remaining <= 0 {
				break
			}
			it := &items[i]
			if it.Quantity <= 0 {
				continue
			}
			take := it.Quantity
			if take > remaining {
				take = remaining
			}
			remaining -= take
			if it.Quantity == take {
				if err := tx.Delete(&model.WarehouseItem{}, it.ID).Error; err != nil {
					return err
				}
			} else {
				rest := it.Quantity - take
				updates := map[string]interface{}{
					"quantity":   rest,
					"total_cost": round2(it.AvgPrice * float64(rest)),
				}
				if err := tx.Model(&model.WarehouseItem{}).Where("id = ?", it.ID).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
		if remaining > 0 {
			return errors.New("仓库中该商品数量不足")
		}

		income := round2(dp.Price * float64(quantity))
		newMoney := round2(user.Money + income)
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Update("money", newMoney).Error; err != nil {
			return err
		}

		tr := model.Transaction{
			UserID: userID, Day: user.Day, Type: model.TradeTypeSell, Direction: model.DirectionIncome,
			ProductID: productID, ProductName: product.Name, Quantity: quantity,
			UnitPrice: dp.Price, Amount: income, BalanceAfter: newMoney,
		}
		if err := tx.Create(&tr).Error; err != nil {
			return err
		}

		res.ProductID = productID
		res.ProductName = product.Name
		res.StorageType = storageType
		res.Quantity = quantity
		res.UnitPrice = dp.Price
		res.Amount = income
		res.Balance = newMoney
		res.Day = user.Day
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// AdvanceDay 进入明天：天数+1、生成随机价格、判定暴击事件、清理过期商品与到期仓库空间
func (m *MySQL) AdvanceDay(userID uint) (*database.AdvanceResult, error) {
	res := &database.AdvanceResult{
		TriggeredCrits: []database.CritTriggerInfo{},
		ExpiredItems:   []database.ExpiredItemInfo{},
	}
	err := m.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := lockUser(tx, &user, userID); err != nil {
			return err
		}
		newDay := user.Day + 1

		var products []model.Product
		if err := tx.Order("id").Find(&products).Error; err != nil {
			return err
		}
		var events []model.CritEvent
		if err := tx.Order("id").Find(&events).Error; err != nil {
			return err
		}

		// 1. 在 [min,max] 区间为每个商品生成随机价格，并在 [min_stock, max_stock] 区间生成当日限购数量
		prices := map[uint]float64{}
		stockLimits := map[uint]int{}
		critHit := map[uint]bool{} // 当日被暴击加成的商品
		for _, p := range products {
			prices[p.ID] = round2(randRange(p.MinPrice, p.MaxPrice))
			stockLimits[p.ID] = randIntRange(p.MinStock, p.MaxStock)
		}

		// 2. 暴击事件判定：生成 0-99 随机数，小于概率则触发
		for _, e := range events {
			if e.Probability <= 0 || rand.Intn(100) >= e.Probability {
				continue
			}
			info := database.CritTriggerInfo{Description: e.Description, Products: []string{}}
			for i := range products {
				p := &products[i]
				var hit bool
				switch e.TargetType {
				case "category":
					hit = p.Category == e.TargetValue
				case "product":
					hit = p.Name == e.TargetValue
				}
				if !hit {
					continue
				}
				// 涨幅在商品自身 [crit_min, crit_max] 区间取随机值（百分比）
				inc := randRange(p.CritMin, p.CritMax)
				prices[p.ID] = round2(prices[p.ID] * (1 + inc/100.0))
				critHit[p.ID] = true
				info.Products = append(info.Products, p.Name)
			}
			if len(info.Products) > 0 {
				res.TriggeredCrits = append(res.TriggeredCrits, info)
			}
		}

		// 3. 写入每日价格表（含暴击标记与当日限购数量）
		res.Prices = []database.ProductPrice{}
		for _, p := range products {
			dp := model.DailyPrice{UserID: userID, Day: newDay, ProductID: p.ID, Price: prices[p.ID], StockLimit: stockLimits[p.ID], CritApplied: critHit[p.ID]}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&dp).Error; err != nil {
				return err
			}
			res.Prices = append(res.Prices, database.ProductPrice{
				Product: p, Price: prices[p.ID], CritApplied: critHit[p.ID],
				StockLimit: stockLimits[p.ID], StockBought: 0, StockRemaining: stockLimits[p.ID],
			})
		}

		// 4. 清理过期商品（expire_day 为最后有效天，进入新的一天后失效）
		productName := func(id uint) string {
			for _, p := range products {
				if p.ID == id {
					return p.Name
				}
			}
			return ""
		}
		var expired []model.WarehouseItem
		if err := tx.Where("user_id = ? AND expire_day <> ? AND expire_day < ?", userID, model.ExpirePermanent, newDay).Find(&expired).Error; err != nil {
			return err
		}
		for _, it := range expired {
			tr := model.Transaction{
				UserID: userID, Day: newDay, Type: model.TradeTypeExpire, Direction: model.DirectionExpense,
				ProductID: it.ProductID, ProductName: productName(it.ProductID), Quantity: it.Quantity,
				UnitPrice: it.AvgPrice, Amount: 0, BalanceAfter: user.Money,
			}
			if err := tx.Create(&tr).Error; err != nil {
				return err
			}
			if err := tx.Delete(&model.WarehouseItem{}, it.ID).Error; err != nil {
				return err
			}
			res.ExpiredItems = append(res.ExpiredItems, database.ExpiredItemInfo{
				ProductID: it.ProductID, ProductName: tr.ProductName, Quantity: it.Quantity,
				Value: round2(it.AvgPrice * float64(it.Quantity)),
			})
		}

		// 5. 清理到期的仓库空间购买记录
		dl := tx.Where("user_id = ? AND expire_day < ?", userID, newDay).Delete(&model.WarehousePurchase{})
		if dl.Error != nil {
			return dl.Error
		}
		res.ExpiredPurchases = dl.RowsAffected

		// 6. 用户天数 +1
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Update("day", newDay).Error; err != nil {
			return err
		}
		res.Day = newDay
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ---------- 仓库 ----------

func (m *MySQL) ListWarehouseItems(userID uint) (*database.WarehouseItemsResult, error) {
	var user model.User
	if err := m.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	var items []model.WarehouseItem
	if err := m.db.Where("user_id = ?", userID).Order("product_id, storage_type").Find(&items).Error; err != nil {
		return nil, err
	}
	var products []model.Product
	if err := m.db.Find(&products).Error; err != nil {
		return nil, err
	}
	pmap := map[uint]model.Product{}
	for _, p := range products {
		pmap[p.ID] = p
	}
	var prices []model.DailyPrice
	if err := m.db.Where("user_id = ? AND day = ?", userID, user.Day).Find(&prices).Error; err != nil {
		return nil, err
	}
	prMap := map[uint]float64{}
	for _, p := range prices {
		prMap[p.ProductID] = p.Price
	}

	res := &database.WarehouseItemsResult{Day: user.Day, Items: []database.WarehouseItemInfo{}}
	for _, it := range items {
		p := pmap[it.ProductID]
		today := prMap[it.ProductID]
		res.Items = append(res.Items, database.WarehouseItemInfo{
			WarehouseItem: it,
			ProductName:   p.Name,
			Category:      p.Category,
			TodayPrice:    today,
			TrendUp:       today > it.AvgPrice,
			Profit:        round2((today - it.AvgPrice) * float64(it.Quantity)),
		})
	}
	return res, nil
}

func (m *MySQL) GetWarehouseSpace(userID uint) (*database.WarehouseSpace, error) {
	var user model.User
	if err := m.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	space := &database.WarehouseSpace{Day: user.Day, Purchases: []model.WarehousePurchase{}}
	for _, st := range []int{model.StorageNormal, model.StorageCold} {
		capacity, used, err := capacityAndUsed(m.db, userID, user.Day, st)
		if err != nil {
			return nil, err
		}
		s := database.StorageSpace{Capacity: capacity, Used: used, Free: capacity - used}
		if st == model.StorageNormal {
			space.Normal = s
		} else {
			space.Cold = s
		}
	}
	space.UnitPriceNormal, _ = strconv.ParseFloat(getSetting(m.db, SettingWarehousePriceNormal, "10"), 64)
	space.UnitPriceCold, _ = strconv.ParseFloat(getSetting(m.db, SettingWarehousePriceCold, "20"), 64)
	if err := m.db.Where("user_id = ?", userID).Order("id DESC").Find(&space.Purchases).Error; err != nil {
		return nil, err
	}
	return space, nil
}

// BuyWarehouseSpace 特殊商城购买仓库空间：时长 1-12 个月（每月按 30 天计）
func (m *MySQL) BuyWarehouseSpace(userID uint, storageType int, size int64, months int) (*model.Transaction, error) {
	if storageType != model.StorageNormal && storageType != model.StorageCold {
		return nil, errors.New("仓库类型错误：1=普通仓库 2=冷藏仓库")
	}
	if size <= 0 {
		return nil, errors.New("购买空间大小必须大于 0")
	}
	if months < 1 || months > 12 {
		return nil, errors.New("购买时长最短 1 个月（30 天），最长 12 个月")
	}
	var result *model.Transaction
	err := m.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := lockUser(tx, &user, userID); err != nil {
			return err
		}
		var unitPrice float64
		var name string
		if storageType == model.StorageNormal {
			unitPrice, _ = strconv.ParseFloat(getSetting(tx, SettingWarehousePriceNormal, "10"), 64)
			name = "普通仓库"
		} else {
			unitPrice, _ = strconv.ParseFloat(getSetting(tx, SettingWarehousePriceCold, "20"), 64)
			name = "冷藏仓库"
		}
		cost := round2(unitPrice * float64(size) * float64(months))
		if user.Money < cost {
			return fmt.Errorf("余额不足：本次需 %.2f，当前余额 %.2f", cost, user.Money)
		}

		purchase := model.WarehousePurchase{
			UserID: userID, StorageType: storageType, Size: size, Months: months, UnitPrice: unitPrice,
			StartDay: user.Day, ExpireDay: user.Day + months*30,
		}
		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}

		newMoney := round2(user.Money - cost)
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Update("money", newMoney).Error; err != nil {
			return err
		}

		tr := &model.Transaction{
			UserID: userID, Day: user.Day, Type: model.TradeTypeWarehouse, Direction: model.DirectionExpense,
			ProductID: 0, ProductName: fmt.Sprintf("%s空间(%d个月)", name, months), Quantity: int(size),
			UnitPrice: round2(cost / float64(size)), Amount: cost, BalanceAfter: newMoney,
		}
		if err := tx.Create(tr).Error; err != nil {
			return err
		}
		result = tr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---------- 交易记录 ----------

func (m *MySQL) ListTransactions(userID uint, page, pageSize int) ([]model.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int64
	if err := m.db.Model(&model.Transaction{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Transaction
	if err := m.db.Where("user_id = ?", userID).Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ---------- 内部工具 ----------

// ensureDayPricesTx 确保某用户某天的每日价格存在：缺失则为全部商品随机生成价格与当日限购数量（不触发暴击）。
// 用于初始化建管理员、注册新用户，以及查询/交易时的懒生成补齐。
func ensureDayPricesTx(tx *gorm.DB, userID uint, day int) error {
	var cnt int64
	if err := tx.Model(&model.DailyPrice{}).Where("user_id = ? AND day = ?", userID, day).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	var products []model.Product
	if err := tx.Order("id").Find(&products).Error; err != nil {
		return err
	}
	for _, p := range products {
		dp := model.DailyPrice{
			UserID: userID, Day: day, ProductID: p.ID,
			Price: round2(randRange(p.MinPrice, p.MaxPrice)),
			StockLimit: randIntRange(p.MinStock, p.MaxStock),
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&dp).Error; err != nil {
			return err
		}
	}
	return nil
}

// boughtTodayMap 汇总某用户某天各商品已买入数量（用于计算当日剩余限购）
func boughtTodayMap(tx *gorm.DB, userID uint, day int) (map[uint]int, error) {
	var rows []struct {
		ProductID uint
		Total     int64
	}
	if err := tx.Model(&model.Transaction{}).
		Select("product_id, COALESCE(SUM(quantity), 0) AS total").
		Where("user_id = ? AND day = ? AND type = ?", userID, day, model.TradeTypeBuy).
		Group("product_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := map[uint]int{}
	for _, r := range rows {
		m[r.ProductID] = int(r.Total)
	}
	return m, nil
}

func lockUser(tx *gorm.DB, user *model.User, userID uint) error {
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}
	return nil
}

// capacityAndUsed 计算某类仓库容量（默认 + 有效期内购买）与已用空间
func capacityAndUsed(tx *gorm.DB, userID uint, day, storageType int) (int64, int64, error) {
	defKey := SettingDefaultWarehouse
	if storageType != model.StorageNormal {
		defKey = SettingDefaultColdStorage
	}
	capacity := int64(0)
	if v, ok := getSettingOK(tx, defKey); ok {
		capacity, _ = strconv.ParseInt(v, 10, 64)
	}

	var extra int64
	if err := tx.Model(&model.WarehousePurchase{}).
		Where("user_id = ? AND storage_type = ? AND expire_day >= ?", userID, storageType, day).
		Select("COALESCE(SUM(size), 0)").Scan(&extra).Error; err != nil {
		return 0, 0, err
	}
	capacity += extra

	var used int64
	if err := tx.Table("warehouse_items AS w").
		Select("COALESCE(SUM(p.size * w.quantity), 0)").
		Joins("JOIN products AS p ON p.id = w.product_id").
		Where("w.user_id = ? AND w.storage_type = ?", userID, storageType).
		Scan(&used).Error; err != nil {
		return 0, 0, err
	}
	return capacity, used, nil
}

func getSetting(tx *gorm.DB, key, def string) string {
	if v, ok := getSettingOK(tx, key); ok {
		return v
	}
	return def
}

func getSettingOK(tx *gorm.DB, key string) (string, bool) {
	var s model.Setting
	err := tx.Where("k = ?", key).First(&s).Error
	if err != nil {
		return "", false
	}
	return s.V, true
}

func parseStorageType(s string) int {
	switch s {
	case "normal":
		return model.StorageNormal
	case "cold":
		return model.StorageCold
	case "both":
		return model.StorageBoth
	default:
		return 0
	}
}

// minExpireDay 合并过期天：-1 表示永久（视为无穷大），取两者中更早到期的一个
func minExpireDay(a, b int) int {
	if a == model.ExpirePermanent {
		return b
	}
	if b == model.ExpirePermanent {
		return a
	}
	if a < b {
		return a
	}
	return b
}

func randRange(min, max float64) float64 {
	if max <= min {
		return min
	}
	return min + rand.Float64()*(max-min)
}

// randIntRange 在 [min, max] 闭区间内取随机整数
func randIntRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min+1)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
