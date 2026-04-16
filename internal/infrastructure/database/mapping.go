package database

import (
	"context"
	"strings"

	"dynamic_mcp_go_server/internal/domain/model"

	"github.com/google/uuid"
)

// MappingDAO 服务映射数据访问接口
type MappingDAO interface {
	Create(ctx context.Context, mapping *model.ServiceMapping) error
	Get(ctx context.Context, serviceID uint, vauthKey string) (*model.ServiceMapping, error)
	List(ctx context.Context) ([]*model.ServiceMapping, error)
	Update(ctx context.Context, mapping *model.ServiceMapping) error
	Delete(ctx context.Context, id uint) error
	Validate(ctx context.Context, serviceID uint, vauthKey string) (bool, string, error)
}

// GenerateUUID 生成UUID — 保留工具函数供外部使用
// Deprecated: 使用 github.com/google/uuid.New().String() 替代
func GenerateUUID() string {
	return uuid.New().String()
}

// CanonicalVAuthKey 规范化vauthKey
func CanonicalVAuthKey(vauthKey string) string {
	return strings.TrimSpace(vauthKey)
}
