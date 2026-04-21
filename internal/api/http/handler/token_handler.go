package handler

import (
	"strconv"

	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterTokenRoutes(g *gin.RouterGroup, authService *service.AuthService) {
	g.GET("/tokens", listTokens(authService))
	g.POST("/tokens", createToken(authService))
	g.DELETE("/tokens/:token", deleteToken(authService))
	g.POST("/tokens/:token/refresh", refreshToken(authService))
	g.POST("/tokens/:token/enable", enableToken(authService))
	g.POST("/tokens/:token/disable", disableToken(authService))
}

func listTokens(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}

		pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		if err != nil || pageSize < 1 {
			pageSize = 10
		}

		tokens, total := authService.ListTokensPaginated(c.Request.Context(), page, pageSize)
		response.Success(c, gin.H{
			"items":     tokens,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func createToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			KeyID  string `json:"key_id"`
			Token  string `json:"token"`
			Secret string `json:"secret"`
			Name   string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request: "+err.Error())
			return
		}

		if req.KeyID == "" {
			req.KeyID = generateKeyID()
		}
		if req.Token == "" {
			req.Token = generateTokenValue()
		}
		if req.Secret == "" {
			req.Secret = generateTokenValue() + generateTokenValue()
		}

		if err := authService.RegisterToken(c.Request.Context(), req.KeyID, req.Token, req.Secret, req.Name); err != nil {
			response.InternalError(c, "failed to register token: "+err.Error())
			return
		}

		response.Created(c, gin.H{
			"key_id": req.KeyID,
			"token":  req.Token,
			"secret": req.Secret,
		})
	}
}

func deleteToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if !authService.DeleteToken(c.Request.Context(), token) {
			response.NotFound(c, "token not found")
			return
		}
		response.SuccessWithMessage(c, "token deleted", gin.H{"token": token})
	}
}

func refreshToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		newToken, newSecret, err := authService.RefreshToken(c.Request.Context(), token)
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}
		response.Success(c, gin.H{
			"old_token":  token,
			"new_token":  newToken,
			"new_secret": newSecret,
		})
	}
}

func enableToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if !authService.EnableToken(c.Request.Context(), token) {
			response.NotFound(c, "token not found")
			return
		}
		response.SuccessWithMessage(c, "token enabled", gin.H{"token": token})
	}
}

func disableToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if !authService.DisableToken(c.Request.Context(), token) {
			response.NotFound(c, "token not found")
			return
		}
		response.SuccessWithMessage(c, "token disabled", gin.H{"token": token})
	}
}

func generateTokenValue() string {
	return uuid.New().String()
}

func generateKeyID() string {
	id := uuid.New().String()
	return "key-" + id[:8]
}
