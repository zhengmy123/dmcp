package http

import (
	"net/http"

	"dynamic_mcp_go_server/internal/api/http/handler"
	"dynamic_mcp_go_server/internal/api/http/handler/mcp"
	apimw "dynamic_mcp_go_server/internal/api/middleware"
	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/auth"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	e *gin.Engine,
	registry *service.DynamicRegistry,
	groupManager *service.MCPGroupManager,
	proxyHandler *service.ProxyHandler,
	authService *service.AuthService,
	httpServiceManager *service.HTTPServiceManager,
	serviceStore repository.ServiceStore,
	mcpServerStore repository.MCPServerStore,
	toolStore repository.ToolStore,
	toolBindingStore repository.ToolServerBindingStore,
	serverBuildInfoStore repository.ServerBuildInfoStore,
	jwtManager *auth.JWTManager,
	log logger.Logger,
	gormDB interface{},
	systemConfigStore repository.SystemConfigStore,
) {
	metadataHandler := service.NewHTTPHandler(registry)
	scopedMCPHandler := service.NewScopedMCPHandler(groupManager, proxyHandler)

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

	// MCP Server 管理路由 (需要 JWT 认证)
	if jwtManager != nil && mcpServerStore != nil {
		mcpGroup := e.Group("/api/admin")
		mcpGroup.Use(apimw.JWTAuth(jwtManager))

		mcpHandler := mcp.NewMCPServerHandler(mcpServerStore, toolStore, toolBindingStore, serverBuildInfoStore, serviceStore, proxyHandler, log)
		mcpGroup.GET("/mcp-servers", mcpHandler.ListServers)
		mcpGroup.GET("/mcp-servers/:id", mcpHandler.GetServer)
		mcpGroup.POST("/mcp-servers", mcpHandler.CreateServer)
		mcpGroup.PUT("/mcp-servers/:id", mcpHandler.UpdateServer)
		mcpGroup.DELETE("/mcp-servers/:id", mcpHandler.DeleteServer)
		mcpGroup.POST("/mcp-servers/:id/restore", mcpHandler.RestoreServer)
		mcpGroup.GET("/mcp-servers/:id/tools", mcpHandler.GetServerTools)
		mcpGroup.POST("/mcp-servers/:id/tools", mcpHandler.AddToolsToServer)
		mcpGroup.DELETE("/mcp-servers/:id/tools/:toolName", mcpHandler.RemoveToolFromServer)
		mcpGroup.POST("/mcp-servers/:id/tools/from-http-service", mcpHandler.CreateToolFromHTTPService)
		mcpGroup.POST("/mcp-servers/:id/sync-build", mcpHandler.SyncBuild)

		toolHandler := mcp.NewToolHandler(toolStore, toolBindingStore, serviceStore, log)
		mcpGroup.GET("/tools", toolHandler.ListTools)
		mcpGroup.GET("/tools/:id", toolHandler.GetTool)
		mcpGroup.POST("/tools", toolHandler.CreateTool)
		mcpGroup.PUT("/tools/:id", toolHandler.UpdateTool)
		mcpGroup.DELETE("/tools/:id", toolHandler.DeleteTool)

		mcpGroup.GET("/http-services/:id/output-schema", toolHandler.GetHTTPServiceOutputSchema)
		mcpGroup.GET("/http-services/:id/tools", toolHandler.GetHTTPServiceTools)

		var db *gorm.DB
		if gormDB != nil {
			db = gormDB.(*gorm.DB)
		}
		toolBindingHandler := mcp.NewToolBindingHandler(toolBindingStore, mcpServerStore, toolStore, serverBuildInfoStore, serviceStore, log, db)
		mcpGroup.GET("/tool-bindings/:toolId", toolBindingHandler.GetToolBindings)
		mcpGroup.POST("/tool-bindings", toolBindingHandler.BindTool)
		mcpGroup.DELETE("/tool-bindings/:toolId/:serverId", toolBindingHandler.UnbindTool)
		mcpGroup.POST("/tool-bindings/batch-bind", toolBindingHandler.BatchBind)
		mcpGroup.DELETE("/tool-bindings/batch-unbind", toolBindingHandler.BatchUnbind)
		mcpGroup.GET("/server-bindings/:serverId", toolBindingHandler.GetServerBindings)
	}

	if jwtManager != nil && systemConfigStore != nil {
		systemConfigService := service.NewSystemConfigService(systemConfigStore)
		systemConfigHandler := handler.NewSystemConfigHandler(systemConfigService)
		systemGroup := e.Group("/api/v1/system")
		systemGroup.Use(apimw.JWTAuth(jwtManager))
		systemGroup.GET("/config/:key", systemConfigHandler.GetConfig)
		systemGroup.PUT("/config/:key", systemConfigHandler.UpdateConfig)
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
