package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"shop_game/shop_game/tools"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
Mysql 数据库操作结构体
所有方法都要实现一个DB *sql.DB类型
*/
type Mysql struct {
	DB           *sql.DB
	username     string
	password     string
	address      string
	port         int
	dataBaseName string
	config       tools.ConfigFile
}

func (mysql *Mysql) Conn(config tools.ConfigFile) {
	// conn mysql database.
	var err error
	var connstr string

	// init struct
	mysql.username = config.Config.Database.Username
	mysql.password = config.Config.Database.Password
	mysql.port = config.Config.Database.Port
	mysql.address = config.Config.Database.Address
	mysql.dataBaseName = config.Config.Database.Database
	mysql.config = config

	connstr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true", mysql.username, mysql.password, mysql.address, mysql.port, mysql.dataBaseName)

	mysql.DB, err = sql.Open("mysql", connstr)
	if err != nil {
		fmt.Println("[tools.go:NewDB]:conn mysql error!")
		panic(err)
	} else {
		fmt.Println("Database link successful!")
	}
	(mysql.DB).SetConnMaxLifetime(time.Minute * 3)
	(mysql.DB).SetMaxOpenConns(20)
	(mysql.DB).SetMaxIdleConns(20)
	(mysql.DB).SetConnMaxLifetime(time.Second * 30)
}

func (mysql *Mysql) Init() {
	// read file content
	var yamlData tools.AllGoods = tools.ReadGoods(mysql.config.Config.Goods.Import_file)

	// create tables
	content, err := ioutil.ReadFile("shop_game/database/metadata.sql")
	if err != nil {
		log.Fatalln("[mysql.go:(Mysql)Init()]:Failed to read create table file.", err)
	}
	sqlStatements := string(content)

	_, err = mysql.DB.Exec(sqlStatements)
	if err != nil {
		log.Fatalln(err, "\n[mysql.go:(Mysql)Init()]:exec init sql error!\n",
			"\tPlease check the content filled in the config.database.database field in the configuration file.\n",
			"\tThe database filled in must have already been created.\n",
			"\tIf not, refer to the creation command: CREATE DATABASE IF NOT EXISTS `trading_game` CHARACTER SET utf8;")
	}

	// insert goods data
	for _, v := range yamlData.Goods {
		_, err = mysql.DB.Exec("insert into `store` (`name`,`start_price`,`end_price`,`max_number`,`mini_number`,`commodity_size`,`storage_date`,`refrigerate`,`max_price`,`mini_price`) values (?,?,?,?,?,?,?,?,?,?);",
			v.Name, v.Start_price, v.End_price, v.Max_number, v.Mini_number, v.Commodity_size, v.Storage_date, v.Refrigerate, v.Max_price, v.Mini_price)
		if err != nil {
			log.Fatalln("[mysql.go:(Mysql)Init()]: insert data error!", err)
		}
	}

	// insert opportunity
	for _, v := range yamlData.Oppor {
		_, _ = mysql.DB.Exec("insert into `opportunity` (`events_name`,`critical_chance`,`max_time`,`mini_time`) values (?,?,?,?)",
			v.EventsName, v.CriticalChance, v.MaxTime, v.MiniTime)
	}

	// insert user information
	_, _ = mysql.DB.Exec("insert into `user` (`username`,`password`,`money`,`warehouse`,`refrigerate_warehouse`,`day`) values (?,?,?,?,?, 0)",
		mysql.config.Config.Login_method.Other_information.Username, tools.EncryptPassword(mysql.config.Config.Login_method.Other_information.Password),
		mysql.config.Config.Initialize_configuration.Default_money,
		mysql.config.Config.Initialize_configuration.Warehouse.Default_size,
		mysql.config.Config.Initialize_configuration.Refrigerate_warehouse.Default_size)

	// insert special
	_, _ = mysql.DB.Exec("insert into `special_goods` (`special_name`,`special_price`,`special_content`) values (?,?,'{}'),(?,?,'{}')",
		"warehouse", mysql.config.Config.Initialize_configuration.Warehouse.Expansion_costs,
		"refrigerate", mysql.config.Config.Initialize_configuration.Refrigerate_warehouse.Expansion_costs)

	log.Println("init database end.")
}

func (mysql *Mysql) Login(username string, encryptpassword string) bool {
	var executeStatement string
	var password string
	executeStatement = fmt.Sprintf("select `password` from `user` where `username` = '%s';", username)
	rows, err := mysql.DB.Query(executeStatement)
	if err != nil {
		log.Fatalln("login failed! Please check if the database configuration is correct or if the database is started.")
	}
	defer func() {
		_ = rows.Close()
	}()
	rows.Next()
	_ = rows.Scan(&password)
	if password == encryptpassword {
		return true
	}

	return false
}

func (mysql *Mysql) SameDayGoods() []tools.TmpDay {
	// 返回当天商城的所有内容
	var mpi []tools.TmpDay
	rows, _ := mysql.DB.Query("select * from tmp_day;")
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatalln("[mysql.go:SameDayGoods()]:: rows.Close error!")
		}
	}()
	for rows.Next() {
		var tmpMpi tools.TmpDay
		_ = rows.Scan(&tmpMpi.ID, &tmpMpi.Name, &tmpMpi.Price, &tmpMpi.Size, &tmpMpi.PurchaseNumber,
			&tmpMpi.SellingNumber, &tmpMpi.Refrigerate, &tmpMpi.CriticalStrike)
		mpi = append(mpi, tmpMpi)
	}

	return mpi
}

func (mysql *Mysql) AllWarehouse() []tools.Warehouse {
	// 返回仓库的所有货物信息
	var uw []tools.Warehouse
	rows, _ := mysql.DB.Query("select * from warehouse;")
	defer rows.Close()
	for rows.Next() {
		var tmpUserWare tools.Warehouse
		_ = rows.Scan(&tmpUserWare.ID, &tmpUserWare.Name, &tmpUserWare.Day, &tmpUserWare.AvgPrice, &tmpUserWare.Number, &tmpUserWare.StorageDate)
		uw = append(uw, tmpUserWare)
	}
	return uw

}

func (mysql *Mysql) PurchaseGoods(userInfo *tools.UserInformation, tr tools.TransactionRecord, td tools.TmpDay) bool {
	// 买入商品 订单表增加记录 仓库增加或者更新记录 用户表保存

	// 保存用户信息(用户信息已经处理过可以直接保存)
	var err error
	begin, err := mysql.DB.Begin()
	//_, err = begin.Exec("update `user` set `money` = ?,warehouse = ? ,refrigerate_warehouse = ? where `id` = ?",
	//	userInfo.Money, userInfo.Warehouse, userInfo.RefrigerateWarehouse, userInfo.ID)
	mysql.FlushUserInfo(userInfo)

	// 在每日限制减去已购买的数量
	_, err = begin.Exec("update `tmp_day` set `purchase_number` = ? where id = ?;",
		td.PurchaseNumber-tr.Number, td.ID)

	// 增加购买记录
	_, err = begin.Exec("insert into `transaction_record` (`name`,`shop_price`,`number`,`type`,`shop_day`,`op_type`) values (?,?,?,?,?,?)", tr.Name, tr.ShopPrice, tr.Number, tr.Type, tr.ShopDay, tr.OpType)

	// 增加仓库记录 需要先查找检查是否存在
	row, err := begin.Query("select * from `warehouse` where `name` = ?;", tr.Name)

	if row.Next() {
		// 如果进入判断 表示存在相同记录 那么需要先查出来
		var wh tools.Warehouse
		err = row.Scan(&wh.ID, &wh.Name, &wh.Day, &wh.AvgPrice, &wh.Number, &wh.StorageDate)
		// 读取完成 关闭链接
		if err := row.Close(); err != nil {
			log.Println("[mysql.go:PurchaseGoods()]: row.Close() error!")
			panic(err)
		}
		totalNumber := wh.Number + tr.Number
		avgPrice := ((wh.AvgPrice * wh.Number) + (tr.Number * tr.ShopPrice)) / totalNumber

		_, err = begin.Exec("update `warehouse` set `day` = ?,`avg_price` = ?, `number` = ? where `name` = ?;",
			userInfo.Day, avgPrice, totalNumber, wh.Name)
		if err != nil {
			log.Println("[mysql.go]:", err)
			panic("")
		}
	} else {
		// 没有记录 直接插入
		_, err = begin.Exec("insert into `warehouse` (`name`,`day`,`avg_price`,`number`,`storage_date`) values (?,?,?,?,?)",
			tr.Name, userInfo.Day, tr.ShopPrice, tr.Number, 0)
	}

	if err != nil {
		// 不为空表示前面执行语句存在错误 回滚 返回
		if err := begin.Rollback(); err != nil {
			log.Println("[mysql.go:PurchaseGoods()]: Rollback error!")
			panic(err)
		}
		return false
	}

	if err := begin.Commit(); err != nil {
		log.Println("[mysql.go:PurchaseGoods()]: Commit error!")
		panic(err)
	}

	return true
}

func (mysql *Mysql) ClotheSold(userInfo *tools.UserInformation, tr tools.TransactionRecord) bool {
	// 卖出商品 需要增加用户的金钱 增加用户剩余空间 删除商品的买入记录
	// tr TransactionRecord 用户订单数据
	// td tmp_day 产品当日对应的信息
	// wh 仓库表对应的信息
	// userInfo 用户对应的信息
	var err error
	// 开启一个事务
	begin, err := mysql.DB.Begin()
	// 查询仓库对应的订单表的内容
	row, err := begin.Query("select * from warehouse where `id` = ?", tr.ID)
	if !row.Next() {
		log.Println(tools.Red + "Products without corresponding ID!" + tools.End)
		return false
	}
	var wh tools.Warehouse
	err = row.Scan(&wh.ID, &wh.Name, &wh.Day, &wh.AvgPrice, &wh.Number, &wh.StorageDate)
	err = row.Close()

	// 查询当日该商品的价格
	row, err = begin.Query("select * from tmp_day where `name` = ?;", wh.Name)
	if !row.Next() {
		log.Println(tools.Red + "This product has not been listed..." + tools.End)
		return false
	}
	var td tools.TmpDay
	err = row.Scan(&td.ID, &td.Name, &td.Price, &td.Size, &td.PurchaseNumber, &td.SellingNumber, &td.Refrigerate, &td.CriticalStrike)
	err = row.Close()

	// 检查总数 如果 仓库数量 不大于等于 卖出数量 返回
	if !(wh.Number >= tr.Number) {
		log.Println(tools.Red + "The quantity sold is greater than the quantity in stock!" + tools.End)
		return false
	}
	// 计算是否超出当日可允许的卖出的最大限制
	if tr.Number > td.SellingNumber {
		log.Println("Exceeding sales quantity limit！")
		return false
	}
	// 修改最大限制 减去已购买的数量
	_, err = begin.Exec("update `tmp_day` set `selling_number` = ? where `id` = ?",
		td.SellingNumber-tr.Number, td.ID)

	// 插入订单表
	_, err = begin.Exec("insert into `transaction_record` (`name`,`shop_price`,`number`,`type`,`shop_day`,`op_type`)values (?,?,?,?,?,?);",
		wh.Name, td.Price, tr.Number, td.Refrigerate, userInfo.Day, 1)
	// 修改库存
	if wh.Number == tr.Number {
		// remove
		_, err = begin.Exec("delete from `warehouse` where `id` = ?", tr.ID)
	} else {
		_, err = begin.Exec("update `warehouse` set `number` = ? where `id` = ?;", wh.Number-tr.Number, tr.ID)
	}
	// 计算出售价格 增加用户余额
	Price := userInfo.Money + (td.Price * tr.Number)
	_, err = begin.Exec("update `user` set `money` = ? where username = ?", Price, userInfo.Username)
	// 减去占用空间 判断商品是否是冷藏商品

	if td.Refrigerate == 0 {
		// 冷藏
		totalSize := userInfo.RefrigerateWarehouse + (td.Size * tr.Number)
		_, err = begin.Exec("update `user` set `refrigerate_warehouse` = ? where username = ?", totalSize, userInfo.Username)
	} else {
		// 普通
		totalSize := userInfo.Warehouse + (td.Size * tr.Number)
		_, err = begin.Exec("update `user` set `warehouse` = ? where username = ?", totalSize, userInfo.Username)
	}

	err = begin.Commit()
	if err != nil {
		fmt.Println(tools.Red + "[mysql.go:ClotheSold] save data error!" + tools.End)
		return false
	}

	return true
}

func (mysql *Mysql) SearchByID(id int64) (bool, tools.TmpDay) {
	// 根据ID查找
	var mpi tools.TmpDay
	row, err := mysql.DB.Query("select * from `tmp_day` where id = ?;", id)
	if err != nil {
		log.Println(tools.Red + "[mysql.go:SearchByID()]:: error!" + tools.End)
		return false, mpi
	}
	defer func() {
		_ = row.Close()
	}()
	if !row.Next() {
		return false, mpi
	}
	_ = row.Scan(&mpi.ID, &mpi.Name, &mpi.Price, &mpi.Size, &mpi.PurchaseNumber, &mpi.SellingNumber, &mpi.Refrigerate, &mpi.CriticalStrike)

	return true, mpi
}

func (mysql *Mysql) QueryUserInfo(username string) tools.UserInformation {
	// 查询用户 返回对应的用户结构体
	var ui tools.UserInformation

	row, err := mysql.DB.Query("select * from `user` where `username` = ?;", username)
	if err != nil {
		log.Fatalln(tools.Red + "[mysql.go:QueryUserInfo()]:: Format user information failed!" + tools.End)
	}
	if row.Next() {
		_ = row.Scan(&ui.ID, &ui.Username, &ui.Password, &ui.Money, &ui.Warehouse, &ui.RefrigerateWarehouse, &ui.Day)
	} else {
		log.Fatalln(tools.Red + "No user information found!" + tools.End)
	}

	return ui
}

// func (mysql *Mysql) PrintUserInfor(userInfo *tools.UserInformation) {
// 	fmt.Printf("----------------------- user information start -----------------------\n")
// 	fmt.Printf("Username\tDay\tWarehouse\tREWarehouse\tMoney\n")
// 	fmt.Printf("%s\t\t%d\t%d\t\t%d\t\t%d\n", userInfo.Username, userInfo.Day, userInfo.Warehouse, userInfo.RefrigerateWarehouse, userInfo.Money)
// 	fmt.Printf("------------------------ user information end ------------------------\n")
// }

func (mysql *Mysql) QueryTable_store() []tools.Store {
	var goods []tools.Store
	rows, _ := mysql.DB.Query("select * from `store`;")
	for rows.Next() {
		var good tools.Store
		_ = rows.Scan(&good.ID, &good.Name, &good.Start_price, &good.End_price, &good.Max_number, &good.Mini_number, &good.Commodity_size, &good.Storage_date, &good.Refrigerate, &good.Max_price, &good.Mini_price)
		goods = append(goods, good)
	}
	return goods
}

func (mysql *Mysql) WriteTable_tmp_day(mpi []tools.TmpDay) {
	/*
		根据传入的数据 循环写入到tmp_day表中
	*/
	_, _ = mysql.DB.Exec("truncate `tmp_day`;")
	for _, v := range mpi {
		_, _ = mysql.DB.Exec("insert into `tmp_day` (`name`,`price`,`size`,`purchase_number`,`selling_number`,`refrigerate`,`critical_strike`)values(?,?,?,?,?,?,?);",
			v.Name, v.Price, v.Size, v.PurchaseNumber, v.SellingNumber, v.Refrigerate, v.CriticalStrike)
	}
}

func (mysql *Mysql) FlushUserInfo(userinfo *tools.UserInformation) {
	/*
		根据传入的userinfo的值 写入到数据库
	*/
	_, _ = mysql.DB.Exec("update `user` set `money` = ?,`warehouse` = ?,`refrigerate_warehouse` = ?,`day` = ? where `username` = ?;",
		userinfo.Money, userinfo.Warehouse, userinfo.RefrigerateWarehouse, userinfo.Day, userinfo.Username)
}

func (mysql *Mysql) EntireTable() []tools.TransactionRecord {
	/*
		返回 transaction_record 的所有内容
	*/
	var trs []tools.TransactionRecord
	rows, _ := mysql.DB.Query("select * from transaction_record;")
	for rows.Next() {
		var tr tools.TransactionRecord
		_ = rows.Scan(&tr.ID, &tr.Name, &tr.ShopPrice, &tr.Number, &tr.Type, &tr.ShopDay, &tr.OpType)
		trs = append(trs, tr)
	}
	return trs
}

/*
特殊商品对应函数
*/
func (m *Mysql) ListTable() []tools.SpecialGoods {
	var specialgoods []tools.SpecialGoods
	rows, _ := m.DB.Query("select * from `special_goods`;")
	for rows.Next() {
		var sg tools.SpecialGoods
		rows.Scan(&sg.ID, &sg.SpecialName, &sg.SpecialPrice, &sg.SpecialContent)
		specialgoods = append(specialgoods, sg)
	}
	return specialgoods
}

func (m *Mysql) SearchByIDForSpecialGoodsTable(id int8) (bool, tools.SpecialGoods) {
	var sg tools.SpecialGoods
	rows, _ := m.DB.Query("select * from `special_goods` where `id` = ?;", id)
	if rows.Next() {
		rows.Scan(&sg.ID, &sg.SpecialName, &sg.SpecialPrice, &sg.SpecialContent)
	} else {
		return false, sg
	}
	return true, sg
}
