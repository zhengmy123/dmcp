package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

type MCPServerQuery struct {
	Page     int
	PageSize int
	Name     string
	State    *int
}

type MCPServerWithToolCount struct {
	Server    *model.MCPServer
	ToolCount int64
}

type MCPServerStore interface {
	List(ctx context.Context) ([]*model.MCPServer, error)
	ListWithToolCount(ctx context.Context, query *MCPServerQuery) ([]*MCPServerWithToolCount, int64, error)
	GetByID(ctx context.Context, id uint) (*model.MCPServer, error)
	GetByVAuthKey(ctx context.Context, vauthKey string) (*model.MCPServer, error)
	Save(ctx context.Context, server *model.MCPServer) error
	Delete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
}
