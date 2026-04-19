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

	appLog := comp.Registry.GetLogger()
	engine.Use(commonmw.Recovery(appLog))
	engine.Use(commonmw.RequestID())
	engine.Use(commonmw.Trace(appLog))
	engine.Use(commonmw.Cors())
	engine.Use(commonmw.Logger(appLog))

	engine.RedirectTrailingSlash = false

	httpAPI.RegisterRoutes(
		engine, comp.Registry, comp.GroupMCP, comp.AuthService,
		comp.HTTPServiceMgr, comp.ServiceStore, comp.MCPServerStore,
		comp.ToolStore, comp.ToolBindingStore, comp.BuildInfoStore,
		comp.JWTManager, appLog,
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
