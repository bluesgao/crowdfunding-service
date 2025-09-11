package handler

import (
	"net/http"
	"strconv"

	"github.com/blues/cfs/internal/logic"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ContributeRecordHandler 贡献记录处理器
type ContributeRecordHandler struct {
	contributeLogic *logic.ContributeRecordLogic
}

// NewContributeRecordHandler 创建贡献记录处理器
func NewContributeRecordHandler(db *gorm.DB) *ContributeRecordHandler {
	return &ContributeRecordHandler{
		contributeLogic: logic.NewContributeRecordLogic(db),
	}
}

// GetProjectContributeRecords 获取项目贡献记录
func (h *ContributeRecordHandler) GetProjectContributeRecords(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用logic层获取项目贡献记录
	records, total, err := h.contributeLogic.GetProjectContributeRecords(uint(projectID), page, pageSize)
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
func (h *ContributeRecordHandler) GetUserContributeRecords(c *gin.Context) {
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
func (h *ContributeRecordHandler) GetContributeRecordByTxHash(c *gin.Context) {
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
func (h *ContributeRecordHandler) GetContributeStatistics(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 调用logic层获取统计信息
	stats, err := h.contributeLogic.GetContributeStatistics(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
