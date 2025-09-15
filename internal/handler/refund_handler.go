package handler

import (
	"net/http"
	"strconv"

	"github.com/blues/cfs/internal/logic"
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
		ErrorResponse(c, http.StatusBadRequest, "无效的项目ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用logic层获取项目退款记录
	refunds, total, err := h.refundLogic.GetProjectRefunds(int64(projectId), page, pageSize)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	pagination := Pagination{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: (total + int64(pageSize) - 1) / int64(pageSize),
	}

	SuccessResponse(c, http.StatusOK, "获取项目退款记录成功", GetProjectRefundsResponse{
		Refunds:    ToRefundRecordResponseList(refunds),
		Pagination: pagination,
	})
}

// GetRefundStats 获取退款统计信息
func (h *RefundHandler) GetRefundStats(c *gin.Context) {
	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "无效的项目ID")
		return
	}

	// 调用logic层获取统计信息
	stats, err := h.refundLogic.GetRefundStats(int64(projectId))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "获取退款统计信息成功", GetRefundStatsResponse{Stats: stats})
}
