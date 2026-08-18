package handler

import (
	"math"

	"github.com/gin-gonic/gin"
	"net/http"

	"trade_game/internal/database"
)

// Home 首页：剩余金钱、天数、仓库使用率等
func (h *Handler) Home(c *gin.Context) {
	userID := currentUserID(c)
	user, err := h.db.GetUser(userID)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	space, err := h.db.GetWarehouseSpace(userID)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	usage := func(s database.StorageSpace) float64 {
		if s.Capacity <= 0 {
			return 0
		}
		return math.Round(float64(s.Used)/float64(s.Capacity)*1000) / 10
	}
	ok(c, gin.H{
		"username": user.Username,
		"money":    user.Money,
		"day":      user.Day,
		"normal": gin.H{
			"capacity":      space.Normal.Capacity,
			"used":          space.Normal.Used,
			"free":          space.Normal.Free,
			"usage_percent": usage(space.Normal),
		},
		"cold": gin.H{
			"capacity":      space.Cold.Capacity,
			"used":          space.Cold.Used,
			"free":          space.Cold.Free,
			"usage_percent": usage(space.Cold),
		},
	})
}
