# MCP VAuth Key 调用逻辑和工具绑定逻辑重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构 MCP VAuth Key 调用逻辑和工具绑定逻辑，新增构建信息表，工具绑定时生成构建版本，工具删除时校验绑定关系

**Architecture:** 新增 `server_build_info` 表存储构建快照，工具绑定/变动时触发构建更新，VAuth Key 查询改为从构建表获取

**Tech Stack:** Go, GORM, MySQL, Sonic JSON

---

## 文件结构

```
domain/model/server_build_info.go           # 新增：构建信息模型
domain/repository/server_build_info_repository.go  # 新增：仓储接口
infrastructure/database/gorm_server_build_info.go # 新增：GORM 实现
service/server_build_service.go              # 新增：构建更新服务
service/tool_binding_service.go              # 修改：绑定后触发构建
service/tool_service.go                      # 修改：工具变动触发构建
service/registry.go                         # 修改：使用构建信息表查询
domain/service/tool_domain_service.go        # 修改：增加删除校验
docs/mysql_migration.sql                     # 修改：添加建表语句
```

---

## Task 1: 创建 ServerBuildInfo 数据模型

**Files:**
- Create: `internal/domain/model/server_build_info.go`
- Test: `test/domain/model/server_build_info_test.go`

- [ ] **Step 1: 创建模型文件**

```go
package model

import (
	"time"
)

type ServerBuildInfo struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ServerID  uint      `json:"server_id" gorm:"not null;index:idx_server_state"`
	Version   int       `json:"version" gorm:"not null;default:1"`
	BuildUUID string    `json:"build_uuid" gorm:"size:36;not null;uniqueIndex"`
	Hash      string    `json:"hash" gorm:"size:64;not null;index"`
	BuildData string    `json:"build_data" gorm:"type:text"`
	State     int       `json:"state" gorm:"not null;default:1;index:idx_server_state"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ServerBuildInfo) TableName() string {
	return "server_build_info"
}

type BuildData struct {
	Tools       []ToolSnapshot       `json:"tools"`
	HTTPServices []HTTPServiceSnapshot `json:"http_services"`
}

type ToolSnapshot struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  []byte `json:"parameters"`
	Enabled     bool   `json:"enabled"`
}

type HTTPServiceSnapshot struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	TargetURL   string                 `json:"target_url"`
	Method      string                 `json:"method"`
	Headers     map[string]string      `json:"headers"`
	BodyType    string                 `json:"body_type"`
	Timeout     int                    `json:"timeout_seconds"`
	InputSchema []byte                 `json:"input_schema"`
	OutputSchema []byte                `json:"output_schema"`
}
```

- [ ] **Step 2: 创建测试文件**

```go
package model

import (
	"testing"
)

func TestServerBuildInfo_TableName(t *testing.T) {
	info := ServerBuildInfo{}
	if info.TableName() != "server_build_info" {
		t.Errorf("expected table name 'server_build_info', got '%s'", info.TableName())
	}
}

func TestBuildData_Structure(t *testing.T) {
	data := BuildData{
		Tools: []ToolSnapshot{
			{ID: 1, Name: "test_tool", Description: "test", Enabled: true},
		},
		HTTPServices: []HTTPServiceSnapshot{
			{ID: 1, Name: "test_service", TargetURL: "http://test.com"},
		},
	}
	if len(data.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(data.Tools))
	}
	if len(data.HTTPServices) != 1 {
		t.Errorf("expected 1 http service, got %d", len(data.HTTPServices))
	}
}
```

- [ ] **Step 3: 运行测试验证**

Run: `go test ./test/domain/model/server_build_info_test.go -v`
Expected: PASS

---

## Task 2: 创建 ServerBuildInfo 仓储接口

**Files:**
- Create: `internal/domain/repository/server_build_info_repository.go`
- Test: `test/infrastructure/store/server_build_info_test.go`

- [ ] **Step 1: 创建仓储接口**

```go
package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

type ServerBuildInfoStore interface {
	GetByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error)
	GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error)
	GetByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error)
	Save(ctx context.Context, info *model.ServerBuildInfo) error
	UpdateState(ctx context.Context, id uint, state int) error
	GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error)
}
```

- [ ] **Step 2: 创建测试文件**

```go
package store

import (
	"testing"
)

func TestServerBuildInfoStore_Interface(t *testing.T) {
	// Verify interface is implemented correctly
	var _ interface{} = (*mockServerBuildInfoStore)(nil)
}

type mockServerBuildInfoStore struct{}

func (m *mockServerBuildInfoStore) GetByServerID(ctx interface{}, serverID uint) error {
	return nil
}
```

---

## Task 3: 创建 GORM ServerBuildInfo 实现

**Files:**
- Create: `internal/infrastructure/database/gorm_server_build_info.go`

- [ ] **Step 1: 创建 GORM 实现**

```go
package database

import (
	"context"
	"dynamic_mcp_go_server/internal/domain/model"
	"errors"

	"gorm.io/gorm"
)

type GORMServerBuildInfoDAO struct {
	db *gorm.DB
}

func NewGORMServerBuildInfoDAO(db *gorm.DB) *GORMServerBuildInfoDAO {
	return &GORMServerBuildInfoDAO{db: db}
}

func (d *GORMServerBuildInfoDAO) GetByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	var info model.ServerBuildInfo
	err := d.db.WithContext(ctx).Where("server_id = ?", serverID).First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *GORMServerBuildInfoDAO) GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	var info model.ServerBuildInfo
	err := d.db.WithContext(ctx).Where("server_id = ? AND state = ?", serverID, 1).First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *GORMServerBuildInfoDAO) GetByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error) {
	var info model.ServerBuildInfo
	err := d.db.WithContext(ctx).Where("build_uuid = ?", buildUUID).First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *GORMServerBuildInfoDAO) Save(ctx context.Context, info *model.ServerBuildInfo) error {
	return d.db.WithContext(ctx).Create(info).Error
}

func (d *GORMServerBuildInfoDAO) UpdateState(ctx context.Context, id uint, state int) error {
	return d.db.WithContext(ctx).Model(&model.ServerBuildInfo{}).Where("id = ?", id).Update("state", state).Error
}

func (d *GORMServerBuildInfoDAO) GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error) {
	var maxVersion int
	err := d.db.WithContext(ctx).Model(&model.ServerBuildInfo{}).
		Where("server_id = ?", serverID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return 0, err
	}
	return maxVersion, nil
}
```

---

## Task 4: 创建 ServerBuildService

**Files:**
- Create: `service/server_build_service.go`
- Test: `test/service/server_build_service_test.go`

- [ ] **Step 1: 创建服务**

```go
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
}

func NewServerBuildService(
	serverStore repository.MCPServerStore,
	toolStore repository.ToolStore,
	bindingStore repository.ToolServerBindingStore,
	buildInfoStore repository.ServerBuildInfoStore,
) *ServerBuildService {
	return &ServerBuildService{
		serverStore:    serverStore,
		toolStore:      toolStore,
		bindingStore:   bindingStore,
		buildInfoStore: buildInfoStore,
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
			httpService, err := s.getHTTPService(ctx, tool.ServiceID)
			if err == nil && httpService != nil {
				httpServices = append(httpServices, *httpService)
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

	newHash := s.computeHash(buildDataJSON)

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

	return nil
}

func (s *ServerBuildService) computeHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *ServerBuildService) getHTTPService(ctx context.Context, serviceID uint) (*model.HTTPServiceSnapshot, error) {
	return nil, nil
}
```

- [ ] **Step 2: 创建测试文件**

```go
package service

import (
	"testing"
)

func TestServerBuildService_ComputeHash(t *testing.T) {
	svc := &ServerBuildService{}

	data1 := []byte(`{"tools":[{"name":"a"}]}`)
	data2 := []byte(`{"tools":[{"name":"a"}]}`)
	data3 := []byte(`{"tools":[{"name":"b"}]}`)

	hash1 := svc.computeHash(data1)
	hash2 := svc.computeHash(data2)
	hash3 := svc.computeHash(data3)

	if hash1 != hash2 {
		t.Error("same data should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("different data should produce different hash")
	}

	if len(hash1) != 64 {
		t.Errorf("expected SHA256 hash length 64, got %d", len(hash1))
	}
}
```

- [ ] **Step 3: 运行测试**

Run: `go test ./test/service/server_build_service_test.go -v`
Expected: PASS

---

## Task 5: 修改 ToolBindingService 绑定后触发构建

**Files:**
- Modify: `internal/service/tool_binding_service.go`

- [ ] **Step 1: 添加 ServerBuildService 依赖**

```go
type ToolBindingService struct {
	bindingStore repository.ToolServerBindingStore
	toolStore    repository.ToolStore
	serverStore  repository.MCPServerStore
	buildSvc     *ServerBuildService
}

func NewToolBindingService(
	bindingStore repository.ToolServerBindingStore,
	toolStore repository.ToolStore,
	serverStore repository.MCPServerStore,
	buildSvc *ServerBuildService,
) *ToolBindingService {
	return &ToolBindingService{
		bindingStore: bindingStore,
		toolStore:    toolStore,
		serverStore:  serverStore,
		buildSvc:     buildSvc,
	}
}
```

- [ ] **Step 2: 修改 BindTool 方法，绑定成功后触发构建**

在 `BindTool` 方法的 `Save` 成功后添加：
```go
if err := s.bindingStore.Save(ctx, binding); err != nil {
	return nil, err
}

if s.buildSvc != nil {
	_ = s.buildSvc.BuildOrUpdate(ctx, req.ServerID)
}

return binding, nil
```

- [ ] **Step 3: 修改 BatchBindTools 方法，绑定成功后触发构建**

在批量绑定循环后，对涉及的每个 server 调用构建：
```go
if len(toCreate) > 0 {
	if err := s.bindingStore.BatchSave(ctx, toCreate); err != nil {
		return 0, fmt.Errorf("failed to create bindings: %w", err)
	}
}

affectedServers := make(map[uint]bool)
for _, toolID := range req.ToolIDs {
	for _, serverID := range req.ServerIDs {
		affectedServers[serverID] = true
	}
}
for serverID := range affectedServers {
	if s.buildSvc != nil {
		_ = s.buildSvc.BuildOrUpdate(ctx, serverID)
	}
}

return len(toRestore) + len(toCreate), nil
```

---

## Task 6: 修改 ToolDomainService 增加删除校验

**Files:**
- Modify: `internal/domain/service/tool_domain_service.go`

- [ ] **Step 1: 添加 ErrToolHasActiveBinding 错误**

```go
var (
	ErrOnlyHTTPServiceServerCanHaveTools = errors.New("only http_service server can have tools")
	ErrToolNameAlreadyExists             = errors.New("tool with same name already exists in this server")
	ErrHTTPServiceNotFound               = errors.New("http service not found")
	ErrToolHasActiveBinding              = errors.New("tool has active binding, unbind first")
)
```

- [ ] **Step 2: 添加 HasActiveBinding 方法**

```go
func (s *ToolDomainService) HasActiveBinding(ctx context.Context, toolID uint) (bool, error) {
	bindings, err := s.toolStore.GetBindingsByToolID(ctx, toolID)
	if err != nil {
		return false, err
	}
	return len(bindings) > 0, nil
}
```

- [ ] **Step 3: 修改 ToolStore 接口添加 GetBindingsByToolID**

在 `internal/domain/repository/tool_repository.go` 添加：
```go
GetBindingsByToolID(ctx context.Context, toolID uint) ([]*model.ToolServerBinding, error)
```

- [ ] **Step 4: 在 GORM 实现中添加 GetBindingsByToolID**

在 `internal/infrastructure/database/gorm_tool_store.go` 添加：
```go
func (d *GORMToolStore) GetBindingsByToolID(ctx context.Context, toolID uint) ([]*model.ToolServerBinding, error) {
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("tool_id = ? AND state = ?", toolID, 1).Find(&bindings).Error
	return bindings, err
}
```

---

## Task 7: 修改 Registry 实现 VAuth Key 查询

**Files:**
- Modify: `internal/service/registry.go`

- [ ] **Step 1: 添加 ServerBuildService 依赖**

```go
type DynamicRegistry struct {
	server        *server.MCPServer
	store         tooldef.Store
	interval      time.Duration
	logger        logger.Logger
	groupMCP      *MCPGroupManager
	serverName    string
	serverVersion string
	lastHash      string
	mu            sync.RWMutex
	lastDefs      []tooldef.ToolDefinition
	buildSvc      *ServerBuildService
	serverStore   repository.MCPServerStore
}
```

- [ ] **Step 2: 修改 ListDefinitionsByVAuthKey**

```go
func (d *DynamicRegistry) ListDefinitionsByVAuthKey(vauthKey string) []tooldef.ToolDefinition {
	if d.buildSvc == nil || d.serverStore == nil {
		return nil
	}

	ctx := context.Background()
	mcpServer, err := d.serverStore.GetByVAuthKey(ctx, vauthKey)
	if err != nil || mcpServer == nil {
		return nil
	}

	buildInfo, err := d.buildSvc.GetActiveBuild(ctx, mcpServer.ID)
	if err != nil || buildInfo == nil {
		return nil
	}

	var buildData model.BuildData
	if err := sonic.Unmarshal([]byte(buildInfo.BuildData), &buildData); err != nil {
		return nil
	}

	defs := make([]tooldef.ToolDefinition, 0, len(buildData.Tools))
	for _, t := range buildData.Tools {
		if t.Enabled {
			defs = append(defs, tooldef.ToolDefinition{
				ID:          t.ID,
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
				Enabled:     t.Enabled,
			})
		}
	}
	return defs
}
```

- [ ] **Step 3: 修改 GetDefinitionByVAuthKey**

```go
func (d *DynamicRegistry) GetDefinitionByVAuthKey(vauthKey, name string) (tooldef.ToolDefinition, bool) {
	defs := d.ListDefinitionsByVAuthKey(vauthKey)
	for _, def := range defs {
		if def.Name == name {
			return def, true
		}
	}
	return tooldef.ToolDefinition{}, false
}
```

- [ ] **Step 4: 在 ServerBuildService 添加 GetActiveBuild 方法**

```go
func (s *ServerBuildService) GetActiveBuild(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	return s.buildInfoStore.GetActiveByServerID(ctx, serverID)
}
```

---

## Task 8: 更新 MySQL 迁移脚本

**Files:**
- Modify: `docs/mysql_migration.sql`

- [ ] **Step 1: 添加建表语句**

```sql
CREATE TABLE IF NOT EXISTS server_build_info (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    server_id BIGINT NOT NULL COMMENT '关联 mcp_servers.id',
    version INT NOT NULL DEFAULT 1 COMMENT '版本号',
    build_uuid VARCHAR(36) NOT NULL COMMENT '构建UUID',
    hash VARCHAR(64) NOT NULL COMMENT 'SHA256',
    build_data TEXT COMMENT 'JSON: 工具和HTTP服务快照合并',
    state INT NOT NULL DEFAULT 1 COMMENT '1-有效 0-失效',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_build_uuid (build_uuid),
    INDEX idx_hash (hash),
    INDEX idx_server_state (server_id, state)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MCP Server 构建信息表';
```

---

## Task 9: 更新 main.go 初始化逻辑

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 添加 ServerBuildService 初始化**

找到现有服务初始化位置，添加：
```go
serverBuildInfoDAO := database.NewGORMServerBuildInfoDAO(db)
serverBuildInfoStore := serverBuildInfoDAO

buildSvc := service.NewServerBuildService(
	serverStore,
	toolStore,
	bindingStore,
	serverBuildInfoStore,
)
```

- [ ] **Step 2: 传递 buildSvc 到相关服务**

在 `NewToolBindingService`、`NewDynamicRegistry` 等调用处添加 buildSvc 参数。

---

## Task 10: 更新 ToolService 工具变动触发构建

**Files:**
- Modify: `internal/service/tool_service.go`

- [ ] **Step 1: 添加 ServerBuildService 依赖**

```go
type ToolService struct {
	toolDomainService *domainService.ToolDomainService
	buildSvc          *ServerBuildService
}

func NewToolService(
	toolDomainService *domainService.ToolDomainService,
	buildSvc *ServerBuildService,
) *ToolService {
	return &ToolService{
		toolDomainService: toolDomainService,
		buildSvc:          buildSvc,
	}
}
```

- [ ] **Step 2: 在 CreateFromHTTPService 成功后触发构建**

```go
func (s *ToolService) CreateFromHTTPService(ctx context.Context, cmd domainService.CreateToolFromHTTPServiceCommand) (*model.ToolDefinition, error) {
	tool, err := s.toolDomainService.CreateToolFromHTTPService(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.buildSvc != nil {
		_ = s.buildSvc.BuildOrUpdate(ctx, cmd.ServerID)
	}

	return tool, nil
}
```

---

## 自检清单

- [ ] 所有 Task 完成
- [ ] 所有测试通过
- [ ] 符合 AGENTS.md 规范
- [ ] MySQL 迁移脚本已更新
- [ ] 错误处理完善
- [ ] 代码无 TODO/TBD 占位符

---

**Plan complete.** 实施方式选择：
1. **Subagent-Driven (recommended)** - 派遣子代理逐任务执行
2. **Inline Execution** - 在当前会话中批量执行
