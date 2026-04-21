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
	"gorm.io/gorm"
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

func (s *ServerBuildService) ComputeHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *ServerBuildService) BuildOrUpdate(ctx context.Context, serverID uint, tx ...*gorm.DB) error {
	var bindings []*model.ToolServerBinding
	var err error

	if len(tx) > 0 && tx[0] != nil {
		err = tx[0].WithContext(ctx).Where("server_id = ? AND state = ?", serverID, 1).Find(&bindings).Error
	} else {
		bindings, err = s.bindingStore.ListByServerID(ctx, serverID)
	}
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

		var params []model.ParameterDefinition
		if len(tool.Parameters) > 0 {
			_ = sonic.Unmarshal(tool.Parameters, &params)
		}

		var inputMapping []model.InputMappingField
		if len(tool.InputMapping) > 0 {
			_ = sonic.Unmarshal(tool.InputMapping, &inputMapping)
		}

		var outputMapping *model.OutputMappingConfig
		if len(tool.OutputMapping) > 0 {
			var fields []model.OutputMappingField
			if err := sonic.Unmarshal(tool.OutputMapping, &fields); err == nil && len(fields) > 0 {
				outputMapping = &model.OutputMappingConfig{Fields: fields}
			}
		}

		tools = append(tools, model.ToolSnapshot{
			ID:            tool.ID,
			Name:          tool.Name,
			Description:   tool.Description,
			ServiceID:     tool.ServiceID,
			Parameters:    params,
			InputMapping:  inputMapping,
			OutputMapping: outputMapping,
			State:         tool.State,
		})

		if tool.ServiceID > 0 && !seenServiceIDs[tool.ServiceID] {
			seenServiceIDs[tool.ServiceID] = true
			httpService, err := s.serviceStore.Get(ctx, tool.ServiceID)
			if err == nil && httpService != nil {
				headersBytes, _ := sonic.Marshal(httpService.Headers)
				httpServices = append(httpServices, model.HTTPServiceSnapshot{
					ID:           httpService.ID,
					Name:         httpService.Name,
					TargetURL:    httpService.TargetURL,
					Method:       httpService.Method,
					Headers:      headersBytes,
					BodyType:     httpService.BodyType,
					Timeout:      httpService.TimeoutSeconds,
					InputSchema:  httpService.InputSchema,
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

	var executor buildInfoExecutor
	if len(tx) > 0 && tx[0] != nil {
		executor = &serverBuildInfoHelper{tx: tx[0]}
	} else {
		executor = &buildInfoStoreWrapper{store: s.buildInfoStore}
	}

	activeBuild, err := executor.GetActiveByServerID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to get active build: %w", err)
	}

	if activeBuild != nil && activeBuild.Hash == newHash {
		return nil
	}

	if activeBuild != nil {
		if err := executor.UpdateState(ctx, activeBuild.ID, 0); err != nil {
			return fmt.Errorf("failed to deactivate old build: %w", err)
		}
	}

	maxVersion, err := executor.GetMaxVersionByServerID(ctx, serverID)
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

	if err := executor.Save(ctx, newBuild); err != nil {
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

type buildInfoExecutor interface {
	GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error)
	UpdateState(ctx context.Context, id uint, state int) error
	GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error)
	Save(ctx context.Context, build *model.ServerBuildInfo) error
}

type serverBuildInfoHelper struct {
	tx *gorm.DB
}

func (h *serverBuildInfoHelper) GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	var build model.ServerBuildInfo
	err := h.tx.WithContext(ctx).Where("server_id = ? AND state = ?", serverID, 1).First(&build).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &build, nil
}

func (h *serverBuildInfoHelper) UpdateState(ctx context.Context, id uint, state int) error {
	return h.tx.WithContext(ctx).Model(&model.ServerBuildInfo{}).Where("id = ?", id).Update("state", state).Error
}

func (h *serverBuildInfoHelper) GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error) {
	var maxVersion int
	err := h.tx.WithContext(ctx).Model(&model.ServerBuildInfo{}).Where("server_id = ?", serverID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error
	return maxVersion, err
}

func (h *serverBuildInfoHelper) Save(ctx context.Context, build *model.ServerBuildInfo) error {
	return h.tx.WithContext(ctx).Create(build).Error
}

type buildInfoStoreWrapper struct {
	store repository.ServerBuildInfoStore
}

func (w *buildInfoStoreWrapper) GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	return w.store.GetActiveByServerID(ctx, serverID)
}

func (w *buildInfoStoreWrapper) UpdateState(ctx context.Context, id uint, state int) error {
	return w.store.UpdateState(ctx, id, state)
}

func (w *buildInfoStoreWrapper) GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error) {
	return w.store.GetMaxVersionByServerID(ctx, serverID)
}

func (w *buildInfoStoreWrapper) Save(ctx context.Context, build *model.ServerBuildInfo) error {
	return w.store.Save(ctx, build)
}

func (s *ServerBuildService) GetActiveBuild(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	return s.buildInfoStore.GetActiveByServerID(ctx, serverID)
}
