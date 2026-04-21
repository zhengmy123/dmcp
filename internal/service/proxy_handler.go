package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"github.com/bytedance/sonic"
)

type ProxyHandler struct {
	serverStore repository.MCPServerStore
	proxyCache  *ProxyServerCacheService
	logger      logger.Logger
	httpClient  *http.Client
}

func NewProxyHandler(
	serverStore repository.MCPServerStore,
	proxyCache *ProxyServerCacheService,
	log logger.Logger,
) *ProxyHandler {
	return &ProxyHandler{
		serverStore: serverStore,
		proxyCache:  proxyCache,
		logger:      log,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vauthKey := strings.Trim(strings.TrimPrefix(r.URL.Path, mcpPathPrefix), "/")
	if vauthKey == "" || strings.Contains(vauthKey, "/") {
		mcpErrorResponse(w, http.StatusBadRequest, mcpCodeBadRequest, "path must be /mcp/{vauth_key}", r.URL.Path)
		return
	}

	proxyServer, err := h.getProxyServerWithCache(r.Context(), vauthKey)
	if err != nil {
		h.logger.Error("get proxy server failed", logger.String("vauth_key", vauthKey), logger.Error(err))
		mcpErrorResponse(w, http.StatusInternalServerError, mcpCodeInternalError, "internal server error", "")
		return
	}

	if proxyServer == nil {
		mcpErrorResponse(w, http.StatusNotFound, mcpCodeNotFound, "proxy server not found", vauthKey)
		return
	}

	if proxyServer.State != 1 {
		mcpErrorResponse(w, http.StatusForbidden, mcpCodeForbidden, "proxy server is disabled", "")
		return
	}

	resp, err := h.proxyRequest(r.Context(), proxyServer, r)
	if err != nil {
		h.logger.Error("proxy request failed", logger.String("vauth_key", vauthKey), logger.Error(err))
		if strings.Contains(err.Error(), "timeout") {
			mcpErrorResponse(w, http.StatusGatewayTimeout, mcpCodeInternalError, "proxy request timeout", "")
			return
		}
		mcpErrorResponse(w, http.StatusBadGateway, mcpCodeInternalError, "proxy request failed", err.Error())
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("read proxy response body failed", logger.Error(err))
		return
	}
	_, _ = w.Write(body)
}

func (h *ProxyHandler) getProxyServerWithCache(ctx context.Context, vauthKey string) (*CachedProxyServer, error) {
	if h.proxyCache != nil {
		cached, err := h.proxyCache.GetProxyServer(ctx, vauthKey)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	server, err := h.serverStore.GetByVAuthKey(ctx, vauthKey)
	if err != nil {
		return nil, fmt.Errorf("get server by vauth key: %w", err)
	}
	if server == nil {
		return nil, nil
	}
	if server.Type != "proxy" {
		return nil, nil
	}

	var extraHeaders map[string]string
	if server.ExtraHeaders != "" {
		if err := sonic.Unmarshal([]byte(server.ExtraHeaders), &extraHeaders); err != nil {
			h.logger.Warn("parse extra headers failed", logger.Error(err))
			extraHeaders = make(map[string]string)
		}
	}

	timeoutSeconds := server.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	cached := &CachedProxyServer{
		ID:             server.ID,
		VAuthKey:       server.VAuthKey,
		Name:           server.Name,
		HTTPServerURL:  server.HTTPServerURL,
		AuthHeader:     server.AuthHeader,
		TimeoutSeconds: timeoutSeconds,
		ExtraHeaders:   extraHeaders,
		State:          server.State,
	}

	if h.proxyCache != nil {
		if err := h.proxyCache.SetProxyServer(ctx, vauthKey, cached); err != nil {
			h.logger.Warn("set proxy server cache failed", logger.Error(err))
		}
	}

	return cached, nil
}

func (h *ProxyHandler) proxyRequest(ctx context.Context, server *CachedProxyServer, r *http.Request) (*http.Response, error) {
	targetURL := server.HTTPServerURL
	if !strings.Contains(targetURL, "/mcp") {
		targetURL = strings.TrimSuffix(targetURL, "/") + "/mcp"
	} else if strings.HasSuffix(targetURL, "/") {
		targetURL = strings.TrimSuffix(targetURL, "/")
	}

	var body io.ReadCloser
	if r.Body != nil {
		body = r.Body
	} else {
		body = io.NopCloser(bytes.NewReader([]byte{}))
	}

	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("create proxy request: %w", err)
	}

	for k, vs := range r.Header {
		for _, v := range vs {
			proxyReq.Header.Add(k, v)
		}
	}

	h.logger.Info("proxy request headers before auth",
		logger.String("auth_header", server.AuthHeader),
		logger.Any("extra_headers", server.ExtraHeaders))

	if server.AuthHeader != "" {
		parts := strings.SplitN(server.AuthHeader, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			proxyReq.Header.Set(key, value)
			h.logger.Info("set auth header", logger.String("key", key), logger.String("value", value))
		}
	}

	for k, v := range server.ExtraHeaders {
		proxyReq.Header.Set(k, v)
		h.logger.Info("set extra header", logger.String("key", k), logger.String("value", v))
	}

	client := &http.Client{Timeout: time.Duration(server.TimeoutSeconds) * time.Second}
	return client.Do(proxyReq)
}

func (h *ProxyHandler) InvalidateCache(ctx context.Context, vauthKey string) {
	if h.proxyCache != nil {
		if err := h.proxyCache.DeleteProxyServer(ctx, vauthKey); err != nil {
			h.logger.Warn("invalidate proxy cache failed", logger.String("vauth_key", vauthKey), logger.Error(err))
		}
	}
}

func (h *ProxyHandler) UpdateCache(ctx context.Context, server *model.MCPServer) {
	if h.proxyCache == nil || server.Type != "proxy" {
		return
	}

	var extraHeaders map[string]string
	if server.ExtraHeaders != "" {
		if err := sonic.Unmarshal([]byte(server.ExtraHeaders), &extraHeaders); err != nil {
			h.logger.Warn("parse extra headers failed", logger.Error(err))
			extraHeaders = make(map[string]string)
		}
	}

	timeoutSeconds := server.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	cached := &CachedProxyServer{
		ID:             server.ID,
		VAuthKey:       server.VAuthKey,
		Name:           server.Name,
		HTTPServerURL:  server.HTTPServerURL,
		AuthHeader:     server.AuthHeader,
		TimeoutSeconds: timeoutSeconds,
		ExtraHeaders:   extraHeaders,
		State:          server.State,
	}

	if err := h.proxyCache.UpdateProxyServer(ctx, server.VAuthKey, cached); err != nil {
		h.logger.Warn("update proxy cache failed", logger.String("vauth_key", server.VAuthKey), logger.Error(err))
	}
}

var _ http.Handler = (*ProxyHandler)(nil)
