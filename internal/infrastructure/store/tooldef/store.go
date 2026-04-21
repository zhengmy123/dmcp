package tooldef

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

// Store 工具定义存储接口
type Store interface {
	List(ctx context.Context) ([]model.ToolDefinition, error)
}

// MemoryStore 内存存储实现
type MemoryStore struct {
	definitions []model.ToolDefinition
}

// NewMemoryStore 创建内存存储
func NewMemoryStore(definitions []model.ToolDefinition) *MemoryStore {
	return &MemoryStore{definitions: definitions}
}

func (m *MemoryStore) List(context.Context) ([]model.ToolDefinition, error) {
	out := make([]model.ToolDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if def.State != 1 {
			continue
		}
		out = append(out, def)
	}
	return out, nil
}
