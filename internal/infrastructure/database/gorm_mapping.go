package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"

	"gorm.io/gorm"
)

// GORMMappingDAO GORM实现的服务映射DAO
type GORMMappingDAO struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewGORMMappingDAO 创建GORM映射DAO
func NewGORMMappingDAO(db *gorm.DB, log logger.Logger) *GORMMappingDAO {
	return &GORMMappingDAO{
		db:     db,
		logger: log,
	}
}

// Create 创建服务映射
func (d *GORMMappingDAO) Create(ctx context.Context, mapping *model.ServiceMapping) error {
	hash, err := d.calculateSchemaHash(mapping.JSONSchema)
	if err != nil {
		return fmt.Errorf("calculate schema hash failed: %w", err)
	}
	mapping.SchemaHash = hash

	result := d.db.WithContext(ctx).Create(mapping)
	if result.Error == nil {
		d.logger.Info("Service mapping created",
			logger.String("service_id", strconv.FormatUint(uint64(mapping.ServiceID), 10)),
			logger.String("vauth_key", mapping.VAuthKey),
		)
	}

	return result.Error
}

// Get 获取服务映射
func (d *GORMMappingDAO) Get(ctx context.Context, serviceID uint, vauthKey string) (*model.ServiceMapping, error) {
	var mapping model.ServiceMapping

	result := d.db.WithContext(ctx).Where("service_id = ? AND vauth_key = ? AND enabled = ?", serviceID, vauthKey, true).First(&mapping)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return &mapping, nil
}

// List 获取所有服务映射
func (d *GORMMappingDAO) List(ctx context.Context) ([]*model.ServiceMapping, error) {
	var mappings []*model.ServiceMapping

	result := d.db.WithContext(ctx).Where("enabled = ?", true).Order("updated_at DESC").Find(&mappings)
	if result.Error != nil {
		return nil, result.Error
	}

	return mappings, nil
}

// Update 更新服务映射
func (d *GORMMappingDAO) Update(ctx context.Context, mapping *model.ServiceMapping) error {
	hash, err := d.calculateSchemaHash(mapping.JSONSchema)
	if err != nil {
		return fmt.Errorf("calculate schema hash failed: %w", err)
	}
	mapping.SchemaHash = hash

	result := d.db.WithContext(ctx).Model(&model.ServiceMapping{}).Where("id = ?", mapping.ID).Updates(mapping)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mapping not found: %d", mapping.ID)
	}

	d.logger.Info("Service mapping updated",
		logger.String("service_id", strconv.FormatUint(uint64(mapping.ServiceID), 10)),
		logger.String("vauth_key", mapping.VAuthKey),
	)

	return nil
}

// Delete 删除服务映射
func (d *GORMMappingDAO) Delete(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Delete(&model.ServiceMapping{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mapping not found: %d", id)
	}
	return nil
}

// Validate 验证服务映射是否有效
func (d *GORMMappingDAO) Validate(ctx context.Context, serviceID uint, vauthKey string) (bool, string, error) {
	mapping, err := d.Get(ctx, serviceID, vauthKey)
	if err != nil {
		return false, "", fmt.Errorf("get service mapping failed: %w", err)
	}

	if mapping == nil {
		return false, "service mapping not found", nil
	}

	if len(mapping.JSONSchema) > 0 {
		var schema interface{}
		if err := json.Unmarshal(mapping.JSONSchema, &schema); err != nil {
			return false, fmt.Sprintf("invalid JSON schema: %v", err), nil
		}
	}

	currentHash, err := d.calculateSchemaHash(mapping.JSONSchema)
	if err != nil {
		return false, fmt.Sprintf("calculate current schema hash failed: %v", err), nil
	}

	if currentHash != mapping.SchemaHash {
		return false, "schema has been modified, hash mismatch", nil
	}

	return true, "valid", nil
}

func (d *GORMMappingDAO) calculateSchemaHash(schema json.RawMessage) (string, error) {
	if len(schema) == 0 {
		return "", fmt.Errorf("schema is empty")
	}

	var schemaObj interface{}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		return "", fmt.Errorf("unmarshal schema failed: %w", err)
	}

	normalized, err := json.Marshal(schemaObj)
	if err != nil {
		return "", fmt.Errorf("marshal normalized schema failed: %w", err)
	}

	hash := sha256.Sum256(normalized)
	return hex.EncodeToString(hash[:]), nil
}
