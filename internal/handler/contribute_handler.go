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
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用logic层获取项目贡献记录
	records, total, err := h.contributeLogic.GetProjectContributeRecords(int64(projectId), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUserContributeRecords 获取用户贡献记录
func (h *ContributeHandler) GetUserContributeRecords(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户地址不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用logic层获取用户贡献记录
	records, total, err := h.contributeLogic.GetUserContributeRecords(address, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetContributeRecordByTxHash 根据交易哈希获取贡献记录
func (h *ContributeHandler) GetContributeRecordByTxHash(c *gin.Context) {
	txHash := c.Param("tx_hash")
	if txHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "交易哈希不能为空"})
		return
	}

	// 调用logic层获取贡献记录
	record, err := h.contributeLogic.GetContributeRecordByTxHash(txHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": record,
	})
}

// GetContributeStatistics 获取贡献统计信息
func (h *ContributeHandler) GetContributeStatistics(c *gin.Context) {
	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 调用logic层获取统计信息
	stats, err := h.contributeLogic.GetContributeStatistics(int64(projectId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
