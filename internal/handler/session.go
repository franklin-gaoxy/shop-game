package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// session 登录会话（内存存储）
type session struct {
	UserID   uint
	Username string
	IsAdmin  bool
}

type sessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*session
}

func newSessionManager() *sessionManager {
	return &sessionManager{sessions: map[string]*session{}}
}

func (s *sessionManager) create(userID uint, username string, isAdmin bool) string {
	token := randomToken()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = &session{UserID: userID, Username: username, IsAdmin: isAdmin}
	return token
}

func (s *sessionManager) get(token string) *session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[token]
}

func (s *sessionManager) remove(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// tokenFromRequest 从 X-Token 或 Authorization: Bearer 头中取令牌
func tokenFromRequest(c *gin.Context) string {
	token := c.GetHeader("X-Token")
	if token == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	return token
}

// Auth 登录校验中间件：检查请求是否携带合法令牌
func (h *Handler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromRequest(c)
		if token == "" {
			fail(c, http.StatusUnauthorized, "未登录，请先登录")
			c.Abort()
			return
		}
		s := h.sessions.get(token)
		if s == nil {
			fail(c, http.StatusUnauthorized, "登录已失效，请重新登录")
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Set("user_id", s.UserID)
		c.Set("username", s.Username)
		c.Set("is_admin", s.IsAdmin)
		c.Next()
	}
}

// AdminRequired 管理员权限中间件（需在 Auth 之后）
func (h *Handler) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, _ := c.Get("is_admin")
		isAdmin, _ := v.(bool)
		if !isAdmin {
			fail(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
