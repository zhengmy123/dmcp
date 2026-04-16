package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

// MCPServerStore 定义MCP服务器存储接口
type MCPServerStore interface {
	List(ctx context.Context) ([]*model.MCPServer, error)
	GetByID(ctx context.Context, id uint) (*model.MCPServer, error)
	GetByVAuthKey(ctx context.Context, vauthKey string) (*model.MCPServer, error)
	Save(ctx context.Context, server *model.MCPServer) error
	Delete(ctx context.Context, id uint) error
}
