package database

import (
	"context"
	"fmt"

	"dynamic_mcp_go_server/internal/domain/model"

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

	result := d.db.WithContext(ctx).Where("enabled = ?", true).Find(&servers)
	if result.Error != nil {
		return nil, result.Error
	}

	return servers, nil
}

// GetByID 根据ID获取MCP服务器
func (d *GORMMCPServerDAO) GetByID(ctx context.Context, id uint) (*model.MCPServer, error) {
	var server model.MCPServer

	result := d.db.WithContext(ctx).Where("id = ? AND enabled = ?", id, true).First(&server)
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

	result := d.db.WithContext(ctx).Where("vauth_key = ? AND enabled = ?", vauthKey, true).First(&server)
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

// Delete 删除MCP服务器
func (d *GORMMCPServerDAO) Delete(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Delete(&model.MCPServer{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete mcp server failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mcp server not found: %d", id)
	}
	return nil
}
