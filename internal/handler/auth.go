package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trade_game/internal/model"
)

type registerReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Key      string `json:"key" binding:"required"`
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func userView(u *model.User) gin.H {
	return gin.H{
		"id":         u.ID,
		"username":   u.Username,
		"money":      u.Money,
		"day":        u.Day,
		"is_admin":   u.IsAdmin,
		"key_id":     u.KeyID,
		"created_at": u.CreatedAt,
	}
}

// Register 用户注册：用户名不可重复，密码 md5 存储，必须携带有效密钥
func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误：用户名、密码、密钥均不能为空")
		return
	}
	user, err := h.db.Register(req.Username, model.MD5Hash(req.Password), req.Key)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	token := h.sessions.create(user.ID, user.Username, user.IsAdmin)
	ok(c, gin.H{"token": token, "user": userView(user)})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误：用户名和密码不能为空")
		return
	}
	user, err := h.db.LoginCheck(req.Username, model.MD5Hash(req.Password))
	if err != nil {
		fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	token := h.sessions.create(user.ID, user.Username, user.IsAdmin)
	ok(c, gin.H{"token": token, "user": userView(user)})
}

// Logout 退出登录
func (h *Handler) Logout(c *gin.Context) {
	if token, _ := c.Get("token"); token != nil {
		h.sessions.remove(token.(string))
	}
	ok(c, nil)
}

// Me 当前用户信息
func (h *Handler) Me(c *gin.Context) {
	user, err := h.db.GetUser(currentUserID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, userView(user))
}
