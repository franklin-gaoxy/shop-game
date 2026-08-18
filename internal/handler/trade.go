package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"trade_game/internal/model"
)

type tradeReq struct {
	ProductID   uint `json:"product_id" binding:"required"`
	Quantity    int  `json:"quantity" binding:"required"`
	StorageType int  `json:"storage_type"` // 0=自动 1=普通仓库 2=冷藏仓库
}

// Products 商城：今日全部商品与随机价格
func (h *Handler) Products(c *gin.Context) {
	res, err := h.db.GetTodayPrices(currentUserID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, res)
}

// Buy 买入商品
func (h *Handler) Buy(c *gin.Context) {
	var req tradeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误：product_id 与 quantity 不能为空")
		return
	}
	res, err := h.db.Buy(currentUserID(c), req.ProductID, req.Quantity, req.StorageType)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, res)
}

// Sell 卖出商品
func (h *Handler) Sell(c *gin.Context) {
	var req tradeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误：product_id 与 quantity 不能为空")
		return
	}
	res, err := h.db.Sell(currentUserID(c), req.ProductID, req.Quantity, req.StorageType)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, res)
}

// Tomorrow 进入明天：天数+1、重新生成价格、判定暴击、清理过期
func (h *Handler) Tomorrow(c *gin.Context) {
	res, err := h.db.AdvanceDay(currentUserID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, res)
}

// Transactions 交易记录（分页，默认每页 20 条）
func (h *Handler) Transactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.db.ListTransactions(currentUserID(c), page, pageSize)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if list == nil {
		list = []model.Transaction{}
	}
	ok(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}
