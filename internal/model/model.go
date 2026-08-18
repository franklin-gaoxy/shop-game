package model

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

// 仓库存储类型
const (
	StorageNormal = 1 // 普通仓库
	StorageCold   = 2 // 冷藏仓库
	StorageBoth   = 3 // 两者皆可
)

// 交易方向
const (
	DirectionExpense = 1 // 支出
	DirectionIncome  = 2 // 收入
)

// 交易类型
const (
	TradeTypeBuy       = "buy"       // 买入
	TradeTypeSell      = "sell"      // 卖出
	TradeTypeExpire    = "expire"    // 过期损耗
	TradeTypeWarehouse = "warehouse" // 购买仓库空间
)

// 过期时间常量
const (
	ExpirePermanent = -1 // 永久
	ExpireNA        = 0  // 不适用（如冷藏商品存储普通仓库）
)

// User 用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex" json:"username"`
	Password  string    `gorm:"size:32" json:"-"` // md5
	Money     float64   `json:"money"`
	Day       int       `gorm:"column:day" json:"day"`
	IsAdmin   bool      `json:"is_admin"`
	KeyID     uint      `gorm:"column:key_id" json:"key_id"` // 注册时使用的密钥
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// RegKey 注册密钥表
type RegKey struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"column:key;size:10;uniqueIndex" json:"key"`
	MaxUses   int       `json:"max_uses"`
	UsedCount int       `json:"used_count"`
	CreatedBy string    `gorm:"size:64" json:"created_by"`
	Status    int       `json:"status"` // 1 启用 0 禁用
	CreatedAt time.Time `json:"created_at"`
}

func (RegKey) TableName() string { return "reg_keys" }

// Product 商品元数据表
type Product struct {
	ID               uint    `gorm:"primaryKey" json:"id"`
	Name             string  `gorm:"size:64;uniqueIndex" json:"name"`
	Category         string  `gorm:"size:32;index" json:"category"` // 商品种类（食物/材料/科技等）
	StorageType      int     `gorm:"column:storage_type" json:"storage_type"`
	MinPrice         float64 `gorm:"column:min_price" json:"min_price"`
	MaxPrice         float64 `gorm:"column:max_price" json:"max_price"`
	Size             int64   `json:"size"`                                                // 单位商品占用仓库大小
	NormalExpireDays int     `gorm:"column:normal_expire_days" json:"normal_expire_days"` // 普通仓库过期天数，-1 永久
	ColdExpireDays   int     `gorm:"column:cold_expire_days" json:"cold_expire_days"`     // 冷藏仓库过期天数，-1 永久
	CritMin          float64 `gorm:"column:crit_min" json:"crit_min"`                     // 暴击涨幅最小值(%)
	CritMax          float64 `gorm:"column:crit_max" json:"crit_max"`                     // 暴击涨幅最大值(%)
}

func (Product) TableName() string { return "products" }

// CritEvent 暴击事件元数据表
type CritEvent struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Description string `gorm:"size:256" json:"description"`
	TargetType  string `gorm:"size:16" json:"target_type"`  // category=针对种类 / product=针对单一商品
	TargetValue string `gorm:"size:64" json:"target_value"` // 种类名或商品名
	Probability int    `json:"probability"`                 // 发生概率(%)，如 10 表示 10%
}

func (CritEvent) TableName() string { return "crit_events" }

// DailyPrice 每日商品价格表（保证同一天价格一致性）
type DailyPrice struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	UserID      uint    `gorm:"column:user_id;uniqueIndex:idx_user_day_product" json:"user_id"`
	Day         int     `gorm:"column:day;uniqueIndex:idx_user_day_product" json:"day"`
	ProductID   uint    `gorm:"column:product_id;uniqueIndex:idx_user_day_product" json:"product_id"`
	Price       float64 `json:"price"`
	CritApplied bool    `gorm:"column:crit_applied" json:"crit_applied"` // 当日价格是否由暴击事件加成
}

func (DailyPrice) TableName() string { return "daily_prices" }

// WarehouseItem 仓库存储表（同一商品合并记录总价值并计算均价）
type WarehouseItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"column:user_id;uniqueIndex:idx_user_product_storage" json:"user_id"`
	ProductID   uint      `gorm:"column:product_id;uniqueIndex:idx_user_product_storage" json:"product_id"`
	StorageType int       `gorm:"column:storage_type;uniqueIndex:idx_user_product_storage" json:"storage_type"`
	Quantity    int       `json:"quantity"`
	TotalCost   float64   `gorm:"column:total_cost" json:"total_cost"` // 总价值
	AvgPrice    float64   `gorm:"column:avg_price" json:"avg_price"`   // 均价 = 总价值 / 数量
	BuyDay      int       `gorm:"column:buy_day" json:"buy_day"`       // 首次买入天数
	ExpireDay   int       `gorm:"column:expire_day" json:"expire_day"` // 过期天数，-1 永久
	UpdatedAt   time.Time `json:"updated_at"`
}

func (WarehouseItem) TableName() string { return "warehouse_items" }

// Transaction 交易记录表
type Transaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"column:user_id;index" json:"user_id"`
	Day          int       `gorm:"column:day" json:"day"`
	Type         string    `gorm:"size:16" json:"type"` // buy/sell/expire/warehouse
	Direction    int       `json:"direction"`           // 1 支出 2 收入
	ProductID    uint      `gorm:"column:product_id" json:"product_id"`
	ProductName  string    `gorm:"size:64" json:"product_name"`
	Quantity     int       `json:"quantity"`
	UnitPrice    float64   `gorm:"column:unit_price" json:"unit_price"`
	Amount       float64   `json:"amount"`                                    // 支出或收入金额
	BalanceAfter float64   `gorm:"column:balance_after" json:"balance_after"` // 交易后余额
	CreatedAt    time.Time `json:"created_at"`
}

func (Transaction) TableName() string { return "transactions" }

// WarehousePurchase 特殊商城仓库空间购买记录表
type WarehousePurchase struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"column:user_id;index" json:"user_id"`
	StorageType int       `gorm:"column:storage_type" json:"storage_type"`
	Size        int64     `json:"size"`
	Months      int       `json:"months"`                              // 购买月数（1-12，每月按 30 天）
	UnitPrice   float64   `gorm:"column:unit_price" json:"unit_price"` // 每单位空间每月价格
	StartDay    int       `gorm:"column:start_day" json:"start_day"`
	ExpireDay   int       `gorm:"column:expire_day" json:"expire_day"`
	CreatedAt   time.Time `json:"created_at"`
}

func (WarehousePurchase) TableName() string { return "warehouse_purchases" }

// Setting 系统配置表
type Setting struct {
	K string `gorm:"column:k;primaryKey;size:64" json:"k"`
	V string `gorm:"column:v;size:255" json:"v"`
}

func (Setting) TableName() string { return "settings" }

// MD5Hash 计算字符串 md5（用于密码存储与校验）
func MD5Hash(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
