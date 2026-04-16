package mcp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ToolHandler 工具管理的 HTTP Handler
type ToolHandler struct {
	db           *gorm.DB
	serviceStore  repository.ServiceStore
	logger       logger.Logger
}

// NewToolHandler 创建 ToolHandler
func NewToolHandler(db *gorm.DB, serviceStore repository.ServiceStore, log logger.Logger) *ToolHandler {
	return &ToolHandler{
		db:          db,
		serviceStore: serviceStore,
		logger:      log,
	}
}

// ListTools GET /api/admin/tools - 列表
func (h *ToolHandler) ListTools(ctx *gin.Context) {
	var tools []model.ToolDefinition
	result := h.db.WithContext(ctx.Request.Context()).
		Where("enabled = ?", true).
		Find(&tools)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list tools",
			"details": result.Error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tools": tools,
		"count": len(tools),
	})
}

// GetTool GET /api/admin/tools/:id - 详情
func (h *ToolHandler) GetTool(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tool id",
			"id":   idParam,
		})
		return
	}

	var tool model.ToolDefinition
	result := h.db.WithContext(ctx.Request.Context()).
		Where("id = ? AND enabled = ?", id, true).
		First(&tool)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "tool not found",
				"id":   idParam,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get tool",
			"details": result.Error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tool": tool,
	})
}

// CreateToolRequest 创建工具请求
type CreateToolRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Parameters  []model.ParameterDefinition `json:"parameters"`
	VAuthKey    string                  `json:"vauth_key" binding:"required"`
	ServerDesc  string                  `json:"server_desc"`
	ServiceID   uint                    `json:"service_id"`
	InputExtra  json.RawMessage         `json:"input_extra"`
}

// CreateTool POST /api/admin/tools - 创建
func (h *ToolHandler) CreateTool(ctx *gin.Context) {
	var req CreateToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	tool := model.ToolDefinition{
		Name:        req.Name,
		Description: req.Description,
		Parameters:  req.Parameters,
		Enabled:     true,
		VAuthKey:    req.VAuthKey,
		ServerDesc:  req.ServerDesc,
		ServiceID:   req.ServiceID,
		InputExtra:  req.InputExtra,
	}

	result := h.db.WithContext(ctx.Request.Context()).Create(&tool)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create tool",
			"details": result.Error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "tool created successfully",
		"tool":   tool,
	})
}

// UpdateToolRequest 更新工具请求
type UpdateToolRequest struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Parameters    []model.ParameterDefinition `json:"parameters"`
	Enabled       *bool                   `json:"enabled"`
	ServerDesc    string                  `json:"server_desc"`
	InputExtra    json.RawMessage         `json:"input_extra"`
	OutputMapping json.RawMessage         `json:"output_mapping"`
}

// UpdateTool PUT /api/admin/tools/:id - 更新
func (h *ToolHandler) UpdateTool(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tool id",
			"id":   idParam,
		})
		return
	}

	var tool model.ToolDefinition
	result := h.db.WithContext(ctx.Request.Context()).
		Where("id = ?", id).
		First(&tool)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "tool not found",
				"id":   idParam,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get tool",
			"details": result.Error.Error(),
		})
		return
	}

	var req UpdateToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if req.Name != "" {
		tool.Name = req.Name
	}
	if req.Description != "" {
		tool.Description = req.Description
	}
	if req.Parameters != nil {
		tool.Parameters = req.Parameters
	}
	if req.Enabled != nil {
		tool.Enabled = *req.Enabled
	}
	if req.ServerDesc != "" {
		tool.ServerDesc = req.ServerDesc
	}
	if req.InputExtra != nil {
		tool.InputExtra = req.InputExtra
	}
	if req.OutputMapping != nil {
		tool.OutputMapping = req.OutputMapping
	}

	result = h.db.WithContext(ctx.Request.Context()).Save(&tool)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update tool",
			"details": result.Error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "tool updated successfully",
	})
}

// DeleteTool DELETE /api/admin/tools/:id - 删除
func (h *ToolHandler) DeleteTool(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tool id",
			"id":   idParam,
		})
		return
	}

	// 软删除
	result := h.db.WithContext(ctx.Request.Context()).
		Model(&model.ToolDefinition{}).
		Where("id = ?", id).
		Update("enabled", false)

	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete tool",
			"details": result.Error.Error(),
		})
		return
	}
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "tool not found",
			"id":   idParam,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "tool deleted successfully",
	})
}

// GetHTTPServiceOutputSchema GET /api/admin/http-services/:id/output-schema - 获取指定服务的 OutputSchema
func (h *ToolHandler) GetHTTPServiceOutputSchema(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service id",
			"id":   idParam,
		})
		return
	}

	service, err := h.serviceStore.Get(ctx.Request.Context(), uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "service not found",
			"details": err.Error(),
			"id":      idParam,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"service_id":    service.ID,
		"name":          service.Name,
		"output_schema": service.OutputSchema,
	})
}
