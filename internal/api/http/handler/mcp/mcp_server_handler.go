package mcp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/domain/model"
	domainService "dynamic_mcp_go_server/internal/domain/service"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func generateVAuthKeyFromUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

type MCPServerHandler struct {
	service     *service.MCPServerService
	toolService *service.ToolService
	db          *gorm.DB
	logger      logger.Logger
}

func NewMCPServerHandler(svc *service.MCPServerService, toolSvc *service.ToolService, db *gorm.DB, log logger.Logger) *MCPServerHandler {
	return &MCPServerHandler{
		service:     svc,
		toolService: toolSvc,
		db:          db,
		logger:      log,
	}
}

type ListServersResponse struct {
	Servers   []*ServerWithToolCount `json:"servers"`
	Total     int64                  `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"page_size"`
	TotalPage int64                  `json:"total_page"`
}

type ServerWithToolCount struct {
	*model.MCPServer
	ToolCount int64 `json:"tool_count"`
}

func (h *MCPServerHandler) ListServers(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	name := ctx.Query("name")
	stateStr := ctx.Query("state")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var state *int
	if stateStr != "" {
		s, err := strconv.Atoi(stateStr)
		if err == nil && (s == 0 || s == 1) {
			state = &s
		}
	}

	serversWithCount, total, err := h.service.ListServersWithToolCount(ctx.Request.Context(), page, pageSize, name, state)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	servers := make([]*ServerWithToolCount, len(serversWithCount))
	for i, sc := range serversWithCount {
		servers[i] = &ServerWithToolCount{
			MCPServer: sc.Server,
			ToolCount: sc.ToolCount,
		}
	}

	totalPage := total / int64(pageSize)
	if total%int64(pageSize) != 0 {
		totalPage++
	}

	response.Success(ctx, ListServersResponse{
		Servers:   servers,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	})
}

func (h *MCPServerHandler) GetServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	server, err := h.service.GetServer(ctx.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrMCPServerNotFound {
			response.NotFound(ctx, "mcp server not found")
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"server": server,
	})
}

type CreateServerRequest struct {
	Type           string `json:"type" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	HTTPServerURL  string `json:"http_server_url"`
	AuthHeader     string `json:"auth_header"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	ExtraHeaders   string `json:"extra_headers"`
}

func (h *MCPServerHandler) CreateServer(ctx *gin.Context) {
	var req CreateServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	vauthKey := generateVAuthKeyFromUUID()

	server := model.MCPServer{
		VAuthKey:       vauthKey,
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		HTTPServerURL:  req.HTTPServerURL,
		AuthHeader:     req.AuthHeader,
		TimeoutSeconds: req.TimeoutSeconds,
		ExtraHeaders:   req.ExtraHeaders,
	}

	if err := h.service.CreateServer(ctx.Request.Context(), &server); err != nil {
		if err == service.ErrMCPServerExists {
			response.Conflict(ctx, "mcp server with this vauth_key already exists")
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Created(ctx, gin.H{
		"server_id": server.ID,
		"server":    server,
	})
}

type UpdateServerRequest struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	HTTPServerURL  string `json:"http_server_url"`
	AuthHeader     string `json:"auth_header"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	ExtraHeaders   string `json:"extra_headers"`
}

func (h *MCPServerHandler) UpdateServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	server, err := h.service.GetServer(ctx.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrMCPServerNotFound {
			response.NotFound(ctx, "mcp server not found")
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	var req UpdateServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Description != "" {
		server.Description = req.Description
	}
	if req.Type != "" {
		server.Type = req.Type
	}
	if req.HTTPServerURL != "" {
		server.HTTPServerURL = req.HTTPServerURL
	}
	if req.AuthHeader != "" {
		server.AuthHeader = req.AuthHeader
	}
	if req.TimeoutSeconds > 0 {
		server.TimeoutSeconds = req.TimeoutSeconds
	}
	if req.ExtraHeaders != "" {
		server.ExtraHeaders = req.ExtraHeaders
	}

	if err := h.service.UpdateServer(ctx.Request.Context(), server); err != nil {
		switch err {
		case service.ErrMCPServerNotFound:
			response.NotFound(ctx, "mcp server not found")
		case service.ErrMCPServerExists:
			response.Conflict(ctx, "mcp server with this vauth_key already exists")
		case service.ErrServerTypeCannotBeChanged:
			response.BadRequest(ctx, "server type cannot be changed after creation")
		default:
			response.InternalError(ctx, err.Error())
		}
		return
	}

	response.Success(ctx, gin.H{
		"server_id": server.ID,
	})
}

func (h *MCPServerHandler) DeleteServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	if err := h.service.DeleteServer(ctx.Request.Context(), uint(id)); err != nil {
		if err == service.ErrMCPServerNotFound {
			response.NotFound(ctx, "mcp server not found")
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"server_id": id,
	})
}

func (h *MCPServerHandler) RestoreServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	if err := h.service.RestoreServer(ctx.Request.Context(), uint(id)); err != nil {
		if err == service.ErrMCPServerNotFound {
			response.NotFound(ctx, "mcp server not found")
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"server_id": id,
	})
}

func (h *MCPServerHandler) GetServerTools(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	server, err := h.service.GetServer(ctx.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrMCPServerNotFound {
			response.NotFound(ctx, "mcp server not found")
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	tools, err := h.service.GetServerTools(ctx.Request.Context(), uint(id))
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"server_id": id,
		"vauth_key": server.VAuthKey,
		"tools":     tools,
		"count":     len(tools),
	})
}

type AddToolsToServerRequest struct {
	Tools []model.ToolDefinition `json:"tools" binding:"required"`
}

func (h *MCPServerHandler) AddToolsToServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	var req AddToolsToServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	if len(req.Tools) == 0 {
		response.BadRequest(ctx, "tools array is empty")
		return
	}

	addedCount := 0
	for i := range req.Tools {
		if err := h.service.AddToolToServer(ctx.Request.Context(), uint(id), &req.Tools[i]); err == nil {
			addedCount++
		}
	}

	response.Created(ctx, gin.H{
		"server_id":   id,
		"added_count": addedCount,
		"total_count": len(req.Tools),
	})
}

func (h *MCPServerHandler) RemoveToolFromServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	toolName := ctx.Param("toolName")
	if toolName == "" {
		response.BadRequest(ctx, "tool name is required")
		return
	}

	err = h.service.RemoveToolFromServerByName(ctx.Request.Context(), uint(id), toolName)
	if err != nil {
		switch {
		case err == service.ErrMCPServerNotFound:
			response.NotFound(ctx, "mcp server not found")
		case err.Error() == fmt.Sprintf("tool not found: %s", toolName):
			response.NotFound(ctx, fmt.Sprintf("tool '%s' not found", toolName))
		default:
			response.InternalError(ctx, err.Error())
		}
		return
	}

	response.Success(ctx, gin.H{
		"server_id": id,
		"tool_name": toolName,
	})
}

type CreateToolFromHTTPServiceRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	ServiceID     uint   `json:"service_id" binding:"required"`
	InputExtra    []byte `json:"input_extra"`
	OutputMapping []byte `json:"output_mapping"`
}

func (h *MCPServerHandler) CreateToolFromHTTPService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid server id")
		return
	}

	var req CreateToolFromHTTPServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	tool, err := h.toolService.CreateFromHTTPService(ctx.Request.Context(), domainService.CreateToolFromHTTPServiceCommand{
		Name:          req.Name,
		Description:   req.Description,
		ServerID:      uint(id),
		ServiceID:     req.ServiceID,
		OutputMapping: req.OutputMapping,
	})
	if err != nil {
		switch {
		case err.Error() == "mcp server not found":
			response.NotFound(ctx, "mcp server not found")
		case errors.Is(err, domainService.ErrOnlyHTTPServiceServerCanHaveTools):
			response.BadRequest(ctx, err.Error())
		case errors.Is(err, domainService.ErrToolNameAlreadyExists):
			response.Conflict(ctx, err.Error())
		case errors.Is(err, domainService.ErrHTTPServiceNotFound):
			response.NotFound(ctx, err.Error())
		default:
			response.InternalError(ctx, err.Error())
		}
		return
	}

	response.Created(ctx, gin.H{
		"tool": tool,
	})
}
