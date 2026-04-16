package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	domainService "dynamic_mcp_go_server/internal/domain/service"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// generateVAuthKeyFromUUID 使用 UUID 生成 VAuthKey（去掉横线）
func generateVAuthKeyFromUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// MCPServerHandler MCPServer 管理的 HTTP Handler
type MCPServerHandler struct {
	service     *service.MCPServerService
	toolService *service.ToolService
	db          *gorm.DB
	logger      logger.Logger
}

// NewMCPServerHandler 创建 MCPServerHandler
func NewMCPServerHandler(svc *service.MCPServerService, toolSvc *service.ToolService, db *gorm.DB, log logger.Logger) *MCPServerHandler {
	return &MCPServerHandler{
		service:     svc,
		toolService: toolSvc,
		db:          db,
		logger:      log,
	}
}

// ListServers GET /api/admin/mcp-servers - 列表
func (h *MCPServerHandler) ListServers(ctx *gin.Context) {
	servers, err := h.service.ListServers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list mcp servers",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"servers": servers,
		"count":   len(servers),
	})
}

// GetServer GET /api/admin/mcp-servers/:id - 详情
func (h *MCPServerHandler) GetServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid server id",
			"id":    idParam,
		})
		return
	}

	server, err := h.service.GetServer(ctx.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrMCPServerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "mcp server not found",
				"id":    idParam,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get mcp server",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server": server,
	})
}

// CreateServerRequest 创建 MCPServer 请求
type CreateServerRequest struct {
	Type           string `json:"type" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	Enabled        *bool  `json:"enabled"`
	HTTPServerURL  string `json:"http_server_url"`
	AuthHeader     string `json:"auth_header"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	ExtraHeaders   string `json:"extra_headers"`
}

// CreateServer POST /api/admin/mcp-servers - 创建
func (h *MCPServerHandler) CreateServer(ctx *gin.Context) {
	var req CreateServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
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
		Enabled:        true,
	}
	if req.Enabled != nil {
		server.Enabled = *req.Enabled
	}

	if err := h.service.CreateServer(ctx.Request.Context(), &server); err != nil {
		if err == service.ErrMCPServerExists {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "mcp server with this vauth_key already exists",
				"details": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create mcp server",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":   "mcp server created successfully",
		"server_id": server.ID,
		"server":    server,
	})
}

// UpdateServerRequest 更新 MCPServer 请求
type UpdateServerRequest struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Enabled        *bool  `json:"enabled"`
	HTTPServerURL  string `json:"http_server_url"`
	AuthHeader     string `json:"auth_header"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	ExtraHeaders   string `json:"extra_headers"`
}

// UpdateServer PUT /api/admin/mcp-servers/:id - 更新
func (h *MCPServerHandler) UpdateServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid server id",
			"id":    idParam,
		})
		return
	}

	server, err := h.service.GetServer(ctx.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrMCPServerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "mcp server not found",
				"id":    idParam,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get mcp server",
			"details": err.Error(),
		})
		return
	}

	var req UpdateServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Description != "" {
		server.Description = req.Description
	}
	if req.Enabled != nil {
		server.Enabled = *req.Enabled
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
		if err == service.ErrMCPServerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "mcp server not found",
				"id":    idParam,
			})
			return
		}
		if err == service.ErrMCPServerExists {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "mcp server with this vauth_key already exists",
				"details": err.Error(),
			})
			return
		}
		if err == service.ErrServerTypeCannotBeChanged {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "server type cannot be changed after creation",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update mcp server",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "mcp server updated successfully",
		"server_id": server.ID,
	})
}

// DeleteServer DELETE /api/admin/mcp-servers/:id - 删除
func (h *MCPServerHandler) DeleteServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid server id",
			"id":    idParam,
		})
		return
	}

	if err := h.service.DeleteServer(ctx.Request.Context(), uint(id)); err != nil {
		if err == service.ErrMCPServerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "mcp server not found",
				"id":    idParam,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete mcp server",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "mcp server deleted successfully",
		"server_id": id,
	})
}

// GetServerTools GET /api/admin/mcp-servers/:id/tools - 获取 Server 下的工具
func (h *MCPServerHandler) GetServerTools(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid server id",
			"id":    idParam,
		})
		return
	}

	// 获取 server
	server, err := h.service.GetServer(ctx.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrMCPServerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "mcp server not found",
				"id":    idParam,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get mcp server",
			"details": err.Error(),
		})
		return
	}

	// 获取该 server 下的所有工具
	tools, err := h.service.GetServerTools(ctx.Request.Context(), uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list tools",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_id": id,
		"vauth_key": server.VAuthKey,
		"tools":     tools,
		"count":     len(tools),
	})
}

// AddToolsToServerRequest 添加工具到 Server 请求
type AddToolsToServerRequest struct {
	Tools []model.ToolDefinition `json:"tools" binding:"required"`
}

// AddToolsToServer POST /api/admin/mcp-servers/:id/tools - 向 Server 添加工具
func (h *MCPServerHandler) AddToolsToServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid server id",
			"id":    idParam,
		})
		return
	}

	var req AddToolsToServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if len(req.Tools) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "tools array is empty",
		})
		return
	}

	// 设置 vauth_key 并保存工具
	addedCount := 0
	for i := range req.Tools {
		if err := h.service.AddToolToServer(ctx.Request.Context(), uint(id), &req.Tools[i]); err == nil {
			addedCount++
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":     "tools added successfully",
		"server_id":   id,
		"added_count": addedCount,
		"total_count": len(req.Tools),
	})
}

// RemoveToolFromServer DELETE /api/admin/mcp-servers/:id/tools/:toolName - 从 Server 移除工具
func (h *MCPServerHandler) RemoveToolFromServer(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid server id",
			"id":    idParam,
		})
		return
	}

	toolName := ctx.Param("toolName")
	if toolName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "tool name is required",
		})
		return
	}

	err = h.service.RemoveToolFromServerByName(ctx.Request.Context(), uint(id), toolName)
	if err != nil {
		if err == service.ErrMCPServerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "mcp server not found",
				"id":    idParam,
			})
			return
		}
		if err.Error() == fmt.Sprintf("tool not found: %s", toolName) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":     "tool not found",
				"server_id": id,
				"tool_name": toolName,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to remove tool",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "tool removed successfully",
		"server_id": id,
		"tool_name": toolName,
	})
}

// CreateToolFromHTTPServiceRequest 从 HTTP Service 创建工具请求
type CreateToolFromHTTPServiceRequest struct {
	Name          string          `json:"name" binding:"required"`
	Description   string          `json:"description"`
	ServiceID     uint            `json:"service_id" binding:"required"`
	InputExtra    json.RawMessage `json:"input_extra"`
	OutputMapping json.RawMessage `json:"output_mapping"`
}

// CreateToolFromHTTPService POST /api/admin/mcp-servers/:id/tools/from-http-service
func (h *MCPServerHandler) CreateToolFromHTTPService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid server id",
			"id":    idParam,
		})
		return
	}

	var req CreateToolFromHTTPServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	tool, err := h.toolService.CreateFromHTTPService(ctx.Request.Context(), domainService.CreateToolFromHTTPServiceCommand{
		Name:          req.Name,
		Description:   req.Description,
		ServerID:      uint(id),
		ServiceID:     req.ServiceID,
		InputExtra:    req.InputExtra,
		OutputMapping: req.OutputMapping,
	})
	if err != nil {
		switch {
		case err.Error() == "mcp server not found":
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "mcp server not found",
			})
		case errors.Is(err, domainService.ErrOnlyHTTPServiceServerCanHaveTools):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, domainService.ErrToolNameAlreadyExists):
			ctx.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, domainService.ErrHTTPServiceNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to create tool",
				"details": err.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "tool created successfully",
		"tool":    tool,
	})
}
