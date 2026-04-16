package http

import (
	"net/http"

	"dynamic_mcp_go_server/internal/api/http/handler"
	apimw "dynamic_mcp_go_server/internal/api/middleware"
	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/auth"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	e *gin.Engine,
	registry *service.DynamicRegistry,
	groupManager *service.MCPGroupManager,
	authService *service.AuthService,
	httpServiceManager *service.HTTPServiceManager,
	serviceStore repository.ServiceStore,
	jwtManager *auth.JWTManager,
	log logger.Logger,
) {
	metadataHandler := service.NewHTTPHandler(registry)
	scopedMCPHandler := service.NewScopedMCPHandler(groupManager)

	meta := gin.WrapH(metadataHandler)

	var mcpHandler http.Handler = scopedMCPHandler
	if authService != nil {
		mcpHandler = apimw.TokenAuthMiddleware(authService.ValidateToken)(scopedMCPHandler)
	}
	mcpGin := gin.WrapH(mcpHandler)

	e.GET(service.RootPath+"/:vauth_key/:tool_name", meta)
	e.GET(service.RootPath, meta)
	e.GET(service.RootPath+"/", meta)
	e.Any(service.RootPath+"/:vauth_key", mcpGin)

	registerPublicAuthRoutes(e)

	if httpServiceManager != nil && jwtManager != nil {
		serviceController := handler.NewController(serviceStore, httpServiceManager, log)

		apiGroup := e.Group("/api/v1")
		apiGroup.Use(apimw.JWTAuth(jwtManager))
		serviceController.RegisterRoutes(apiGroup)

		e.POST("/webhook/:id", apimw.JWTAuth(jwtManager), serviceController.WebhookHandler)
	}

	if authService != nil && jwtManager != nil {
		authGroup := e.Group("/api/v1/auth")
		authGroup.Use(apimw.JWTAuth(jwtManager))
		handler.RegisterTokenRoutes(authGroup, authService)
	}
}

// RegisterUserAuthRoutes 注册用户认证路由
func RegisterUserAuthRoutes(e *gin.Engine, userService *service.UserService, jwtManager *auth.JWTManager) {
	e.POST("/auth/login", handler.LoginHandler(userService, jwtManager))

	jwtAuth := e.Group("/auth")
	jwtAuth.Use(apimw.JWTAuth(jwtManager))
	{
		jwtAuth.GET("/me", handler.GetCurrentUserHandler(userService))
		jwtAuth.POST("/change-password", handler.ChangePasswordHandler(userService))
	}

	adminAuth := e.Group("/api/v1/users")
	adminAuth.Use(apimw.JWTAuth(jwtManager), apimw.AdminRequired())
	{
		adminAuth.GET("", handler.ListUsersHandler(userService))
		adminAuth.POST("", handler.CreateUserHandler(userService))
		adminAuth.PUT("/:id", handler.UpdateUserHandler(userService))
		adminAuth.DELETE("/:id", handler.DeleteUserHandler(userService))
	}
}

func registerPublicAuthRoutes(e *gin.Engine) {
	// 登录路由占位 - 实际路由在 RegisterUserAuthRoutes 中注册
}
