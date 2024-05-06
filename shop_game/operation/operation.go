package operation

import (
	"fmt"
	"log"
	"math/rand"
	"shop_game/shop_game/database"
	"shop_game/shop_game/tools"
	"time"
)

/*
所有操作相关内容
*/
func NewStart(configfilename string) {
	var config tools.ConfigFile
	config = tools.ReadConfig(configfilename)
	DBO, err := database.IsSupported(config.Config.Database.Type)
	if err != nil {
		panic("init database error!")
	}
	// login auth.
	DBO.Conn(config)
	if config.Config.Login_method.Model == "builtin" {
		login(DBO)
	}
	// 字体测试
	fmt.Println("字体测试: "+tools.Red+"Red"+tools.End,
		tools.Yellow+"Yellow"+tools.End,
		tools.Green+"Green"+tools.End)
	homePage(config, DBO)
}

func login(base database.DataBase) bool {
	var username string
	var password string

	fmt.Printf("\n\nplease input username and password.\n[username]:")
	_, _ = fmt.Scanln(&username)
	fmt.Printf("[password]:")
	_, _ = fmt.Scanln(&password)

	// encrypt user password.
	if !base.Login(username, tools.EncryptPassword(password)) {
		log.Fatalln(tools.Red + "\nLogin failed, username or password incorrect." + tools.End)
	}

	// 格式化全局变量 用户结构体
	userInfo = base.QueryUserInfo(username)

	return true
}

// 前面的认证全部通过结束以后 进入主页
func homePage(config tools.ConfigFile, DBO database.DataBase) {
	fmt.Println("***************************")
	fmt.Println("*                         *")
	fmt.Println("*   Welcome to the game   *")
	fmt.Println("*                         *")
	fmt.Println("***************************")

	for {
		fmt.Println("1. 普通商店")
		fmt.Println("2. 特殊商店")
		fmt.Println("q. 退出")

		fmt.Printf("please enter the operation: ")
		_, _ = fmt.Scanln(&tools.Oper)

		switch tools.Oper {
		case "1":
			regularStores(DBO)
		case "2":
			specialStores(DBO)
		case "q":
			log.Print("exit.")
			return
		default:
			tools.ReEnter()
		}
	}
}

// 普通商店
func regularStores(DBO database.DataBase) {
	for {
		fmt.Println("\n\nWelcome to the regular store!")
		fmt.Println("1. 集市")
		fmt.Println("2. 检查仓库")
		fmt.Println("q. 退出")

		fmt.Printf("please input operation: ")
		_, _ = fmt.Scanln(&tools.Oper)
		switch tools.Oper {
		case "1":
			// 初始化集市和调用启动方法
			var m Market
			m.DBO = DBO
			m.marketMain()
		case "2":
			var w Warehouse
			w.DBO = DBO
			w.WarehouseMain()
		case "q":
			return
		default:
			tools.ReEnter()
		}
	}
}

// Market 集市
type Market struct {
	DBO database.DataBase
}

func (m *Market) marketMain() {
	for {
		fmt.Println("\n\nWelcome to the market！")

		fmt.Println("1. 列出当前仓库的所有库存")
		fmt.Println("2. 列出集市的所有商品信息")
		fmt.Println("3. 买入商品")
		fmt.Println("4. 卖出商品")
		fmt.Println("5. 明天")
		fmt.Println("6. 输出用户信息")
		fmt.Println("q. 退出")
		fmt.Printf("please input operation: ")
		_, _ = fmt.Scanln(&tools.Oper)

		switch tools.Oper {
		case "1":
			m.ListInventoryOfWarehouse()
		case "2":
			m.ListAllProductInformationForTheMarket()
		case "3":
			m.PurchaseOfGoods()
		case "4":
			m.SellingOfGoods()
		case "5":
			m.EnteringTomorrow()
		case "6":
			m.PrintUserInfor()
		case "q":
			return
		default:
			tools.ReEnter()
		}
	}
}

/*
下面为需要操作部分
*/

func (m *Market) ListInventoryOfWarehouse() {
	// 1. 列出当前仓库的所有库存
	uw := m.DBO.AllWarehouse()
	if len(uw) == 0 {
		log.Print(tools.Yellow + "You haven't bought anything yet ..." + tools.End)
		return
	}

	fmt.Println("ID\tName\tDay\tAvgPrice\tNumber")
	for _, v := range uw {
		fmt.Printf("%d\t%s\t%d\t%d\t\t%d\n", v.ID, v.Name, v.Day, v.AvgPrice, v.Number)
	}
}

func (m *Market) ListAllProductInformationForTheMarket() {
	// 2. 列出集市的所有商品信息
	var refrigerate string
	var criticalStrike string

	mpi := m.DBO.SameDayGoods()

	// 集市没有商品信息...
	if len(mpi) == 0 {
		fmt.Println(tools.Yellow + "not found ..." + tools.End)
	}

	fmt.Println("ID\tName\t\tPrice\tSize\tPurchaseNumber\tSellingNumber\tRefrigerate\tCriticalStrike")
	for _, v := range mpi {
		// 判断是否是普通商品
		if v.Refrigerate == 1 {
			refrigerate = "普通"
		} else {
			refrigerate = "冷藏"
		}
		// 判断是否发生了暴击
		if v.CriticalStrike == 0 {
			criticalStrike = "无效果"
		} else if v.CriticalStrike == 1 {
			criticalStrike = "暴涨"
		} else if v.CriticalStrike == 2 {
			criticalStrike = "暴跌"
		}
		fmt.Printf("%d\t%s\t\t%d\t%d\t%d\t\t%d\t\t%s\t\t%s\n", v.ID, v.Name, v.Price, v.Size, v.PurchaseNumber,
			v.SellingNumber, refrigerate, criticalStrike)
	}
}

func (m *Market) PurchaseOfGoods() {
	// 3. 买入商品
	var op int64
	var err error
	var mpi tools.TmpDay
	var key bool
	var status bool
	var tr tools.TransactionRecord

	// 输入购买商品的ID
	for {
		fmt.Printf("input purch of goods number: ")
		_, err = fmt.Scanln(&op)
		if err != nil {
			tools.ReEnter()
			// log.Println(tools.Yellow + "" + tools.End)
			continue
		}

		// 如果输入正确 那么检查是否有此ID的商品
		key, mpi = m.DBO.SearchByID(op)
		if !key {
			// 如果没有
			log.Println(tools.Yellow + "This product was not found." + tools.End)
			continue
		} else {
			break
		}
	}

	// 输入购买商品的数量
	for {
		fmt.Printf("Please enter the purchase quantity:")
		_, err = fmt.Scanln(&op)
		if err != nil {
			tools.ReEnter()
			continue
		}

		// 创建订单结构体
		tr.Name = mpi.Name
		tr.Number = op
		tr.ShopPrice = mpi.Price
		tr.Type = mpi.Refrigerate
		tr.ShopDay = userInfo.Day
		tr.OpType = 0

		// 计算购买数量是否超出商城所存在的上限
		if tr.Number > mpi.PurchaseNumber {
			log.Println(tools.Red + "Exceeded purchase limit!" + tools.End)
		}
		// 如果输入正确 那么开始检余额和剩余空间是否足够
		// 计算需要花费的总额
		var totalAmount int64
		totalAmount = mpi.Price * op
		// 计算总空间
		var totalSpace int64
		totalSpace = int64(mpi.Size) * op
		if mpi.Refrigerate == 1 {
			// 普通商品
			if totalAmount <= userInfo.Money && totalSpace <= userInfo.Warehouse {
				// 购买成功 减去对应内容
				userInfo.Money = userInfo.Money - totalAmount
				userInfo.Warehouse = userInfo.Warehouse - totalSpace
				// 保存
				status = m.DBO.PurchaseGoods(&userInfo, tr, mpi)
			} else {
				fmt.Println(tools.Red + "Purchase failed due to insufficient balance or space!" + tools.End)
			}
		} else {
			// 冷冻商品
			if totalAmount <= userInfo.Money && totalSpace <= userInfo.RefrigerateWarehouse {
				//购买成功
				userInfo.Money = userInfo.Money - totalAmount
				userInfo.RefrigerateWarehouse = userInfo.RefrigerateWarehouse - totalSpace
				// 保存
				status = m.DBO.PurchaseGoods(&userInfo, tr, mpi)
			} else {
				fmt.Println(tools.Red + "Purchase failed due to insufficient balance or space!" + tools.End)
			}
		}

		if status {
			log.Println(tools.Green + "Purchase successful!" + tools.End)
		} else {
			log.Println(tools.Red + "Purchase failed!" + tools.End)
		}
		/*
			如果能执行到此位置 表示前面全部成功 那么直接跳出循环
			循环存在的意义是为了捕获输入时候遇到错误可以重新开始
		*/
		// 更新用户结构体
		userInfo = m.DBO.QueryUserInfo(userInfo.Username)
		break
	}
	// 交易结束 输出当前用户信息
	// m.DBO.PrintUserInfor(&userInfo)
	m.PrintUserInfor()

}
func (m *Market) SellingOfGoods() {
	/*
		4. 卖出商品
		先获取用户输入卖出的数量和ID 然后检查是否超出库存
		接下来开启事务 卖出对于数量的内容 增加余额 增加对应的仓库容量
	*/
	// 创建订单结构体
	var tr tools.TransactionRecord

	fmt.Printf("Please enter the ID of the product you want to sell: ")
	_, _ = fmt.Scanln(&tr.ID)
	fmt.Printf("Please enter the quantity to be sold: ")
	_, _ = fmt.Scanln(&tr.Number)

	// 检查是否有对于的ID和对应的数量
	if m.DBO.ClotheSold(&userInfo, tr) {
		fmt.Println(tools.Green + "selling success!" + tools.End)
		userInfo = m.DBO.QueryUserInfo(userInfo.Username)
	} else {
		fmt.Println(tools.Red + "selling failed!" + tools.End)
	}
}
func (m *Market) EnteringTomorrow() {
	// 5. 明天
	/*
		1. 获取所有商品 然后产生随机数 清空tmp_day表 写入到tmp_day表
		2. 更新用户天数
	*/
	var goods []tools.Store
	var mpi []tools.TmpDay
	goods = m.DBO.QueryTable_store()

	// 生成随机数字
	//fmt.Println(goods, mpi)
	rand.Seed(time.Now().UnixNano())
	for _, v := range goods {
		var tmpmpi tools.TmpDay
		// 常规赋值
		tmpmpi.Name = v.Name
		tmpmpi.Size = v.Commodity_size
		v.Refrigerate = int64(tmpmpi.Refrigerate)
		// 生成随机价格
		tmpmpi.Price = rand.Int63n(v.End_price-v.Start_price+1) + v.Start_price
		// 生成随机购买限制
		tmpmpi.PurchaseNumber = rand.Int63n(v.Max_number-v.Mini_number) + v.Mini_number
		// 生成随机卖出限制
		tmpmpi.SellingNumber = rand.Int63n(v.Max_number-v.Mini_number) + v.Mini_number
		// 是否触发暴击
		num := rand.Int63n(1000)
		if num >= 950 {
			// 大于则触发暴击
			tmpmpi.CriticalStrike = 1
			// 重新生成价格
			tmpmpi.Price = rand.Int63n(v.Max_price-v.End_price) + v.End_price
		} else if num <= 50 {
			// 暴跌
			tmpmpi.CriticalStrike = 2
			tmpmpi.Price = rand.Int63n(v.End_price-v.Mini_price) + v.Mini_price
		}
		mpi = append(mpi, tmpmpi)
	}
	m.DBO.WriteTable_tmp_day(mpi)
	// 刷新用户
	userInfo.Day = userInfo.Day + 1
	m.DBO.FlushUserInfo(&userInfo)
}

func (m *Market) PrintUserInfor() {
	// m.DBO.PrintUserInfor(&userInfo)
	userInfo = m.DBO.QueryUserInfo(userInfo.Username)
	fmt.Printf("----------------------- user information start -----------------------\n")
	fmt.Printf("Username\tDay\tWarehouse\tREWarehouse\tMoney\n")
	fmt.Printf("%s\t\t%d\t%d\t\t%d\t\t%d\n", userInfo.Username, userInfo.Day, userInfo.Warehouse, userInfo.RefrigerateWarehouse, userInfo.Money)
	fmt.Printf("------------------------ user information end ------------------------\n")
}

/*
2. 检查仓库
*/
type Warehouse struct {
	DBO database.DataBase
}

func (w *Warehouse) WarehouseMain() {
	for {
		fmt.Println("1. 查看所有交易记录")
		fmt.Println("2. 查看用户信息")
		fmt.Println("q. 退出")

		var op string
		fmt.Printf("[input operation]>")
		_, _ = fmt.Scanln(&op)

		switch op {
		case "1":
			w.ViewAllTransactionRecords()
		case "2":
			w.ViewUserInformation()
		case "q":
			return
		default:
			tools.ReEnter()
			// fmt.Println(tools.Red + "not found ..." + tools.End)
		}
	}
}

func (w *Warehouse) ViewAllTransactionRecords() {
	// 查看所有交易记录
	var storeClass string
	var operationType string
	var trs []tools.TransactionRecord
	trs = w.DBO.EntireTable()
	fmt.Println("-------------------------------- 所有记录 --------------------------------")
	fmt.Println("ID\tName\tPrice\tType\tDay\tType")
	for _, v := range trs {
		// 判断类型
		if v.Type == 0 {
			storeClass = "冷冻"
		} else {
			storeClass = "普通"
		}
		// 判断操作类型
		if v.OpType == 0 {
			operationType = "买入"
		} else {
			operationType = "卖出"
		}
		fmt.Printf("%d\t%s\t%d\t%s\t%d\t%s\n", v.ID, v.Name, v.ShopPrice, storeClass, v.ShopDay, operationType)
	}
	fmt.Println("-------------------------------- 所有记录 --------------------------------")
}
func (w *Warehouse) ViewUserInformation() {
	// m.DBO.PrintUserInfor(&userInfo)
	userInfo = w.DBO.QueryUserInfo(userInfo.Username)
	fmt.Printf("----------------------- user information start -----------------------\n")
	fmt.Printf("Username\tDay\tWarehouse\tREWarehouse\tMoney\n")
	fmt.Printf("%s\t\t%d\t%d\t\t%d\t\t%d\n", userInfo.Username, userInfo.Day, userInfo.Warehouse, userInfo.RefrigerateWarehouse, userInfo.Money)
	fmt.Printf("------------------------ user information end ------------------------\n")
}

/*
特殊商店开始代码
*/
func specialStores(DBO database.DataBase) {
	for {
		fmt.Println()
		var op string
		var m Market
		m.DBO = DBO
		fmt.Println("1. 查看商城货物")
		fmt.Println("2. 购买")
		fmt.Println("q. 退出")

		fmt.Printf("[please input]>")
		fmt.Scanln(&op)
		switch op {
		case "1":
			m.SpecialListGoods()
		case "2":
			m.SpecialShopGoods()
		case "q":
			return
		default:
			tools.ReEnter()
		}
	}
}

func (m *Market) SpecialListGoods() {
	// 1. 查看商城货物
	var specialgoods []tools.SpecialGoods
	specialgoods = m.DBO.ListTable()
	fmt.Println("ID\tName\tPrice\tDesc")
	for _, v := range specialgoods {
		fmt.Printf("%d\t%s\t%d\t%s\n", v.ID, v.SpecialName, v.SpecialPrice, "None")
	}
}
func (m *Market) SpecialShopGoods() {
	// 2. 购买

	var tr tools.TransactionRecord
	fmt.Printf("[please input ID]>")
	fmt.Scan(&tr.ID)
	fmt.Printf("[please input number]>")
	fmt.Scan(&tr.Number)
	// 根据ID获取对应的内容
	kbool, sg := m.DBO.SearchByIDForSpecialGoodsTable(tr.ID)
	if !kbool {
		log.Println(tools.Yellow + "No product with corresponding ID found!" + tools.End)
		return
	}
	// 计算
	if sg.SpecialPrice*tr.Number > userInfo.Money {
		log.Println(tools.Yellow + "Insufficient balance!" + tools.End)
	}

	// 增加用户信息
	if tr.ID == 1 {
		userInfo.Warehouse = tr.Number + userInfo.Warehouse
	} else if tr.ID == 2 {
		userInfo.RefrigerateWarehouse = tr.Number + userInfo.RefrigerateWarehouse
	}
	m.DBO.FlushUserInfo(&userInfo)
	log.Println(tools.Green + "success!" + tools.End)
}
