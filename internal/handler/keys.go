package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"trade_game/internal/model"
)

type createKeyReq struct {
	MaxUses int `json:"max_uses"` // 密钥可用次数，默认 10
}

// parseKeyID 从路径参数解析密钥 ID
func parseKeyID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		fail(c, http.StatusBadRequest, "密钥 ID 无效")
		return 0, false
	}
	return uint(id), true
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

// DeleteKey 管理员删除密钥（已被使用注册的密钥不可删除）
func (h *Handler) DeleteKey(c *gin.Context) {
	id, valid := parseKeyID(c)
	if !valid {
		return
	}
	if err := h.db.DeleteRegKey(id); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, nil)
}

// KeyUsers 查询使用某密钥注册的用户列表
func (h *Handler) KeyUsers(c *gin.Context) {
	id, valid := parseKeyID(c)
	if !valid {
		return
	}
	users, err := h.db.ListKeyUsers(id)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if users == nil {
		users = []model.User{}
	}
	ok(c, users)
}
