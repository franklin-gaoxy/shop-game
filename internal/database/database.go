// Package database 定义数据存储抽象接口与业务结果类型。
// 当前提供 MySQL 实现（internal/database/mysql，通过 init 注册），
// 后续如需支持其他数据库，只需实现本接口并注册即可，上层代码无需改动。
package database

import (
	"errors"
	"fmt"
	"strings"

	"trade_game/internal/config"
	"trade_game/internal/model"
)

// ErrAlreadyInitialized 表示数据库已经初始化过
var ErrAlreadyInitialized = errors.New("数据库已初始化，如需重新初始化请使用 --force 参数")

// Database 数据存储接口（领域操作）
type Database interface {
	Close() error
	Ping() error

	// 初始化
	IsInitialized() (bool, error)
	Initialize(sqlFile string, force bool) (bool, error)
	GenerateKey() (string, error)
	ImportInitData(data *InitData, adminUser, adminPassword string) (string, error)

	// 用户与密钥
	LoginCheck(username, passwordMD5 string) (*model.User, error)
	GetUser(id uint) (*model.User, error)
	Register(username, passwordMD5, regKey string) (*model.User, error)
	CreateRegKey(maxUses int, createdBy string) (*model.RegKey, error)
	ListRegKeys() ([]model.RegKey, error)

	// 商品与交易
	ListProducts() ([]model.Product, error)
	GetTodayPrices(userID uint) (*ProductPriceResult, error)
	Buy(userID, productID uint, quantity, storageType int) (*TradeResult, error)
	Sell(userID, productID uint, quantity, storageType int) (*TradeResult, error)
	AdvanceDay(userID uint) (*AdvanceResult, error)

	// 仓库
	ListWarehouseItems(userID uint) (*WarehouseItemsResult, error)
	GetWarehouseSpace(userID uint) (*WarehouseSpace, error)
	BuyWarehouseSpace(userID uint, storageType int, size int64, months int) (*model.Transaction, error)

	// 交易记录
	ListTransactions(userID uint, page, pageSize int) ([]model.Transaction, int64, error)
}

// ---------- 数据库驱动注册（类似 database/sql 的注册模式） ----------

var factories = map[string]func(cfg config.DatabaseConfig) (Database, error){}

// RegisterFactory 注册数据库实现
func RegisterFactory(dbType string, factory func(cfg config.DatabaseConfig) (Database, error)) {
	factories[strings.ToLower(dbType)] = factory
}

// New 根据配置创建数据库实现
func New(cfg config.DatabaseConfig) (Database, error) {
	dbType := strings.ToLower(cfg.Type)
	if dbType == "" {
		dbType = "mysql"
	}
	factory, ok := factories[dbType]
	if !ok {
		return nil, fmt.Errorf("暂不支持的数据库类型: %s", cfg.Type)
	}
	return factory(cfg)
}

// ---------- 初始化数据结构（对应 data/init_data.yaml） ----------

type InitDefaults struct {
	DefaultMoney         float64 `yaml:"default_money"`
	DefaultWarehouse     int64   `yaml:"default_warehouse"`
	DefaultColdStorage   int64   `yaml:"default_cold_storage"`
	WarehousePriceNormal float64 `yaml:"warehouse_price_normal"` // 普通仓库每单位空间每月价格
	WarehousePriceCold   float64 `yaml:"warehouse_price_cold"`   // 冷藏仓库每单位空间每月价格
	InitialKey           string  `yaml:"initial_key"`            // 指定初始密钥，为空则随机生成
	InitialKeyMaxUses    int     `yaml:"initial_key_max_uses"`   // 初始密钥可用次数
}

type InitProduct struct {
	Name             string  `yaml:"name"`
	Category         string  `yaml:"category"`
	StorageType      string  `yaml:"storage_type"` // normal/cold/both
	MinPrice         float64 `yaml:"min_price"`
	MaxPrice         float64 `yaml:"max_price"`
	Size             int64   `yaml:"size"`
	NormalExpireDays int     `yaml:"normal_expire_days"`
	ColdExpireDays   int     `yaml:"cold_expire_days"`
	CritMin          float64 `yaml:"crit_min"`
	CritMax          float64 `yaml:"crit_max"`
}

type InitCritEvent struct {
	Description string `yaml:"description"`
	TargetType  string `yaml:"target_type"` // category/product
	TargetValue string `yaml:"target_value"`
	Probability int    `yaml:"probability"`
}

type InitData struct {
	Defaults   InitDefaults    `yaml:"defaults"`
	Products   []InitProduct   `yaml:"products"`
	CritEvents []InitCritEvent `yaml:"crit_events"`
}

// ---------- 接口返回的业务结果结构 ----------

// ProductPrice 商品 + 当日价格
type ProductPrice struct {
	model.Product
	Price float64 `json:"price"`
}

// ProductPriceResult 当日商城价格
type ProductPriceResult struct {
	Day      int            `json:"day"`
	Products []ProductPrice `json:"products"`
}

// TradeResult 买入/卖出结果
type TradeResult struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	StorageType int     `json:"storage_type"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
	Balance     float64 `json:"balance"`
	Day         int     `json:"day"`
	ExpireDay   int     `json:"expire_day"`
}

// CritTriggerInfo 触发的暴击事件
type CritTriggerInfo struct {
	Description string   `json:"description"`
	Products    []string `json:"products"`
}

// ExpiredItemInfo 过期商品信息
type ExpiredItemInfo struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Value       float64 `json:"value"` // 损失价值
}

// AdvanceResult “明天”接口结果
type AdvanceResult struct {
	Day              int               `json:"day"`
	Prices           []ProductPrice    `json:"prices"`
	TriggeredCrits   []CritTriggerInfo `json:"triggered_crits"`
	ExpiredItems     []ExpiredItemInfo `json:"expired_items"`
	ExpiredPurchases int64             `json:"expired_purchases"` // 到期失效的仓库空间购买记录数
}

// StorageSpace 单类仓库空间统计
type StorageSpace struct {
	Capacity int64 `json:"capacity"` // 总容量（默认 + 有效期内购买）
	Used     int64 `json:"used"`     // 已用
	Free     int64 `json:"free"`     // 剩余
}

// WarehouseSpace 仓库空间信息（特殊商城）
type WarehouseSpace struct {
	Day             int                       `json:"day"`
	Normal          StorageSpace              `json:"normal"`
	Cold            StorageSpace              `json:"cold"`
	UnitPriceNormal float64                   `json:"unit_price_normal"`
	UnitPriceCold   float64                   `json:"unit_price_cold"`
	Purchases       []model.WarehousePurchase `json:"purchases"`
}

// WarehouseItemInfo 仓库商品详情（含今日价格与预计盈亏）
type WarehouseItemInfo struct {
	model.WarehouseItem
	ProductName string  `json:"product_name"`
	Category    string  `json:"category"`
	TodayPrice  float64 `json:"today_price"`
	TrendUp     bool    `json:"trend_up"` // 今日价格是否高于均价
	Profit      float64 `json:"profit"`   // 全部卖出的预计获利(正)/亏损(负)
}

// WarehouseItemsResult 仓库列表
type WarehouseItemsResult struct {
	Day   int                 `json:"day"`
	Items []WarehouseItemInfo `json:"items"`
}
