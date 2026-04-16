package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

// TokenServerBindingStore 定义令牌与服务器绑定存储接口
type TokenServerBindingStore interface {
	ListByTokenID(ctx context.Context, tokenID uint) ([]*model.TokenServerBinding, error)
	ListByServerID(ctx context.Context, serverID uint) ([]*model.TokenServerBinding, error)
	Save(ctx context.Context, binding *model.TokenServerBinding) error
	DeleteByTokenID(ctx context.Context, tokenID uint) error
	ReplaceByTokenID(ctx context.Context, tokenID uint, serverIDs []uint) error
}
