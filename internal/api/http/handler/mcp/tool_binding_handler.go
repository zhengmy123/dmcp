package mcp

import (
	"strconv"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ToolBindingHandler struct {
	service *service.ToolBindingService
	logger  logger.Logger
}

func NewToolBindingHandler(
	toolBindingStore repository.ToolServerBindingStore,
	mcpServerStore repository.MCPServerStore,
	toolStore repository.ToolStore,
	serverBuildInfoStore repository.ServerBuildInfoStore,
	serviceStore repository.ServiceStore,
	log logger.Logger,
	db *gorm.DB,
) *ToolBindingHandler {
	serverBuildService := service.NewServerBuildService(mcpServerStore, toolStore, toolBindingStore, serverBuildInfoStore, serviceStore, nil)
	svc := service.NewToolBindingService(toolBindingStore, toolStore, mcpServerStore, serverBuildService, db)

	return &ToolBindingHandler{
		service: svc,
		logger:  log,
	}
}

type BindRequest struct {
	ToolID   uint `json:"tool_id"`
	ServerID uint `json:"server_id"`
}

type BatchBindRequest struct {
	ToolIDs   []uint `json:"tool_ids"`
	ServerIDs []uint `json:"server_ids"`
}

type BatchUnbindRequest struct {
	BindingIDs []uint `json:"binding_ids"`
}

func (h *ToolBindingHandler) BindTool(ctx *gin.Context) {
	var req BindRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	binding, err := h.service.BindTool(ctx.Request.Context(), service.BindToolRequest{
		ToolID:   req.ToolID,
		ServerID: req.ServerID,
	})
	if err != nil {
		switch err {
		case service.ErrToolNotFound:
			response.NotFound(ctx, "tool not found")
		case service.ErrServerNotFound:
			response.NotFound(ctx, "server not found")
		case service.ErrBindingExists:
			response.Conflict(ctx, "binding already exists")
		default:
			response.InternalError(ctx, err.Error())
		}
		return
	}

	response.Created(ctx, gin.H{
		"binding": binding,
	})
}

func (h *ToolBindingHandler) UnbindTool(ctx *gin.Context) {
	toolIDParam := ctx.Param("toolId")
	toolID, err := strconv.ParseUint(toolIDParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid tool id")
		return
	}

	serverIDParam := ctx.Param("serverId")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	err = h.service.UnbindTool(ctx.Request.Context(), service.BindToolRequest{
		ToolID:   uint(toolID),
		ServerID: uint(serverID),
	})
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, nil)
}

func (h *ToolBindingHandler) BatchBind(ctx *gin.Context) {
	var req BatchBindRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	count, err := h.service.BatchBindTools(ctx.Request.Context(), service.BatchBindRequest{
		ToolIDs:   req.ToolIDs,
		ServerIDs: req.ServerIDs,
	})
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Created(ctx, gin.H{
		"count": count,
	})
}

func (h *ToolBindingHandler) BatchUnbind(ctx *gin.Context) {
	var req BatchUnbindRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	count, err := h.service.BatchUnbindTools(ctx.Request.Context(), req.BindingIDs)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"count": count,
	})
}

func (h *ToolBindingHandler) GetToolBindings(ctx *gin.Context) {
	toolIDParam := ctx.Param("toolId")
	toolID, err := strconv.ParseUint(toolIDParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid tool id")
		return
	}

	bindings, err := h.service.GetToolBindings(ctx.Request.Context(), uint(toolID))
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"bindings": bindings,
		"count":    len(bindings),
	})
}

func (h *ToolBindingHandler) GetServerBindings(ctx *gin.Context) {
	serverIDParam := ctx.Param("serverId")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	bindings, err := h.service.GetServerBindings(ctx.Request.Context(), uint(serverID))
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"bindings": bindings,
		"count":    len(bindings),
	})
}
