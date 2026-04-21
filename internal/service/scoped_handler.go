package service

import (
	"log"
	"net/http"
	"strings"
)

type ScopedMCPHandler struct {
	manager      *MCPGroupManager
	proxyHandler *ProxyHandler
}

func NewScopedMCPHandler(manager *MCPGroupManager, proxyHandler *ProxyHandler) *ScopedMCPHandler {
	return &ScopedMCPHandler{
		manager:      manager,
		proxyHandler: proxyHandler,
	}
}

func (h *ScopedMCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] ScopedMCPHandler.ServeHTTP called: path=%s", r.URL.Path)
	
	groupKey := strings.Trim(strings.TrimPrefix(r.URL.Path, mcpPathPrefix), "/")
	log.Printf("[DEBUG] Extracted groupKey: %s", groupKey)
	
	if groupKey == "" || strings.Contains(groupKey, "/") {
		log.Printf("[DEBUG] Invalid groupKey, returning 404")
		mcpErrorResponse(w, http.StatusNotFound, mcpCodeNotFound, "path must be /mcp/{vauth_key}", r.URL.Path)
		return
	}

	if h.proxyHandler != nil {
		log.Printf("[DEBUG] Checking proxy handler for groupKey: %s", groupKey)
		proxyServer, err := h.proxyHandler.getProxyServerWithCache(r.Context(), groupKey)
		if err == nil && proxyServer != nil && proxyServer.State == 1 {
			log.Printf("[DEBUG] Proxy server found, delegating to proxy handler")
			h.proxyHandler.ServeHTTP(w, r)
			return
		}
		log.Printf("[DEBUG] Proxy server not available or not active for groupKey: %s", groupKey)
	}

	log.Printf("[DEBUG] Getting handler from manager for groupKey: %s", groupKey)
	handler, err := h.manager.GetHandler(groupKey)
	if err != nil {
		log.Printf("[DEBUG] Handler not found for groupKey: %s, err: %v", groupKey, err)
		mcpErrorResponse(w, http.StatusNotFound, mcpCodeNotFound, "mcp server not found", groupKey)
		return
	}

	log.Printf("[DEBUG] Handler found, delegating request for groupKey: %s", groupKey)
	handler.ServeHTTP(w, r)
}

var _ http.Handler = (*ScopedMCPHandler)(nil)

func CanonicalVAuthKey(vauthKey string) string {
	return strings.TrimSpace(vauthKey)
}