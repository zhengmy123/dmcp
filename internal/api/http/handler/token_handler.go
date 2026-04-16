package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dynamic_mcp_go_server/internal/service"
)

// RegisterTokenRoutes 注册 Token 管理 API
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
		tokens := authService.ListTokens(c.Request.Context())
		c.JSON(200, gin.H{
			"tokens": tokens,
			"count":  len(tokens),
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
			c.JSON(400, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		// 自动生成 KeyID、Token 和 Secret
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
			c.JSON(500, gin.H{"error": "failed to register token", "details": err.Error()})
			return
		}

		c.JSON(201, gin.H{
			"message": "token registered",
			"key_id":  req.KeyID,
			"token":   req.Token,
			"secret":  req.Secret,
		})
	}
}

func deleteToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if !authService.DeleteToken(c.Request.Context(), token) {
			c.JSON(404, gin.H{"error": "token not found"})
			return
		}
		c.JSON(200, gin.H{"message": "token deleted", "token": token})
	}
}

func refreshToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		newToken, newSecret, err := authService.RefreshToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message":    "token refreshed",
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
			c.JSON(404, gin.H{"error": "token not found"})
			return
		}
		c.JSON(200, gin.H{"message": "token enabled", "token": token})
	}
}

func disableToken(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if !authService.DisableToken(c.Request.Context(), token) {
			c.JSON(404, gin.H{"error": "token not found"})
			return
		}
		c.JSON(200, gin.H{"message": "token disabled", "token": token})
	}
}

func generateTokenValue() string {
	return uuid.New().String()
}

func generateKeyID() string {
	id := uuid.New().String()
	return "key-" + id[:8]
}
