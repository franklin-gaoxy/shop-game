package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type warehouseBuyReq struct {
	StorageType int   `json:"storage_type" binding:"required"` // 1=普通仓库 2=冷藏仓库
	Size        int64 `json:"size" binding:"required"`         // 购买空间大小
	Months      int   `json:"months" binding:"required"`       // 购买月数 1-12（每月按 30 天）
}

// Warehouse 仓库：已存商品、均价、今日价格与预计盈亏
func (h *Handler) Warehouse(c *gin.Context) {
	res, err := h.db.ListWarehouseItems(currentUserID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, res)
}

// WarehouseSpace 特殊商城：仓库空间大小与购买记录
func (h *Handler) WarehouseSpace(c *gin.Context) {
	res, err := h.db.GetWarehouseSpace(currentUserID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, res)
}

// WarehouseBuy 特殊商城：购买仓库空间（1-12 个月）
func (h *Handler) WarehouseBuy(c *gin.Context) {
	var req warehouseBuyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误：storage_type、size、months 不能为空")
		return
	}
	tr, err := h.db.BuyWarehouseSpace(currentUserID(c), req.StorageType, req.Size, req.Months)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, tr)
}
