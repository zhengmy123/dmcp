package mcp

import (
	"fmt"
	"strconv"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ToolHandler struct {
	db           *gorm.DB
	serviceStore repository.ServiceStore
	logger       logger.Logger
}

func NewToolHandler(db *gorm.DB, serviceStore repository.ServiceStore, log logger.Logger) *ToolHandler {
	return &ToolHandler{
		db:           db,
		serviceStore: serviceStore,
		logger:       log,
	}
}

type ToolListResponse struct {
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Tools    []ToolItemResponse `json:"tools"`
}

type ToolItemResponse struct {
	ID            uint                         `json:"id"`
	Name          string                       `json:"name"`
	Description   string                       `json:"description"`
	ServiceID     uint                         `json:"service_id"`
	Parameters    []model.ParameterDefinition  `json:"parameters"`
	InputMapping  []model.InputMappingField    `json:"input_mapping"`
	OutputMapping []model.OutputMappingField   `json:"output_mapping"`
	Enabled       bool                         `json:"enabled"`
	CreatedAt     string                       `json:"created_at"`
	UpdatedAt     string                       `json:"updated_at"`
}

func (h *ToolHandler) ListTools(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	keyword := ctx.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	var tools []model.ToolDefinition

	query := h.db.WithContext(ctx.Request.Context()).Model(&model.ToolDefinition{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&tools).Error; err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	items := make([]ToolItemResponse, 0, len(tools))
	for _, t := range tools {
		params := h.parseParameters(t.Parameters)
		inputMapping := h.parseInputMapping(t.InputMapping)
		outputMapping := h.parseOutputMapping(t.OutputMapping)
		items = append(items, ToolItemResponse{
			ID:            t.ID,
			Name:          t.Name,
			Description:   t.Description,
			ServiceID:     t.ServiceID,
			Parameters:    params,
			InputMapping:  inputMapping,
			OutputMapping: outputMapping,
			Enabled:       t.Enabled,
			CreatedAt:     t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.Success(ctx, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"tools":     items,
	})
}

func (h *ToolHandler) parseParameters(data []byte) []model.ParameterDefinition {
	if len(data) == 0 {
		return nil
	}
	var params []model.ParameterDefinition
	if err := sonic.Unmarshal(data, &params); err != nil {
		return nil
	}
	return params
}

func (h *ToolHandler) parseInputMapping(data []byte) []model.InputMappingField {
	if len(data) == 0 {
		return nil
	}
	var mapping []model.InputMappingField
	if err := sonic.Unmarshal(data, &mapping); err != nil {
		return nil
	}
	return mapping
}

func (h *ToolHandler) parseOutputMapping(data []byte) []model.OutputMappingField {
	if len(data) == 0 {
		return nil
	}
	var mapping []model.OutputMappingField
	if err := sonic.Unmarshal(data, &mapping); err != nil {
		return nil
	}
	return mapping
}

func (h *ToolHandler) serializeParameters(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	var params []model.ParameterDefinition
	if err := sonic.Unmarshal(data, &params); err != nil {
		return nil
	}
	result, _ := sonic.Marshal(params)
	return result
}

func (h *ToolHandler) GetTool(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid tool id")
		return
	}

	var tool model.ToolDefinition
	result := h.db.WithContext(ctx.Request.Context()).
		Where("id = ?", id).
		First(&tool)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			response.NotFound(ctx, "tool not found")
			return
		}
		response.InternalError(ctx, result.Error.Error())
		return
	}

	params := h.parseParameters(tool.Parameters)
	inputMapping := h.parseInputMapping(tool.InputMapping)
	outputMapping := h.parseOutputMapping(tool.OutputMapping)

	response.Success(ctx, gin.H{
		"tool": gin.H{
			"id":             tool.ID,
			"name":           tool.Name,
			"description":    tool.Description,
			"service_id":     tool.ServiceID,
			"parameters":     params,
			"input_mapping":  inputMapping,
			"output_mapping": outputMapping,
			"enabled":        tool.Enabled,
			"created_at":     tool.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":     tool.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

type CreateToolRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	ServiceID     uint   `json:"service_id"`
	Parameters    []byte `json:"parameters"`
	InputMapping  []byte `json:"input_mapping"`
	OutputMapping []byte `json:"output_mapping"`
}

func (h *ToolHandler) CreateTool(ctx *gin.Context) {
	var req CreateToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	if err := model.ValidateToolName(req.Name); err != nil {
		response.BadRequest(ctx, "invalid tool name: "+err.Error())
		return
	}

	var existing model.ToolDefinition
	if err := h.db.WithContext(ctx.Request.Context()).
		Where("name = ? AND enabled = ?", req.Name, true).
		First(&existing).Error; err == nil {
		response.Conflict(ctx, fmt.Sprintf("tool with name %q already exists", req.Name))
		return
	}

	tool := model.ToolDefinition{
		Name:          req.Name,
		Description:   req.Description,
		ServiceID:     req.ServiceID,
		Parameters:    req.Parameters,
		InputMapping:  req.InputMapping,
		Enabled:       true,
		OutputMapping: req.OutputMapping,
	}

	result := h.db.WithContext(ctx.Request.Context()).Create(&tool)
	if result.Error != nil {
		response.InternalError(ctx, result.Error.Error())
		return
	}

	response.Created(ctx, gin.H{
		"tool": tool,
	})
}

type UpdateToolRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ServiceID     *uint  `json:"service_id"`
	Parameters    []byte `json:"parameters"`
	InputMapping  []byte `json:"input_mapping"`
	Enabled       *bool  `json:"enabled"`
	OutputMapping []byte `json:"output_mapping"`
}

func (h *ToolHandler) UpdateTool(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid tool id")
		return
	}

	var tool model.ToolDefinition
	result := h.db.WithContext(ctx.Request.Context()).
		Where("id = ?", id).
		First(&tool)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			response.NotFound(ctx, "tool not found")
			return
		}
		response.InternalError(ctx, result.Error.Error())
		return
	}

	var req UpdateToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	if req.Name != "" {
		if err := model.ValidateToolName(req.Name); err != nil {
			response.BadRequest(ctx, "invalid tool name: "+err.Error())
			return
		}
		var existing model.ToolDefinition
		if err := h.db.WithContext(ctx.Request.Context()).
			Where("name = ? AND enabled = ? AND id != ?", req.Name, true, id).
			First(&existing).Error; err == nil {
			response.Conflict(ctx, fmt.Sprintf("tool with name %q already exists", req.Name))
			return
		}
		tool.Name = req.Name
	}
	if req.Description != "" {
		tool.Description = req.Description
	}
	if req.ServiceID != nil {
		tool.ServiceID = *req.ServiceID
	}
	if req.Parameters != nil {
		tool.Parameters = req.Parameters
	}
	if req.InputMapping != nil {
		tool.InputMapping = req.InputMapping
	}
	if req.Enabled != nil {
		tool.Enabled = *req.Enabled
	}
	if req.OutputMapping != nil {
		tool.OutputMapping = req.OutputMapping
	}

	result = h.db.WithContext(ctx.Request.Context()).Save(&tool)
	if result.Error != nil {
		response.InternalError(ctx, result.Error.Error())
		return
	}

	response.Success(ctx, nil)
}

func (h *ToolHandler) DeleteTool(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid tool id")
		return
	}

	result := h.db.WithContext(ctx.Request.Context()).
		Model(&model.ToolDefinition{}).
		Where("id = ?", id).
		Update("enabled", false)

	if result.Error != nil {
		response.InternalError(ctx, result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(ctx, "tool not found")
		return
	}

	response.Success(ctx, nil)
}

func (h *ToolHandler) GetHTTPServiceOutputSchema(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
		return
	}

	service, err := h.serviceStore.Get(ctx.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(ctx, "service not found")
		return
	}

	response.Success(ctx, gin.H{
		"service_id":    service.ID,
		"name":          service.Name,
		"output_schema": service.OutputSchema,
	})
}
