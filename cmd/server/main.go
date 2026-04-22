package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAPI "dynamic_mcp_go_server/internal/api/http"
	commonmw "dynamic_mcp_go_server/internal/common/middleware"
	"dynamic_mcp_go_server/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	comp, cleanup, err := newServerComponents(cfg)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer cleanup()

	startHTTPServer(cfg, comp)
}

func startHTTPServer(cfg config.Config, comp *ServerComponents) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	appLog := comp.Registry.GetLogger()
	engine.Use(commonmw.Recovery(appLog))
	engine.Use(commonmw.RequestID())
	engine.Use(commonmw.Trace(appLog))
	engine.Use(commonmw.Cors())
	engine.Use(commonmw.Logger(appLog))

	engine.RedirectTrailingSlash = false

	httpAPI.RegisterRoutes(
		engine, comp.Registry, comp.GroupMCP, comp.ProxyHandler, comp.AuthService,
		comp.HTTPServiceMgr, comp.ServiceStore, comp.MCPServerStore,
		comp.ToolStore, comp.ToolBindingStore, comp.BuildInfoStore,
		comp.JWTManager, appLog, comp.GORMDB, comp.SystemConfigStore,
	)

	if comp.UserService != nil && comp.JWTManager != nil {
		httpAPI.RegisterUserAuthRoutes(engine, comp.UserService, comp.JWTManager)
	}

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: engine}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
