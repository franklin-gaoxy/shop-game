package tools

import (
	"database/sql"
)

var DB *sql.DB

// ConfigFile tools config file content
type ConfigFile struct {
	Config struct {
		Database struct { // database information
			Type     string `yaml:"type"`
			Database string `yaml:"database"`
			Address  string `yaml:"address"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"database"`
		Goods struct {
			Import_file string `yaml:"import_file"` // product information file.
		} `yaml:"goods"`
		Difficulty_level struct {
			Probability_of_price_increase int `yaml:"probability_of_price_increase"` // game mode.simple or diffcult
		} `yaml:"difficulty_level"`
		Login_method             loginConfig `yaml:"login_method"`
		Initialize_configuration struct {
			Default_money         int64           `yaml:"default_money"` // default amount of money
			Warehouse             warehouseConfig `yaml:"warehouse"`
			Refrigerate_warehouse warehouseConfig `yaml:"refrigerate_warehouse"`
		} `yaml:"initialize_configuration"`
	} `yaml:"config"`
}

type warehouseConfig struct {
	// 此为 默认仓库大小 和 扩容1点仓库所需金额 不需要太大
	Default_size    int16 `yaml:"default_size"`    // default warehouse size
	Expansion_costs int16 `yaml:"expansion_costs"` // the amount required for expanding size 1
}

type loginConfig struct {
	Model             string `yaml:"model"` // login model
	Other_information struct {
		Username string `yaml:"username"` // login username
		Password string `yaml:"password"` // login password
	} `yaml:"other_information"`
}

// goods file content
type AllGoods struct {
	Goods []Store       `yaml:"goods"`
	Oppor []Opportunity `yaml:"opportunity"`
}

type Opportunity struct {
	EventsName     string `yaml:"events_name"`
	CriticalChance int32  `yaml:"critical_chance"`
	MaxTime        int32  `yaml:"max_time"`
	MiniTime       int32  `yaml:"mini_time"`
}

// Store 数据库: store表内容 同时还是 Goods 配置文件
type Store struct {
	ID             int64
	Name           string `yaml:"name"`
	Start_price    int64  `yaml:"start_price"`
	End_price      int64  `yaml:"end_price"`
	Max_number     int64  `yaml:"max_number"`
	Mini_number    int64  `yaml:"mini_number"`
	Commodity_size int64  `yaml:"commodity_size"`
	Storage_date   int64  `yaml:"storage_date"`
	Refrigerate    int64  `yaml:"refrigerate"`
	Max_price      int64  `yaml:"max_price"`
	Mini_price     int64  `yaml:"mini_price"`
}

// TmpDay 数据库： tmp_day表内容 集市商品信息对应结构体
type TmpDay struct {
	ID             int8   // 商品ID
	Name           string // 商品名称
	Price          int64  // 该商品当天的价格
	Size           int64  // 该商品需要占用仓库的大小
	PurchaseNumber int64  // 该商品当日最大可购买的数量
	SellingNumber  int64  // 该商品当日最大可卖出的数量
	Refrigerate    int8   // 该商品是否为冷冻商品 0冷藏 1普通
	CriticalStrike int8   // 是否触发暴击效果 0无暴击 1暴涨 2暴跌
}

// 数据库: user表 用户表信息对应结构体
type UserInformation struct {
	ID                   int8   // 用户ID
	Username             string // 用户名
	Password             string // 密码
	Money                int64  // 余额
	Warehouse            int64  // 剩余仓库大小
	RefrigerateWarehouse int64  // 冷藏仓库大小
	Day                  int64  // 天
}

// TransactionRecord 数据库:transaction_record 订单表结构数据
type TransactionRecord struct {
	ID        int8   // ID
	Name      string // 名称
	ShopPrice int64  // 价格
	Number    int64  // 购买数量
	Type      int8   // 商品类型 1:普通
	ShopDay   int64  // 位于那一天
	OpType    int8   // 卖出还是买入 买入0卖出1
}

// Warehouse 数据库: warehouse 仓库表
type Warehouse struct {
	ID          int8   // ID
	Name        string // 名称
	Day         int64  // 位于那一天买入
	AvgPrice    int64  // 平均价格
	Number      int64  // 拥有的总数
	StorageDate int64  // 过期时间
}

// SpecialGoods 数据库: special_goods 特殊商品表
type SpecialGoods struct {
	ID             int8
	SpecialName    string
	SpecialPrice   int64
	SpecialContent string
}
