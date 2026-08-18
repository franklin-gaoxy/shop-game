package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// 存储仓库类型（与后端约定一致）
const (
	StorageNormal = 1 // 普通仓库
	StorageCold   = 2 // 冷藏仓库
	StorageBoth   = 3 // 两者皆可
)

// envelope 后端统一响应结构
type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Client 后端 HTTP 客户端
type Client struct {
	base       string
	token      string
	httpClient *http.Client
}

// NewClient 创建客户端，自动补全 http:// 前缀
func NewClient(base string) *Client {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	return &Client{base: base, httpClient: &http.Client{Timeout: 15 * time.Second}}
}

// do 发送请求并解析统一响应
func (c *Client) do(method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Token", c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败：%v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("响应解析失败（HTTP %d）", resp.StatusCode)
	}
	if env.Code != 0 {
		return fmt.Errorf("%s", env.Message)
	}
	if out != nil && len(env.Data) > 0 {
		return json.Unmarshal(env.Data, out)
	}
	return nil
}

// Health 连接检查（GET /healthz）
func (c *Client) Health() error {
	resp, err := c.httpClient.Get(c.base + "/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查返回状态码 %d", resp.StatusCode)
	}
	return nil
}

// ---------- 响应数据结构 ----------

// UserInfo 用户信息
type UserInfo struct {
	ID        uint    `json:"id"`
	Username  string  `json:"username"`
	Money     float64 `json:"money"`
	Day       int     `json:"day"`
	IsAdmin   bool    `json:"is_admin"`
	Status    int     `json:"status"` // 1 启用 0 禁用
	KeyID     uint    `json:"key_id"`
	CreatedAt string  `json:"created_at"`
}

// authResp 登录/注册响应
type authResp struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

// Product 商品元数据
type Product struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	StorageType      int     `json:"storage_type"`
	MinPrice         float64 `json:"min_price"`
	MaxPrice         float64 `json:"max_price"`
	Size             int64   `json:"size"`
	NormalExpireDays int     `json:"normal_expire_days"`
	ColdExpireDays   int     `json:"cold_expire_days"`
}

// ProductPrice 商品 + 今日价格
type ProductPrice struct {
	Product
	Price       float64 `json:"price"`
	CritApplied bool    `json:"crit_applied"` // 今日价格是否由暴击事件加成
}

// productsResp 商城响应
type productsResp struct {
	Day      int            `json:"day"`
	Products []ProductPrice `json:"products"`
}

// tradeResp 买入/卖出结果
type tradeResp struct {
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

// WarehouseItem 仓库商品
type WarehouseItem struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	StorageType int     `json:"storage_type"`
	Quantity    int     `json:"quantity"`
	TotalCost   float64 `json:"total_cost"`
	AvgPrice    float64 `json:"avg_price"`
	BuyDay      int     `json:"buy_day"`
	ExpireDay   int     `json:"expire_day"`
	ProductName string  `json:"product_name"`
	TodayPrice  float64 `json:"today_price"`
	TrendUp     bool    `json:"trend_up"`
	Profit      float64 `json:"profit"`
}

// warehouseResp 仓库响应
type warehouseResp struct {
	Day   int             `json:"day"`
	Items []WarehouseItem `json:"items"`
}

// StorageSpace 仓库空间统计
type StorageSpace struct {
	Capacity int64 `json:"capacity"`
	Used     int64 `json:"used"`
	Free     int64 `json:"free"`
}

// warehousePurchase 仓库空间购买记录
type warehousePurchase struct {
	ID          uint    `json:"id"`
	StorageType int     `json:"storage_type"`
	Size        int64   `json:"size"`
	Months      int     `json:"months"`
	UnitPrice   float64 `json:"unit_price"`
	StartDay    int     `json:"start_day"`
	ExpireDay   int     `json:"expire_day"`
}

// spaceResp 特殊商城响应
type spaceResp struct {
	Day             int                 `json:"day"`
	Normal          StorageSpace        `json:"normal"`
	Cold            StorageSpace        `json:"cold"`
	UnitPriceNormal float64             `json:"unit_price_normal"`
	UnitPriceCold   float64             `json:"unit_price_cold"`
	Purchases       []warehousePurchase `json:"purchases"`
}

// tomorrowResp 明天接口响应
type tomorrowResp struct {
	Day            int `json:"day"`
	TriggeredCrits []struct {
		Description string   `json:"description"`
		Products    []string `json:"products"`
	} `json:"triggered_crits"`
	ExpiredItems []struct {
		ProductID   uint    `json:"product_id"`
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		Value       float64 `json:"value"`
	} `json:"expired_items"`
	ExpiredPurchases int64 `json:"expired_purchases"`
}

// ---------- 接口封装 ----------

// Login 登录
func (c *Client) Login(username, password string) (*authResp, error) {
	var out authResp
	body := map[string]string{"username": username, "password": password}
	if err := c.do(http.MethodPost, "/api/login", body, &out); err != nil {
		return nil, err
	}
	c.token = out.Token
	return &out, nil
}

// Register 注册（成功后自动登录）
func (c *Client) Register(username, password, key string) (*authResp, error) {
	var out authResp
	body := map[string]string{"username": username, "password": password, "key": key}
	if err := c.do(http.MethodPost, "/api/register", body, &out); err != nil {
		return nil, err
	}
	c.token = out.Token
	return &out, nil
}

// Logout 退出登录
func (c *Client) Logout() error {
	return c.do(http.MethodPost, "/api/logout", nil, nil)
}

// Me 当前用户信息
func (c *Client) Me(out *UserInfo) error {
	return c.do(http.MethodGet, "/api/me", nil, out)
}

// Products 今日商品与价格
func (c *Client) Products() (*productsResp, error) {
	var out productsResp
	if err := c.do(http.MethodGet, "/api/products", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Buy 买入商品
func (c *Client) Buy(productID uint, quantity, storageType int) (*tradeResp, error) {
	body := map[string]interface{}{"product_id": productID, "quantity": quantity, "storage_type": storageType}
	var out tradeResp
	if err := c.do(http.MethodPost, "/api/buy", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Sell 卖出商品（storageType=0 表示全部仓库）
func (c *Client) Sell(productID uint, quantity, storageType int) (*tradeResp, error) {
	body := map[string]interface{}{"product_id": productID, "quantity": quantity, "storage_type": storageType}
	var out tradeResp
	if err := c.do(http.MethodPost, "/api/sell", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Tomorrow 推进到明天
func (c *Client) Tomorrow() (*tomorrowResp, error) {
	var out tomorrowResp
	if err := c.do(http.MethodPost, "/api/tomorrow", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Warehouse 仓库商品列表
func (c *Client) Warehouse() (*warehouseResp, error) {
	var out warehouseResp
	if err := c.do(http.MethodGet, "/api/warehouse", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WarehouseSpace 仓库空间信息
func (c *Client) WarehouseSpace() (*spaceResp, error) {
	var out spaceResp
	if err := c.do(http.MethodGet, "/api/warehouse/space", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WarehouseBuy 购买仓库空间
func (c *Client) WarehouseBuy(storageType int, size, months int64) error {
	body := map[string]interface{}{"storage_type": storageType, "size": size, "months": months}
	return c.do(http.MethodPost, "/api/warehouse/buy", body, nil)
}

// Transactions 交易记录（分页）
func (c *Client) Transactions(page, pageSize int) (*TxListResp, error) {
	var out TxListResp
	path := fmt.Sprintf("/api/transactions?page=%d&page_size=%d", page, pageSize)
	if err := c.do(http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- 管理员：密钥管理 ----------

// RegKey 注册密钥
type RegKey struct {
	ID        uint   `json:"id"`
	Key       string `json:"key"`
	MaxUses   int    `json:"max_uses"`
	UsedCount int    `json:"used_count"`
	CreatedBy string `json:"created_by"`
	Status    int    `json:"status"`
}

// ListKeys 查询全部密钥
func (c *Client) ListKeys() ([]RegKey, error) {
	var out []RegKey
	if err := c.do(http.MethodGet, "/api/admin/keys", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateKey 创建密钥
func (c *Client) CreateKey(maxUses int) (*RegKey, error) {
	var out RegKey
	body := map[string]int{"max_uses": maxUses}
	if err := c.do(http.MethodPost, "/api/admin/keys", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteKey 删除密钥
func (c *Client) DeleteKey(id uint) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/api/admin/keys/%d", id), nil, nil)
}

// KeyUsers 查询密钥绑定的用户
func (c *Client) KeyUsers(id uint) ([]UserInfo, error) {
	var out []UserInfo
	if err := c.do(http.MethodGet, fmt.Sprintf("/api/admin/keys/%d/users", id), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- 管理员：用户管理 ----------

// UserListResp 用户分页响应
type UserListResp struct {
	List     []UserInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// ListUsers 查询全部用户（分页）
func (c *Client) ListUsers(page, pageSize int) (*UserListResp, error) {
	var out UserListResp
	path := fmt.Sprintf("/api/admin/users?page=%d&page_size=%d", page, pageSize)
	if err := c.do(http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetUserStatus 禁用/启用用户（status: 0 禁用 1 启用）
func (c *Client) SetUserStatus(id uint, status int) error {
	body := map[string]int{"status": status}
	return c.do(http.MethodPut, fmt.Sprintf("/api/admin/users/%d/status", id), body, nil)
}

// DeleteUser 删除用户
func (c *Client) DeleteUser(id uint) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/api/admin/users/%d", id), nil, nil)
}

// ---------- 交易记录 ----------

// Transaction 单条交易记录
type Transaction struct {
	ID           uint    `json:"id"`
	Day          int     `json:"day"`
	Type         string  `json:"type"`
	Direction    int     `json:"direction"`
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	Amount       float64 `json:"amount"`
	BalanceAfter float64 `json:"balance_after"`
	CreatedAt    string  `json:"created_at"`
}

// TxListResp 交易记录分页响应
type TxListResp struct {
	List     []Transaction `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}
