// Package handler 实现 gin HTTP 接口层。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trade_game/internal/database"
)

// Handler 汇总各接口实现
type Handler struct {
	db       database.Database
	sessions *sessionManager
}

// NewRouter 构建 gin 路由
func NewRouter(db database.Database) *gin.Engine {
	h := &Handler{db: db, sessions: newSessionManager()}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := r.Group("/api")
	{
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)
	}

	auth := api.Group("", h.Auth())
	{
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.Me)
		auth.GET("/home", h.Home)
		auth.GET("/products", h.Products)
		auth.POST("/buy", h.Buy)
		auth.POST("/sell", h.Sell)
		auth.GET("/warehouse", h.Warehouse)
		auth.GET("/warehouse/space", h.WarehouseSpace)
		auth.POST("/warehouse/buy", h.WarehouseBuy)
		auth.POST("/tomorrow", h.Tomorrow)
		auth.GET("/transactions", h.Transactions)
	}

	admin := api.Group("/admin", h.Auth(), h.AdminRequired())
	{
		admin.POST("/keys", h.CreateKey)
		admin.GET("/keys", h.ListKeys)
		admin.DELETE("/keys/:id", h.DeleteKey)
		admin.GET("/keys/:id/users", h.KeyUsers)
	}

	return r
}

// ok 统一成功响应
func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

// fail 统一失败响应
func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"code": -1, "message": msg})
}

// currentUserID 从上下文取当前登录用户 ID
func currentUserID(c *gin.Context) uint {
	v, _ := c.Get("user_id")
	id, _ := v.(uint)
	return id
}
