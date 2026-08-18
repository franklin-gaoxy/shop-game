package main

import (
	"fmt"
	"strings"
)

// App 命令行客户端应用
type App struct {
	client *Client
	user   *UserInfo
}

// NewApp 创建应用
func NewApp(server string) *App {
	return &App{client: NewClient(server)}
}

// Run 启动客户端：连接检查 → 登录/注册 → 主菜单
func (a *App) Run() error {
	if err := a.client.Health(); err != nil {
		return fmt.Errorf("无法连接服务器 %s：%w", a.client.base, err)
	}
	fmt.Printf("已连接到服务器 %s\n", a.client.base)
	if a.authMenu() {
		a.mainMenu()
		_ = a.client.Logout()
	}
	fmt.Println("已退出，再见！")
	return nil
}

// ---------- 登录 / 注册 ----------

func (a *App) authMenu() bool {
	for {
		fmt.Println()
		fmt.Println("========== 交易小游戏 ==========")
		fmt.Println("1. 登录")
		fmt.Println("2. 注册")
		fmt.Println("0. 退出")
		switch readChoice("请输入序号: ", 0, 2) {
		case 1:
			if a.doLogin() {
				return true
			}
		case 2:
			if a.doRegister() {
				return true
			}
		case 0:
			return false
		}
	}
}

func (a *App) doLogin() bool {
	fmt.Println("\n--- 登录 ---")
	username := readLine("用户名: ")
	password := readPassword("密码: ")
	if username == "" || password == "" {
		fmt.Println("用户名和密码不能为空")
		return false
	}
	resp, err := a.client.Login(username, password)
	if err != nil {
		fmt.Println("登录失败:", err)
		return false
	}
	a.user = &resp.User
	fmt.Printf("登录成功！欢迎 %s，当前第 %d 天，余额 %.2f\n", a.user.Username, a.user.Day, a.user.Money)
	return true
}

func (a *App) doRegister() bool {
	fmt.Println("\n--- 注册（需要注册密钥）---")
	username := readLine("用户名: ")
	password := readPassword("密码: ")
	key := readLine("注册密钥: ")
	if username == "" || password == "" || key == "" {
		fmt.Println("用户名、密码、密钥均不能为空")
		return false
	}
	resp, err := a.client.Register(username, password, key)
	if err != nil {
		fmt.Println("注册失败:", err)
		return false
	}
	a.user = &resp.User
	fmt.Printf("注册成功并已自动登录！欢迎 %s，当前第 %d 天，余额 %.2f\n", a.user.Username, a.user.Day, a.user.Money)
	return true
}

// ---------- 主菜单 ----------

func (a *App) mainMenu() {
	for {
		a.refreshUser()
		fmt.Println()
		fmt.Printf("========== 主页面 | 第 %d 天 | 余额 %.2f ==========\n", a.user.Day, a.user.Money)
		fmt.Println("1. 交易")
		fmt.Println("2. 明天")
		fmt.Println("3. 特殊商品")
		fmt.Println("4. 退出")
		switch readChoice("请输入序号: ", 1, 4) {
		case 1:
			a.tradeMenu()
		case 2:
			a.doTomorrow()
		case 3:
			a.specialMenu()
		case 4:
			return
		}
	}
}

func (a *App) refreshUser() {
	var u UserInfo
	if err := a.client.Me(&u); err == nil {
		a.user = &u
	}
}

// ---------- 交易 ----------

func (a *App) tradeMenu() {
	for {
		fmt.Println()
		fmt.Println("--- 交易 ---")
		fmt.Println("1. 列出商品")
		fmt.Println("2. 购买")
		fmt.Println("3. 卖出")
		fmt.Println("4. 查看仓库")
		fmt.Println("5. 返回上一页")
		switch readChoice("请输入序号: ", 1, 5) {
		case 1:
			a.listProducts()
		case 2:
			a.doBuy()
		case 3:
			a.doSell()
		case 4:
			a.listWarehouse()
		case 5:
			return
		}
	}
}

// listProducts 列出今日商品与价格
func (a *App) listProducts() *productsResp {
	res, err := a.client.Products()
	if err != nil {
		fmt.Println("获取商品失败:", err)
		return nil
	}
	fmt.Printf("\n第 %d 天 商城商品：\n", res.Day)
	t := NewTable("ID", "商品", "种类", "今日价格", "占用空间", "存储仓库", "普通过期", "冷藏过期")
	for _, p := range res.Products {
		t.AddRow(
			fmt.Sprint(p.ID),
			p.Name,
			p.Category,
			fmt.Sprintf("%.2f", p.Price),
			fmt.Sprint(p.Size),
			storageTypeName(p.StorageType),
			expireDaysText(p.NormalExpireDays),
			expireDaysText(p.ColdExpireDays),
		)
	}
	t.Print()
	return res
}

// doBuy 购买流程
func (a *App) doBuy() {
	res := a.listProducts()
	if res == nil {
		return
	}
	fmt.Println()
	productID := readIntRange("购买商品ID: ", 1, 0)
	var prod *ProductPrice
	for i := range res.Products {
		if int64(res.Products[i].ID) == productID {
			prod = &res.Products[i]
			break
		}
	}
	if prod == nil {
		fmt.Println("商品ID不存在")
		return
	}
	quantity := readIntRange("购买数量: ", 1, 0)

	storage := StorageNormal
	expire := prod.NormalExpireDays
	switch prod.StorageType {
	case StorageBoth:
		fmt.Println("该商品两种仓库均可存放：1. 普通仓库  2. 冷藏仓库")
		storage = readChoice("存储到: ", 1, 2)
		if storage == StorageCold {
			expire = prod.ColdExpireDays
		}
	case StorageCold:
		storage = StorageCold
		expire = prod.ColdExpireDays
	}

	cost := prod.Price * float64(quantity)
	fmt.Printf("预计支出 %.2f（单价 %.2f × %d），保存期限：%s\n", cost, prod.Price, quantity, expireDaysText(expire))
	if !confirm("确认购买? (y/n): ") {
		fmt.Println("已取消购买")
		return
	}
	tr, err := a.client.Buy(prod.ID, int(quantity), storage)
	if err != nil {
		fmt.Println("购买失败:", err)
		return
	}
	a.user.Money = tr.Balance
	a.user.Day = tr.Day
	fmt.Printf("购买成功：%s x%d 单价 %.2f，支出 %.2f，余额 %.2f（存于%s，%s）\n",
		tr.ProductName, tr.Quantity, tr.UnitPrice, tr.Amount, tr.Balance,
		storageTypeName(tr.StorageType), expireDayText(tr.ExpireDay))
}

// listWarehouse 查看仓库
func (a *App) listWarehouse() *warehouseResp {
	res, err := a.client.Warehouse()
	if err != nil {
		fmt.Println("获取仓库失败:", err)
		return nil
	}
	fmt.Printf("\n第 %d 天 仓库库存：\n", res.Day)
	if len(res.Items) == 0 {
		fmt.Println("（仓库为空）")
		return res
	}
	t := NewTable("商品ID", "商品", "仓库", "数量", "均价", "总价值", "今日价格", "全部卖出盈亏", "到期")
	for _, it := range res.Items {
		t.AddRow(
			fmt.Sprint(it.ProductID),
			it.ProductName,
			storageTypeName(it.StorageType),
			fmt.Sprint(it.Quantity),
			fmt.Sprintf("%.2f", it.AvgPrice),
			fmt.Sprintf("%.2f", it.TotalCost),
			fmt.Sprintf("%.2f", it.TodayPrice),
			fmt.Sprintf("%+.2f", it.Profit),
			expireDayText(it.ExpireDay),
		)
	}
	t.Print()
	return res
}

// doSell 卖出流程
func (a *App) doSell() {
	res := a.listWarehouse()
	if res == nil || len(res.Items) == 0 {
		return
	}
	fmt.Println()
	productID := readIntRange("卖出商品ID: ", 1, 0)
	var matched []WarehouseItem
	storages := map[int]bool{}
	for _, it := range res.Items {
		if int64(it.ProductID) == productID {
			matched = append(matched, it)
			storages[it.StorageType] = true
		}
	}
	if len(matched) == 0 {
		fmt.Println("仓库中没有该商品")
		return
	}

	storage := 0
	if len(storages) > 1 {
		// 同一商品存于两种仓库时让用户选择
		storage = readStorageOptional()
	} else {
		for s := range storages {
			storage = s
		}
	}

	var total int64
	price := matched[0].TodayPrice
	for _, it := range matched {
		if storage == 0 || it.StorageType == storage {
			total += int64(it.Quantity)
		}
	}
	fmt.Printf("可卖出数量：%d，今日单价 %.2f\n", total, price)
	quantity := readIntRange("卖出数量: ", 1, 0)
	fmt.Printf("预计收入 %.2f（单价 %.2f × %d）\n", price*float64(quantity), price, quantity)
	if !confirm("确认卖出? (y/n): ") {
		fmt.Println("已取消卖出")
		return
	}
	tr, err := a.client.Sell(uint(productID), int(quantity), storage)
	if err != nil {
		fmt.Println("卖出失败:", err)
		return
	}
	a.user.Money = tr.Balance
	fmt.Printf("卖出成功：%s x%d 单价 %.2f，收入 %.2f，余额 %.2f\n",
		tr.ProductName, tr.Quantity, tr.UnitPrice, tr.Amount, tr.Balance)
}

// ---------- 明天 ----------

func (a *App) doTomorrow() {
	res, err := a.client.Tomorrow()
	if err != nil {
		fmt.Println("推进到明天失败:", err)
		return
	}
	a.user.Day = res.Day
	fmt.Printf("\n已更新到明天，当前第 %d 天\n", res.Day)
	for _, c := range res.TriggeredCrits {
		fmt.Printf("  [暴击] %s（影响商品：%s）\n", c.Description, strings.Join(c.Products, "、"))
	}
	for _, e := range res.ExpiredItems {
		fmt.Printf("  [过期] %s x%d 已过期销毁（损失成本 %.2f）\n", e.ProductName, e.Quantity, e.Value)
	}
	if res.ExpiredPurchases > 0 {
		fmt.Printf("  [仓库] 有 %d 笔已购买的仓库空间到期失效\n", res.ExpiredPurchases)
	}
}

// ---------- 特殊商品（仓库空间） ----------

func (a *App) specialMenu() {
	for {
		fmt.Println()
		fmt.Println("--- 特殊商品（仓库空间购买）---")
		fmt.Println("1. 查看现有仓库空间大小")
		fmt.Println("2. 购买普通仓库")
		fmt.Println("3. 购买冷藏仓库")
		fmt.Println("4. 返回上一页")
		switch readChoice("请输入序号: ", 1, 4) {
		case 1:
			a.showSpace()
		case 2:
			a.buySpace(StorageNormal)
		case 3:
			a.buySpace(StorageCold)
		case 4:
			return
		}
	}
}

// showSpace 查看仓库空间大小
func (a *App) showSpace() *spaceResp {
	res, err := a.client.WarehouseSpace()
	if err != nil {
		fmt.Println("获取仓库空间失败:", err)
		return nil
	}
	fmt.Printf("\n第 %d 天 仓库空间：\n", res.Day)
	t := NewTable("仓库", "总容量", "已用", "剩余", "使用率")
	t.AddRow("普通仓库", fmt.Sprint(res.Normal.Capacity), fmt.Sprint(res.Normal.Used),
		fmt.Sprint(res.Normal.Free), usageText(res.Normal))
	t.AddRow("冷藏仓库", fmt.Sprint(res.Cold.Capacity), fmt.Sprint(res.Cold.Used),
		fmt.Sprint(res.Cold.Free), usageText(res.Cold))
	t.Print()
	fmt.Printf("空间单价：普通仓库 %.2f/单位/月，冷藏仓库 %.2f/单位/月（每月按30天，最短1个月最长12个月）\n",
		res.UnitPriceNormal, res.UnitPriceCold)

	if len(res.Purchases) > 0 {
		fmt.Println("已购买的空间记录：")
		pt := NewTable("仓库", "大小", "月数", "开始天", "到期天")
		for _, p := range res.Purchases {
			pt.AddRow(storageTypeName(p.StorageType), fmt.Sprint(p.Size),
				fmt.Sprint(p.Months), fmt.Sprint(p.StartDay), fmt.Sprint(p.ExpireDay))
		}
		pt.Print()
	}
	return res
}

// buySpace 购买仓库空间
func (a *App) buySpace(storage int) {
	name := storageTypeName(storage)
	res, err := a.client.WarehouseSpace()
	if err != nil {
		fmt.Println("获取仓库空间失败:", err)
		return
	}
	unitPrice := res.UnitPriceNormal
	if storage == StorageCold {
		unitPrice = res.UnitPriceCold
	}
	fmt.Printf("\n--- 购买%s（单价 %.2f/单位/月）---\n", name, unitPrice)
	size := readIntRange("购买空间大小: ", 1, 0)
	months := readIntRange("购买月数（1-12）: ", 1, 12)
	cost := unitPrice * float64(size) * float64(months)
	fmt.Printf("预计支出 %.2f（%d 空间 × %d 个月），到期天：第 %d 天\n",
		cost, size, months, res.Day+int(months)*30)
	if !confirm("确认购买? (y/n): ") {
		fmt.Println("已取消购买")
		return
	}
	if err := a.client.WarehouseBuy(storage, size, months); err != nil {
		fmt.Println("购买失败:", err)
		return
	}
	fmt.Printf("购买成功：%s %d 空间 × %d 个月，支出 %.2f\n", name, size, months, cost)
}

// ---------- 展示辅助 ----------

func storageTypeName(t int) string {
	switch t {
	case StorageNormal:
		return "普通仓库"
	case StorageCold:
		return "冷藏仓库"
	case StorageBoth:
		return "普通/冷藏均可"
	}
	return "未知"
}

func expireDaysText(days int) string {
	switch days {
	case -1:
		return "永久"
	case 0:
		return "不适用"
	}
	return fmt.Sprintf("%d天", days)
}

func expireDayText(day int) string {
	if day == -1 {
		return "永久"
	}
	return fmt.Sprintf("第%d天", day)
}

func usageText(s StorageSpace) string {
	if s.Capacity <= 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", float64(s.Used)/float64(s.Capacity)*100)
}
