package main

import (
	"context"
	"fmt"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/config"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/auth"
	"dynamic_mcp_go_server/internal/infrastructure/database"
	"dynamic_mcp_go_server/internal/infrastructure/store/httpservice"
	"dynamic_mcp_go_server/internal/infrastructure/store/tooldef"
	"dynamic_mcp_go_server/internal/service"

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

func newServerComponents(cfg config.Config) (*ServerComponents, func(), error) {
	appLogger, loggerCleanup, err := logger.NewFileLogger("logs", "server.log")
	if err != nil {
		return nil, nil, fmt.Errorf("init logger failed: %w", err)
	}

	comp := &ServerComponents{}
	cleanup := func() { loggerCleanup() }

	comp.AuthService = service.NewAuthService(cfg.AdminToken)
	comp.JWTManager = auth.NewJWTManager(cfg.JWTSecret, 0)

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
