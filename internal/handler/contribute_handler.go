package handler

import (
	"net/http"
	"strconv"

	"github.com/blues/cfs/internal/logic"
	"github.com/gin-gonic/gin"
)

// ContributeHandler 贡献处理器
type ContributeHandler struct {
	contributeLogic *logic.ContributeRecordLogic
}

// NewContributeHandler 创建贡献处理器
func NewContributeHandler(contributeLogic *logic.ContributeRecordLogic) *ContributeHandler {
	return &ContributeHandler{
		contributeLogic: contributeLogic,
	}
}

// GetProjectContributeRecords 获取项目贡献记录
func (h *ContributeHandler) GetProjectContributeRecords(c *gin.Context) {
	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "无效的项目ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用logic层获取项目贡献记录
	records, total, err := h.contributeLogic.GetProjectContributeRecords(int64(projectId), page, pageSize)
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

	SuccessResponse(c, http.StatusOK, "获取项目贡献记录成功", GetProjectContributeRecordsResponse{
		Records:    ToContributeRecordResponseList(records),
		Pagination: pagination,
	})
}

// GetContributeStats 获取贡献统计信息
func (h *ContributeHandler) GetContributeStats(c *gin.Context) {
	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "无效的项目ID")
		return
	}

	// 调用logic层获取统计信息
	stats, err := h.contributeLogic.GetContributeStats(int64(projectId))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "获取贡献统计信息成功", GetContributeStatsResponse{Stats: stats})
}
