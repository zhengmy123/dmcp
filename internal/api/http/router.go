package http

import (
	"net/http"

	"dynamic_mcp_go_server/internal/api/http/handler"
	"dynamic_mcp_go_server/internal/api/http/handler/mcp"
	apimw "dynamic_mcp_go_server/internal/api/middleware"
	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/auth"
	"dynamic_mcp_go_server/internal/infrastructure/database"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	e *gin.Engine,
	registry *service.DynamicRegistry,
	groupManager *service.MCPGroupManager,
	authService *service.AuthService,
	httpServiceManager *service.HTTPServiceManager,
	serviceStore repository.ServiceStore,
	gormDB *gorm.DB,
	jwtManager *auth.JWTManager,
	log logger.Logger,
	mcpServerService *service.MCPServerService,
	toolService *service.ToolService,
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

	// MCP Server 管理路由 (需要 JWT 认证)
	if jwtManager != nil && mcpServerService != nil && toolService != nil {
		mcpGroup := e.Group("/api/admin")
		mcpGroup.Use(apimw.JWTAuth(jwtManager))

		mcpHandler := mcp.NewMCPServerHandler(mcpServerService, toolService, gormDB, log)
		mcpGroup.GET("/mcp-servers", mcpHandler.ListServers)
		mcpGroup.GET("/mcp-servers/:id", mcpHandler.GetServer)
		mcpGroup.POST("/mcp-servers", mcpHandler.CreateServer)
		mcpGroup.PUT("/mcp-servers/:id", mcpHandler.UpdateServer)
		mcpGroup.DELETE("/mcp-servers/:id", mcpHandler.DeleteServer)
		mcpGroup.GET("/mcp-servers/:id/tools", mcpHandler.GetServerTools)
		mcpGroup.POST("/mcp-servers/:id/tools", mcpHandler.AddToolsToServer)
		mcpGroup.DELETE("/mcp-servers/:id/tools/:toolName", mcpHandler.RemoveToolFromServer)
		mcpGroup.POST("/mcp-servers/:id/tools/from-http-service", mcpHandler.CreateToolFromHTTPService)

		toolHandler := mcp.NewToolHandler(gormDB, serviceStore, log)
		mcpGroup.GET("/tools", toolHandler.ListTools)
		mcpGroup.GET("/tools/:id", toolHandler.GetTool)
		mcpGroup.POST("/tools", toolHandler.CreateTool)
		mcpGroup.PUT("/tools/:id", toolHandler.UpdateTool)
		mcpGroup.DELETE("/tools/:id", toolHandler.DeleteTool)

		mcpGroup.GET("/http-services/:id/output-schema", toolHandler.GetHTTPServiceOutputSchema)

		toolBindingDAO := database.NewGORMToolServerBindingDAO(gormDB)
		toolStore := database.NewGORMToolStore(gormDB)
		mcpServerDAO := database.NewGORMMCPServerDAO(gormDB)
		toolBindingService := service.NewToolBindingService(toolBindingDAO, toolStore, mcpServerDAO)
		toolBindingHandler := mcp.NewToolBindingHandler(toolBindingService, log)

		mcpGroup.GET("/tool-bindings/:toolId", toolBindingHandler.GetToolBindings)
		mcpGroup.POST("/tool-bindings", toolBindingHandler.BindTool)
		mcpGroup.DELETE("/tool-bindings/:toolId/:serverId", toolBindingHandler.UnbindTool)
		mcpGroup.POST("/tool-bindings/batch-bind", toolBindingHandler.BatchBind)
		mcpGroup.DELETE("/tool-bindings/batch-unbind", toolBindingHandler.BatchUnbind)
		mcpGroup.GET("/server-bindings/:serverId", toolBindingHandler.GetServerBindings)
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
