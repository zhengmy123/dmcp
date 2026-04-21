package mcp

import (
	"fmt"
	"strconv"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/store/tooldef"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

type ToolHandler struct {
	toolStore    repository.ToolStore
	bindingStore repository.ToolServerBindingStore
	serviceStore repository.ServiceStore
	logger       logger.Logger
}

func NewToolHandler(
	toolStore repository.ToolStore,
	bindingStore repository.ToolServerBindingStore,
	serviceStore repository.ServiceStore,
	log logger.Logger,
) *ToolHandler {
	return &ToolHandler{
		toolStore:    toolStore,
		bindingStore: bindingStore,
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
	ID            uint                        `json:"id"`
	Name          string                      `json:"name"`
	Description   string                      `json:"description"`
	ServiceID     uint                        `json:"service_id"`
	Parameters    []model.ParameterDefinition `json:"parameters"`
	InputMapping  []model.InputMappingField   `json:"input_mapping"`
	OutputMapping []model.OutputMappingField  `json:"output_mapping"`
	State         int                         `json:"state"`
	CreatedAt     string                      `json:"created_at"`
	UpdatedAt     string                      `json:"updated_at"`
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

	query := &repository.ToolQuery{
		Keyword: &keyword,
		State:   func() *int { v := 1; return &v }(),
	}

	tools, total, err := h.toolStore.List(ctx.Request.Context(), query, page, pageSize)
	if err != nil {
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
			State:         t.State,
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

	tool, err := h.toolStore.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(ctx, "tool not found")
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
			"state":          tool.State,
			"created_at":     tool.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":     tool.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

type CreateToolRequest struct {
	Name          string                     `json:"name" binding:"required"`
	Description   string                     `json:"description"`
	ServiceID     uint                       `json:"service_id" binding:"required"`
	Parameters    []ToolParameterInput       `json:"parameters"`
	InputMapping  []model.InputMappingField  `json:"input_mapping"`
	OutputMapping []model.OutputMappingField `json:"output_mapping"`
}

type ToolParameterInput struct {
	Name           string `json:"name"`
	OriginalName   string `json:"original_name"`
	Type           string `json:"type"`
	Description    string `json:"description"`
	Required       bool   `json:"required"`
	SchemaRequired bool   `json:"schema_required"`
}

func (h *ToolHandler) CreateTool(ctx *gin.Context) {
	var req CreateToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body", err.Error())
		return
	}

	if err := tooldef.ValidateToolName(req.Name); err != nil {
		response.BadRequest(ctx, "invalid tool name", err.Error())
		return
	}

	existing, _ := h.toolStore.GetByName(ctx.Request.Context(), req.Name)
	if existing != nil && existing.State == 1 {
		response.Conflict(ctx, fmt.Sprintf("tool with name %q already exists", req.Name))
		return
	}

	paramsBytes, err := sonic.Marshal(req.Parameters)
	if err != nil {
		response.BadRequest(ctx, "invalid parameters format")
		return
	}
	inputMappingBytes, err := sonic.Marshal(req.InputMapping)
	if err != nil {
		response.BadRequest(ctx, "invalid input_mapping format")
		return
	}
	outputMappingBytes, err := sonic.Marshal(req.OutputMapping)
	if err != nil {
		response.BadRequest(ctx, "invalid output_mapping format")
		return
	}

	tool := &model.ToolDefinition{
		Name:          req.Name,
		Description:   req.Description,
		ServiceID:     req.ServiceID,
		Parameters:    paramsBytes,
		InputMapping:  inputMappingBytes,
		State:         1,
		OutputMapping: outputMappingBytes,
	}

	if err := h.toolStore.Create(ctx.Request.Context(), tool); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Created(ctx, gin.H{
		"tool": tool,
	})
}

type UpdateToolRequest struct {
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	ServiceID     *uint                      `json:"service_id"`
	Parameters    []ToolParameterInput       `json:"parameters"`
	InputMapping  []model.InputMappingField  `json:"input_mapping"`
	OutputMapping []model.OutputMappingField `json:"output_mapping"`
	State         *int                       `json:"state"`
}

func (h *ToolHandler) UpdateTool(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid tool id")
		return
	}

	tool, err := h.toolStore.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(ctx, "tool not found")
		return
	}

	var req UpdateToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body", err.Error())
		return
	}

	if req.Name != "" {
		if err := tooldef.ValidateToolName(req.Name); err != nil {
			response.BadRequest(ctx, "invalid tool name", err.Error())
			return
		}
		existing, _ := h.toolStore.GetByName(ctx.Request.Context(), req.Name)
		if existing != nil && existing.ID != tool.ID && existing.State == 1 {
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
		paramsBytes, err := sonic.Marshal(req.Parameters)
		if err != nil {
			response.BadRequest(ctx, "invalid parameters format")
			return
		}
		tool.Parameters = paramsBytes
	}
	if req.InputMapping != nil {
		inputMappingBytes, err := sonic.Marshal(req.InputMapping)
		if err != nil {
			response.BadRequest(ctx, "invalid input_mapping format")
			return
		}
		tool.InputMapping = inputMappingBytes
	}
	if req.State != nil {
		tool.State = *req.State
	}
	if req.OutputMapping != nil {
		outputMappingBytes, err := sonic.Marshal(req.OutputMapping)
		if err != nil {
			response.BadRequest(ctx, "invalid output_mapping format")
			return
		}
		tool.OutputMapping = outputMappingBytes
	}

	if err := h.toolStore.Update(ctx.Request.Context(), tool); err != nil {
		response.InternalError(ctx, err.Error())
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

	bindings, err := h.bindingStore.ListByToolID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	if len(bindings) > 0 {
		response.Conflict(ctx, "tool has active binding, unbind first")
		return
	}

	if err := h.toolStore.Delete(ctx.Request.Context(), uint(id)); err != nil {
		response.InternalError(ctx, err.Error())
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
