package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAPI "dynamic_mcp_go_server/internal/api/http"
	"dynamic_mcp_go_server/internal/common/logger"
	commonmw "dynamic_mcp_go_server/internal/common/middleware"
	"dynamic_mcp_go_server/internal/config"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/auth"
	"dynamic_mcp_go_server/internal/infrastructure/database"
	"dynamic_mcp_go_server/internal/infrastructure/store/httpservice"
	"dynamic_mcp_go_server/internal/infrastructure/store/tooldef"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	appLogger, loggerCleanup, err := logger.NewFileLogger("logs", "server.log")
	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
	defer func() { _ = loggerCleanup() }()

	store, cleanup, err := buildStore(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("build store failed", logger.Error(err))
	}
	defer cleanup()

	authService := service.NewAuthService(cfg.AdminToken)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiration)*time.Hour)

	gormDB, gormCleanup := initDatabase(cfg, appLogger, authService)
	if gormCleanup != nil {
		defer gormCleanup()
	}

	userService, userServiceCleanup := buildUserServiceWithDB(gormDB, appLogger)
	if userServiceCleanup != nil {
		defer userServiceCleanup()
	}

	mcpServer := server.NewMCPServer(
		cfg.ServerName,
		cfg.ServerVersion,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	groupMCP := service.NewMCPGroupManager(cfg.ServerName, cfg.ServerVersion, authService)
	httpServiceManager := service.NewHTTPServiceManager(appLogger)

	serviceStore, storeCleanup := buildServiceStore(appLogger)
	if storeCleanup != nil {
		defer storeCleanup()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if serviceStore != nil {
		if err := syncServicesFromStore(ctx, serviceStore, httpServiceManager, appLogger); err != nil {
			appLogger.Error("initial HTTP service sync failed", logger.Error(err))
		}
		go startServiceManagerSync(ctx, serviceStore, httpServiceManager, appLogger)
	}

	var serverBuildService *service.ServerBuildService
	var registry *service.DynamicRegistry
	if gormDB != nil {
		mcpServerDAO := database.NewGORMMCPServerDAO(gormDB)
		toolStore := database.NewGORMToolStore(gormDB)
		toolServerBindingStore := database.NewGORMToolServerBindingDAO(gormDB)
		serverBuildInfoDAO := database.NewGORMServerBuildInfoDAO(gormDB)
		serverBuildService = service.NewServerBuildService(mcpServerDAO, toolStore, toolServerBindingStore, serverBuildInfoDAO, serviceStore)

		registry = service.NewDynamicRegistry(mcpServer, store, cfg.RefreshInterval(), appLogger, groupMCP, mcpServerDAO, serverBuildService)
	} else {
		registry = service.NewDynamicRegistry(mcpServer, store, cfg.RefreshInterval(), appLogger, groupMCP, nil, nil)
	}

	if err := registry.SyncOnce(ctx); err != nil {
		appLogger.Fatal("initial sync failed", logger.Error(err))
	}
	go registry.Start(ctx)

	appLogger.Info("MCP server started",
		logger.String("store", string(cfg.Store)),
		logger.String("mode", "streamable_http"),
		logger.String("http_addr", cfg.HTTPAddr),
		logger.String("mcp_endpoint", "/mcp/{vauth_key}"),
		logger.String("tool_routes", "/mcp/{vauth_key}/{tool_name}"),
	)

	startHTTPServer(ctx, cfg, registry, groupMCP, authService, httpServiceManager, serviceStore, gormDB, userService, jwtManager, appLogger)
}

func buildStore(cfg config.Config, log logger.Logger) (tooldef.Store, func(), error) {
	if cfg.MySQLDSN == "" {
		return nil, nil, errors.New("MYSQL_DSN is required")
	}

	gormDB, err := database.NewGORMDB(cfg.MySQLDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("connect MySQL failed: %w", err)
	}

	enhancedStore := tooldef.NewEnhancedMySQLStore(gormDB, cfg.MySQLTable, log)
	log.Info("MySQL store initialized")
	return enhancedStore, func() {
		sqlDB, _ := gormDB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}, nil
}

func initDatabase(cfg config.Config, log logger.Logger, authService *service.AuthService) (*gorm.DB, func()) {
	if cfg.MySQLDSN == "" {
		log.Warn("MySQL DSN not configured")
		return nil, nil
	}

	gormDB, err := database.NewGORMDB(cfg.MySQLDSN)
	if err != nil {
		log.Warn("connect MySQL failed", logger.Error(err))
		return nil, nil
	}

	authService.InitWithGORM(gormDB, "mcp_auth_keys")
	authService.StartTokenRefresher(context.Background())
	log.Info("database initialized")

	return gormDB, func() {
		sqlDB, _ := gormDB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}

func buildUserServiceWithDB(gormDB *gorm.DB, log logger.Logger) (*service.UserService, func()) {
	if gormDB == nil {
		return nil, nil
	}

	userService := service.NewUserService(gormDB, "mcp_users")
	log.Info("user service initialized")
	return userService, nil
}

func sampleDefinitions() []tooldef.ToolDefinition {
	min := 1.0
	max := 100.0

	searchUsersParams, _ := json.Marshal([]tooldef.ParameterDefinition{
		{
			Name:        "query",
			Type:        tooldef.ParameterTypeString,
			Required:    true,
			Description: "Search keyword.",
		},
		{
			Name:        "limit",
			Type:        tooldef.ParameterTypeInteger,
			Required:    false,
			Description: "Max rows to return.",
			Default:     10,
			Minimum:     &min,
			Maximum:     &max,
		},
	})

	setUserFlagParams, _ := json.Marshal([]tooldef.ParameterDefinition{
		{
			Name:        "user_id",
			Type:        tooldef.ParameterTypeString,
			Required:    true,
			Description: "User identifier.",
		},
		{
			Name:        "enabled",
			Type:        tooldef.ParameterTypeBoolean,
			Required:    true,
			Description: "Flag value.",
		},
	})

	getOrderSummaryParams, _ := json.Marshal([]tooldef.ParameterDefinition{
		{
			Name:        "order_id",
			Type:        tooldef.ParameterTypeString,
			Required:    true,
			Description: "Order id.",
		},
		{
			Name:     "include_items",
			Type:     tooldef.ParameterTypeBoolean,
			Required: false,
			Default:  true,
		},
	})

	return []tooldef.ToolDefinition{
		{
			Name:        "search_users",
			Description: "Search users by keyword.",
			Enabled:     true,
			Parameters:  searchUsersParams,
		},
		{
			Name:        "set_user_flag",
			Description: "Set feature flag for a user.",
			Enabled:     true,
			Parameters:  setUserFlagParams,
		},
		{
			Name:        "get_order_summary",
			Description: "Get order summary by order id.",
			Enabled:     true,
			Parameters:  getOrderSummaryParams,
		},
	}
}

func buildServiceStore(log logger.Logger) (repository.ServiceStore, func()) {
	storeCfg := config.LoadHTTPServiceStoreConfig()

	switch storeCfg.StoreType {
	case "mysql":
		if storeCfg.MySQLDSN == "" {
			log.Warn("MySQL DSN not configured for HTTP service store")
			return nil, nil
		}
		gormDB, err := database.NewGORMDB(storeCfg.MySQLDSN)
		if err != nil {
			log.Error("open GORM for HTTP service store failed", logger.Error(err))
			return nil, nil
		}
		store := httpservice.NewGORMServiceStore(gormDB, log)
		cleanup := func() {
			sqlDB, _ := gormDB.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
		return store, cleanup
	default:
		store := httpservice.NewMemoryServiceStore(log)
		return store, func() {}
	}
}

func startServiceManagerSync(ctx context.Context, store repository.ServiceStore, manager *service.HTTPServiceManager, log logger.Logger) {
	storeCfg := config.LoadHTTPServiceStoreConfig()

	ticker := time.NewTicker(time.Duration(storeCfg.SyncInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("HTTP service sync stopped")
			return
		case <-ticker.C:
			if err := syncServicesFromStore(ctx, store, manager, log); err != nil {
				log.Error("HTTP service sync failed", logger.Error(err))
			}
		}
	}
}

func syncServicesFromStore(ctx context.Context, store repository.ServiceStore, manager *service.HTTPServiceManager, log logger.Logger) error {
	services, err := store.List(ctx)
	if err != nil {
		return fmt.Errorf("list services from store failed: %w", err)
	}

	storeServiceMap := make(map[uint]*model.HTTPService)
	for _, svc := range services {
		storeServiceMap[svc.ID] = svc
	}

	currentServices := manager.ListServices()

	for _, svc := range services {
		if err := manager.RegisterService(svc); err != nil {
			log.Warn("register service failed", logger.Error(err), logger.Uint("service_id", svc.ID))
		}
	}

	for _, svc := range currentServices {
		if _, exists := storeServiceMap[svc.ID]; !exists {
			if err := manager.DeleteService(svc.ID); err != nil {
				log.Warn("delete service failed", logger.Error(err), logger.Uint("service_id", svc.ID))
			}
		}
	}

	log.Debug("HTTP services synced", logger.Int("count", len(services)))
	return nil
}

func startHTTPServer(
	ctx context.Context,
	cfg config.Config,
	registry *service.DynamicRegistry,
	groupMCP *service.MCPGroupManager,
	authService *service.AuthService,
	httpServiceManager *service.HTTPServiceManager,
	serviceStore repository.ServiceStore,
	gormDB *gorm.DB,
	userService *service.UserService,
	jwtManager *auth.JWTManager,
	log logger.Logger,
) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(commonmw.Recovery(log))
	engine.Use(commonmw.RequestID())
	engine.Use(commonmw.Trace(log))
	engine.Use(commonmw.Cors())
	engine.Use(commonmw.Logger(log))

	engine.RedirectTrailingSlash = false

	httpAPI.RegisterRoutes(engine, registry, groupMCP, authService, httpServiceManager, serviceStore, gormDB, jwtManager, log)

	if userService != nil && jwtManager != nil {
		httpAPI.RegisterUserAuthRoutes(engine, userService, jwtManager)
	}

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: engine,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	log.Info("HTTP server starting", logger.String("addr", cfg.HTTPAddr))

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("HTTP server stopped with error", logger.Error(err))
	}
}

var _ = json.Marshal
