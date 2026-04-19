package main

import (
	"context"
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

type ServerComponents struct {
	MCPServer        *server.MCPServer
	Registry         *service.DynamicRegistry
	GroupMCP         *service.MCPGroupManager
	AuthService      *service.AuthService
	JWTManager       *auth.JWTManager
	UserService      *service.UserService
	HTTPServiceMgr   *service.HTTPServiceManager
	ServiceStore     repository.ServiceStore
	MCPServerStore   repository.MCPServerStore
	ToolStore        repository.ToolStore
	ToolBindingStore repository.ToolServerBindingStore
	BuildInfoStore   repository.ServerBuildInfoStore
	BuildService     *service.ServerBuildService
	ToolDefStore     tooldef.Store
	GORMDB           *gorm.DB
}

func initComponents(cfg config.Config) (*ServerComponents, func(), error) {
	appLogger, loggerCleanup, err := logger.NewFileLogger("logs", "server.log")
	if err != nil {
		return nil, nil, fmt.Errorf("init logger failed: %w", err)
	}

	comp := &ServerComponents{}
	cleanup := func() { loggerCleanup() }

	comp.AuthService = service.NewAuthService(cfg.AdminToken)
	comp.JWTManager = auth.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiration)*time.Hour)

	if cfg.MySQLDSN != "" {
		gormDB, err := database.NewGORMDB(cfg.MySQLDSN)
		if err != nil {
			appLogger.Warn("connect MySQL failed", logger.Error(err))
		} else {
			comp.GORMDB = gormDB
			comp.AuthService.InitWithGORM(gormDB, "mcp_auth_keys")
			comp.AuthService.StartTokenRefresher(context.Background())
			comp.UserService = service.NewUserService(gormDB, "mcp_users")
			comp.MCPServerStore = database.NewGORMMCPServerDAO(gormDB)
			comp.ToolStore = database.NewGORMToolStore(gormDB)
			comp.ToolBindingStore = database.NewGORMToolServerBindingDAO(gormDB)
			comp.BuildInfoStore = database.NewGORMServerBuildInfoDAO(gormDB)
			comp.ToolDefStore = tooldef.NewEnhancedMySQLStore(gormDB, cfg.MySQLTable, appLogger)
		}
	}

	storeCfg := config.LoadHTTPServiceStoreConfig()
	if storeCfg.StoreType == "mysql" && storeCfg.MySQLDSN != "" {
		if gormDB, err := database.NewGORMDB(storeCfg.MySQLDSN); err == nil {
			comp.ServiceStore = httpservice.NewGORMServiceStore(gormDB, appLogger)
		}
	}
	if comp.ServiceStore == nil {
		comp.ServiceStore = httpservice.NewMemoryServiceStore(appLogger)
	}

	comp.HTTPServiceMgr = service.NewHTTPServiceManager(appLogger)
	comp.MCPServer = server.NewMCPServer(cfg.ServerName, cfg.ServerVersion, server.WithToolCapabilities(true), server.WithRecovery())
	comp.GroupMCP = service.NewMCPGroupManager(cfg.ServerName, cfg.ServerVersion, comp.AuthService)

	if comp.MCPServerStore != nil {
		comp.BuildService = service.NewServerBuildService(
			comp.MCPServerStore, comp.ToolStore, comp.ToolBindingStore,
			comp.BuildInfoStore, comp.ServiceStore,
		)
		comp.Registry = service.NewDynamicRegistry(
			comp.MCPServer, comp.ToolDefStore, cfg.RefreshInterval(),
			appLogger, comp.GroupMCP, comp.MCPServerStore, comp.BuildService,
		)
	} else {
		comp.Registry = service.NewDynamicRegistry(
			comp.MCPServer, comp.ToolDefStore, cfg.RefreshInterval(),
			appLogger, comp.GroupMCP, nil, nil,
		)
	}

	return comp, cleanup, nil
}

func main() {
	cfg := config.Load()

	comp, cleanup, err := initComponents(cfg)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer cleanup()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := comp.Registry.SyncOnce(ctx); err != nil {
		log.Fatalf("initial sync failed: %v", err)
	}
	go comp.Registry.Start(ctx)

	startHTTPServer(ctx, cfg, comp)
}

func startHTTPServer(ctx context.Context, cfg config.Config, comp *ServerComponents) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(commonmw.Recovery(comp.Registry.GetLogger()))
	engine.Use(commonmw.RequestID())
	engine.Use(commonmw.Trace(comp.Registry.GetLogger()))
	engine.Use(commonmw.Cors())
	engine.Use(commonmw.Logger(comp.Registry.GetLogger()))

	engine.RedirectTrailingSlash = false

	httpAPI.RegisterRoutes(
		engine, comp.Registry, comp.GroupMCP, comp.AuthService,
		comp.HTTPServiceMgr, comp.ServiceStore, comp.MCPServerStore,
		comp.ToolStore, comp.ToolBindingStore, comp.BuildInfoStore,
		comp.JWTManager, comp.Registry.GetLogger(),
	)

	if comp.UserService != nil && comp.JWTManager != nil {
		httpAPI.RegisterUserAuthRoutes(engine, comp.UserService, comp.JWTManager)
	}

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: engine}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("HTTP server error: %v", err)
	}
}
