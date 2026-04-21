package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"

	"gorm.io/gorm"
)

type MCPServerQuery struct {
	Page     int
	PageSize int
	Name     string
	State    *int
	Type     string
}

type MCPServerWithToolCount struct {
	Server    *model.MCPServer
	ToolCount int64
}

type MCPServerStore interface {
	DB() *gorm.DB
	List(ctx context.Context) ([]*model.MCPServer, error)
	ListWithToolCount(ctx context.Context, query *MCPServerQuery) ([]*MCPServerWithToolCount, int64, error)
	GetByID(ctx context.Context, id uint) (*model.MCPServer, error)
	GetByVAuthKey(ctx context.Context, vauthKey string) (*model.MCPServer, error)
	GetByName(ctx context.Context, name string) (*model.MCPServer, error)
	Save(ctx context.Context, server *model.MCPServer) error
	Delete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
	SaveWithTx(ctx context.Context, tx *gorm.DB, server *model.MCPServer) error
}
