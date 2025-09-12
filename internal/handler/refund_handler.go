package handler

import (
	"net/http"
	"strconv"

	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/model"
	"github.com/gin-gonic/gin"
)

// RefundHandler 退款处理器
type RefundHandler struct {
	refundLogic *logic.RefundRecordLogic
}

// NewRefundHandler 创建退款处理器
func NewRefundHandler(refundLogic *logic.RefundRecordLogic) *RefundHandler {
	return &RefundHandler{
		refundLogic: refundLogic,
	}
}

// GetProjectRefunds 获取项目退款记录
func (h *RefundHandler) GetProjectRefunds(c *gin.Context) {
	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用logic层获取项目退款记录
	refunds, total, err := h.refundLogic.GetProjectRefunds(int64(projectId), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": refunds,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUserRefunds 获取用户退款记录
func (h *RefundHandler) GetUserRefunds(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户地址不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用logic层获取用户退款记录
	refunds, total, err := h.refundLogic.GetUserRefunds(address, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": refunds,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetRefundByTxHash 根据交易哈希获取退款记录
func (h *RefundHandler) GetRefundByTxHash(c *gin.Context) {
	txHash := c.Param("tx_hash")
	if txHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "交易哈希不能为空"})
		return
	}

	// 调用logic层获取退款记录
	refund, err := h.refundLogic.GetRefundByTxHash(txHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": refund,
	})
}

// UpdateRefundStatus 更新退款状态
func (h *RefundHandler) UpdateRefundStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的退款记录ID"})
		return
	}

	var request struct {
		Status model.RefundStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用logic层更新退款状态
	if err := h.refundLogic.UpdateRefundStatus(int64(id), request.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "退款状态更新成功",
	})
}

// GetRefundStatistics 获取退款统计信息
func (h *RefundHandler) GetRefundStatistics(c *gin.Context) {
	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 调用logic层获取统计信息
	stats, err := h.refundLogic.GetRefundStatistics(int64(projectId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
