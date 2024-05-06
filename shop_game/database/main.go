package database

import (
	"errors"
	"shop_game/shop_game/tools"
)

func IsSupported(databaseType string) (DataBase, error) {
	switch databaseType {
	case "mysql":
		// return mysql interface
		return &Mysql{}, nil
	default:
		return nil, errors.New("not found supported database type.")
	}
}

// DataBase All databases used must implement this interface
type DataBase interface {
	// Conn The config item of the configuration file is passed, which contains the content required for the link
	Conn(config tools.ConfigFile)
	// Login func
	Login(username string, password string) bool
	// Init Initialization requires creating tables and adding fields
	Init()
	// SameDayGoods Return to all the content of the mall on that day
	SameDayGoods() []tools.TmpDay
	// AllWarehouse Information on all goods in the warehouse
	AllWarehouse() []tools.Warehouse
	// PurchaseGoods To buy goods, you need to deduct the user's money, subtract the user's remaining space, and increase the record of the purchased goods
	PurchaseGoods(userInfo *tools.UserInformation, tr tools.TransactionRecord, td tools.TmpDay) bool
	// ClotheSold To sell goods, you need to increase the user's money, increase the user's remaining space, and delete the purchase record of the product
	ClotheSold(userInfo *tools.UserInformation, tr tools.TransactionRecord) bool
	// SearchByID 根据ID查找
	SearchByID(id int64) (bool, tools.TmpDay)
	// QueryUserInfo 查询用户信息
	QueryUserInfo(username string) tools.UserInformation
	// 输出用户信息
	// PrintUserInfor(userInfo *tools.UserInformation)

	/*
		new logic
		将操作和数据库拆分开 数据库只查询 写入数据
	*/
	// 返回 store 表的所有内容
	QueryTable_store() []tools.Store
	// WriteTable_tmp_day 写入 tmp_day 表
	WriteTable_tmp_day(mpi []tools.TmpDay)
	// FlushUserInfo 写入用户信息
	FlushUserInfo(userinfo *tools.UserInformation)
	// EntireTable 扫描订单表 返回全部数据
	EntireTable() []tools.TransactionRecord

	// Special 特殊商品内容
	ListTable() []tools.SpecialGoods
	// SearchByIDForSpecialGoodsTable 根据id查找special_goods表的数据
	SearchByIDForSpecialGoodsTable(id int8) (bool, tools.SpecialGoods)
}
