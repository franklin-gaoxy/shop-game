package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trade_game/internal/model"
)

type createKeyReq struct {
	MaxUses int `json:"max_uses"` // 密钥可用次数，默认 10
}

// CreateKey 管理员创建注册密钥
func (h *Handler) CreateKey(c *gin.Context) {
	var req createKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		req.MaxUses = 10
	}
	username, _ := c.Get("username")
	createdBy, _ := username.(string)
	key, err := h.db.CreateRegKey(req.MaxUses, createdBy)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, key)
}

// ListKeys 管理员查看全部密钥
func (h *Handler) ListKeys(c *gin.Context) {
	keys, err := h.db.ListRegKeys()
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if keys == nil {
		keys = []model.RegKey{}
	}
	ok(c, keys)
}
