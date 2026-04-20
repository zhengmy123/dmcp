package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

type ServerBuildService struct {
	serverStore    repository.MCPServerStore
	toolStore      repository.ToolStore
	bindingStore   repository.ToolServerBindingStore
	buildInfoStore repository.ServerBuildInfoStore
	serviceStore   repository.ServiceStore
	buildInfoCache *BuildInfoCacheService
}

func NewServerBuildService(
	serverStore repository.MCPServerStore,
	toolStore repository.ToolStore,
	bindingStore repository.ToolServerBindingStore,
	buildInfoStore repository.ServerBuildInfoStore,
	serviceStore repository.ServiceStore,
	buildInfoCache *BuildInfoCacheService,
) *ServerBuildService {
	return &ServerBuildService{
		serverStore:    serverStore,
		toolStore:      toolStore,
		bindingStore:   bindingStore,
		buildInfoStore: buildInfoStore,
		serviceStore:   serviceStore,
		buildInfoCache: buildInfoCache,
	}
}

func (s *ServerBuildService) BuildOrUpdate(ctx context.Context, serverID uint) error {
	bindings, err := s.bindingStore.ListByServerID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to list bindings: %w", err)
	}

	var tools []model.ToolSnapshot
	var httpServices []model.HTTPServiceSnapshot
	seenServiceIDs := make(map[uint]bool)

	for _, binding := range bindings {
		tool, err := s.toolStore.GetByID(ctx, binding.ToolID)
		if err != nil {
			continue
		}
		tools = append(tools, model.ToolSnapshot{
			ID:          tool.ID,
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
			Enabled:     tool.Enabled,
		})

		if tool.ServiceID > 0 && !seenServiceIDs[tool.ServiceID] {
			seenServiceIDs[tool.ServiceID] = true
			httpService, err := s.serviceStore.Get(ctx, tool.ServiceID)
			if err == nil && httpService != nil {
				httpServices = append(httpServices, model.HTTPServiceSnapshot{
					ID:           httpService.ID,
					Name:         httpService.Name,
					TargetURL:    httpService.TargetURL,
					Method:       httpService.Method,
					Headers:      httpService.Headers,
					BodyType:    httpService.BodyType,
					Timeout:     httpService.TimeoutSeconds,
					InputSchema: httpService.InputSchema,
					OutputSchema: httpService.OutputSchema,
				})
			}
		}
	}

	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	buildData := model.BuildData{
		Tools:        tools,
		HTTPServices: httpServices,
	}

	buildDataJSON, err := sonic.Marshal(buildData)
	if err != nil {
		return fmt.Errorf("failed to marshal build data: %w", err)
	}

	newHash := s.ComputeHash(buildDataJSON)

	activeBuild, err := s.buildInfoStore.GetActiveByServerID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to get active build: %w", err)
	}

	if activeBuild != nil && activeBuild.Hash == newHash {
		return nil
	}

	if activeBuild != nil {
		if err := s.buildInfoStore.UpdateState(ctx, activeBuild.ID, 0); err != nil {
			return fmt.Errorf("failed to deactivate old build: %w", err)
		}
	}

	maxVersion, err := s.buildInfoStore.GetMaxVersionByServerID(ctx, serverID)
	if err != nil {
		maxVersion = 0
	}

	newBuild := &model.ServerBuildInfo{
		ServerID:  serverID,
		Version:   maxVersion + 1,
		BuildUUID: uuid.New().String(),
		Hash:      newHash,
		BuildData: string(buildDataJSON),
		State:     1,
	}

	if err := s.buildInfoStore.Save(ctx, newBuild); err != nil {
		return fmt.Errorf("failed to save new build: %w", err)
	}

	if s.buildInfoCache != nil {
		server, err := s.serverStore.GetByID(ctx, serverID)
		if err == nil && server != nil {
			_ = s.buildInfoCache.DeleteBuildUUID(ctx, server.VAuthKey)
		}
	}

	return nil
}

func (s *ServerBuildService) ComputeHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *ServerBuildService) GetActiveBuild(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	return s.buildInfoStore.GetActiveByServerID(ctx, serverID)
}
