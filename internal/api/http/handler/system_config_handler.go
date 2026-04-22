package handler

import (
	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
)

type SystemConfigHandler struct {
	svc *service.SystemConfigService
}

func NewSystemConfigHandler(svc *service.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{svc: svc}
}

type GetConfigResponse struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
}

type UpdateConfigRequest struct {
	ConfigValue string `json:"config_value" binding:"required"`
}

func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "config key is required")
		return
	}

	config, err := h.svc.GetConfig(c.Request.Context(), key)
	if err != nil {
		response.InternalError(c, "failed to get config: "+err.Error())
		return
	}

	if config == nil {
		response.NotFound(c, "config not found")
		return
	}

	response.Success(c, GetConfigResponse{
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
	})
}

func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "config key is required")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	config, err := h.svc.UpdateConfig(c.Request.Context(), key, req.ConfigValue)
	if err != nil {
		response.InternalError(c, "failed to update config: "+err.Error())
		return
	}

	response.Success(c, GetConfigResponse{
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
	})
}
