package database

import (
	"context"
	"fmt"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"gorm.io/gorm"
)

// GORMMCPServerDAO GORM实现的MCP服务器DAO
type GORMMCPServerDAO struct {
	db *gorm.DB
}

// NewGORMMCPServerDAO 创建GORM MCP服务器DAO
func NewGORMMCPServerDAO(db *gorm.DB) *GORMMCPServerDAO {
	return &GORMMCPServerDAO{db: db}
}

// List 获取所有MCP服务器
func (d *GORMMCPServerDAO) List(ctx context.Context) ([]*model.MCPServer, error) {
	var servers []*model.MCPServer

	result := d.db.WithContext(ctx).Where("state = ?", 1).Find(&servers)
	if result.Error != nil {
		return nil, result.Error
	}

	return servers, nil
}

// ListWithToolCount 分页查询MCP服务器并统计工具数量
func (d *GORMMCPServerDAO) ListWithToolCount(ctx context.Context, query *repository.MCPServerQuery) ([]*repository.MCPServerWithToolCount, int64, error) {
	db := d.db.WithContext(ctx)

	whereClause := "1=1"
	args := []interface{}{}

	if query.Name != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+query.Name+"%")
	}
	if query.State != nil {
		whereClause += " AND state = ?"
		args = append(args, *query.State)
	}

	var total int64
	if err := db.Model(&model.MCPServer{}).Where(whereClause, args...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	var servers []*model.MCPServer
	if err := db.Where(whereClause, args...).
		Order("id DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&servers).Error; err != nil {
		return nil, 0, err
	}

	if len(servers) == 0 {
		return nil, total, nil
	}

	serverIDs := make([]uint, len(servers))
	for i, s := range servers {
		serverIDs[i] = s.ID
	}

	type toolCountResult struct {
		ServerID  uint  `gorm:"column:server_id"`
		ToolCount int64 `gorm:"column:tool_count"`
	}

	var toolCounts []toolCountResult
	if err := db.Model(&model.ToolServerBinding{}).
		Select("server_id, COUNT(*) as tool_count").
		Where("server_id IN ? AND state = ?", serverIDs, 1).
		Group("server_id").
		Scan(&toolCounts).Error; err != nil {
		return nil, 0, err
	}

	toolCountMap := make(map[uint]int64)
	for _, tc := range toolCounts {
		toolCountMap[tc.ServerID] = tc.ToolCount
	}

	results := make([]*repository.MCPServerWithToolCount, len(servers))
	for i, s := range servers {
		results[i] = &repository.MCPServerWithToolCount{
			Server:    s,
			ToolCount: toolCountMap[s.ID],
		}
	}

	return results, total, nil
}

// GetByID 根据ID获取MCP服务器
func (d *GORMMCPServerDAO) GetByID(ctx context.Context, id uint) (*model.MCPServer, error) {
	var server model.MCPServer

	result := d.db.WithContext(ctx).Where("id = ? AND state = ?", id, 1).First(&server)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("mcp server not found")
		}
		return nil, result.Error
	}

	return &server, nil
}

// GetByVAuthKey 根据VAuthKey获取MCP服务器
func (d *GORMMCPServerDAO) GetByVAuthKey(ctx context.Context, vauthKey string) (*model.MCPServer, error) {
	var server model.MCPServer

	result := d.db.WithContext(ctx).Where("v_auth_key = ? AND state = ?", vauthKey, 1).First(&server)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("mcp server not found")
		}
		return nil, result.Error
	}

	return &server, nil
}

// Save 保存MCP服务器（创建或更新）
func (d *GORMMCPServerDAO) Save(ctx context.Context, server *model.MCPServer) error {
	// 如果没有ID，创建新记录
	if server.ID == 0 {
		return d.db.WithContext(ctx).Create(server).Error
	}

	var existing model.MCPServer
	result := d.db.WithContext(ctx).Where("id = ?", server.ID).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return d.db.WithContext(ctx).Create(server).Error
		}
		return result.Error
	}

	return d.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ?", server.ID).Updates(server).Error
}

// Delete 删除MCP服务器（软删除）
func (d *GORMMCPServerDAO) Delete(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ? AND state = ?", id, 1).Update("state", 0)
	if result.Error != nil {
		return fmt.Errorf("delete mcp server failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mcp server not found: %d", id)
	}
	return nil
}

// Restore 恢复MCP服务器
func (d *GORMMCPServerDAO) Restore(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ? AND state = ?", id, 0).Update("state", 1)
	if result.Error != nil {
		return fmt.Errorf("restore mcp server failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mcp server not found or already active: %d", id)
	}
	return nil
}
