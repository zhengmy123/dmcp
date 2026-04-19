package repository_test

import (
	"context"
	"testing"

	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/domain/model"
)

type mockServerBuildInfoStore struct{}

func (m *mockServerBuildInfoStore) GetByServerID(ctx context.Context, serverID uint) ([]*model.ServerBuildInfo, error) {
	return nil, nil
}

func (m *mockServerBuildInfoStore) GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	return nil, nil
}

func (m *mockServerBuildInfoStore) GetByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error) {
	return nil, nil
}

func (m *mockServerBuildInfoStore) Save(ctx context.Context, info *model.ServerBuildInfo) error {
	return nil
}

func (m *mockServerBuildInfoStore) UpdateState(ctx context.Context, id uint, state int) error {
	return nil
}

func (m *mockServerBuildInfoStore) GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error) {
	return 0, nil
}

func TestServerBuildInfoStore_Interface(t *testing.T) {
	var _ repository.ServerBuildInfoStore = (*mockServerBuildInfoStore)(nil)
}
