package handler

import (
	"net/http"
	"strconv"

	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/model"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectLogic *logic.ProjectLogic
}

func NewProjectHandler(projectLogic *logic.ProjectLogic) *ProjectHandler {
	return &ProjectHandler{
		projectLogic: projectLogic,
	}
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var project model.ProjectModel
	if err := c.ShouldBindJSON(&project); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 调用logic层创建项目
	if err := h.projectLogic.CreateProject(&project); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, http.StatusCreated, "项目创建成功", CreateProjectResponse{Project: ToProjectResponse(&project)})
}

// GetProjects 获取项目列表
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	// 调用logic层获取所有项目
	projects, err := h.projectLogic.GetProjects()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "获取项目列表成功", GetProjectsResponse{Projects: ToProjectResponseList(projects)})
}

// GetProject 获取单个项目详情
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "无效的项目ID")
		return
	}

	// 调用logic层获取项目详情
	project, err := h.projectLogic.GetProject(int64(id))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "获取项目详情成功", GetProjectResponse{Project: ToProjectResponse(project)})
}

// GetProjectStats 获取项目统计信息
func (h *ProjectHandler) GetProjectStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "无效的项目ID")
		return
	}

	// 调用logic层获取项目统计信息
	stats, err := h.projectLogic.GetProjectStats(int64(id))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "获取项目统计信息成功", GetProjectStatsResponse{Stats: stats})
}

// GetAllProjectStats 获取所有项目的统计信息
func (h *ProjectHandler) GetAllProjectStats(c *gin.Context) {
	stats, err := h.projectLogic.GetAllProjectStats()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, http.StatusOK, "获取所有项目统计信息成功", GetAllProjectStatsResponse{Stats: ToAllProjectStatsResponse(stats)})
}
