package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"trade_game/internal/model"
)

type setUserStatusReq struct {
	// 注意：不能用 binding:"required"，0（禁用）是零值会被误判为缺失
	Status int `json:"status"` // 0 禁用 1 启用
}

// parseUserID 从路径参数解析用户 ID
func parseUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		fail(c, http.StatusBadRequest, "用户 ID 无效")
		return 0, false
	}
	return uint(id), true
}

// ListUsers 管理员查看全部用户（分页）
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	users, total, err := h.db.ListUsers(page, pageSize)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if users == nil {
		users = []model.User{}
	}
	ok(c, gin.H{"list": users, "total": total, "page": page, "page_size": pageSize})
}

// UpdateUserStatus 管理员禁用/启用用户（禁用即时踢下线）
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	id, valid := parseUserID(c)
	if !valid {
		return
	}
	var req setUserStatusReq
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != 0 && req.Status != 1) {
		fail(c, http.StatusBadRequest, "参数错误：status 仅支持 0=禁用 1=启用")
		return
	}
	if err := h.db.SetUserStatus(currentUserID(c), id, req.Status); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Status == 0 {
		h.sessions.removeByUser(id) // 禁用后立即失效其全部登录会话
	}
	ok(c, nil)
}

// DeleteUser 管理员删除用户及其关联数据（删除后立即踢下线）
func (h *Handler) DeleteUser(c *gin.Context) {
	id, valid := parseUserID(c)
	if !valid {
		return
	}
	if err := h.db.DeleteUser(currentUserID(c), id); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	h.sessions.removeByUser(id)
	ok(c, nil)
}
