package database

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
)

// GORMServiceDAO GORM实现的服务DAO
type GORMServiceDAO struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewGORMServiceDAO 创建GORM服务DAO
func NewGORMServiceDAO(db *gorm.DB, log logger.Logger) *GORMServiceDAO {
	return &GORMServiceDAO{
		db:     db,
		logger: log,
	}
}

// ListWithQuery 获取服务列表（支持分页和搜索）
func (d *GORMServiceDAO) ListWithQuery(ctx context.Context, query *model.ServiceQuery) ([]*model.HTTPService, int64, error) {
	var services []*model.HTTPService
	var total int64

	db := d.db.WithContext(ctx).Model(&model.HTTPService{})

	// 名称模糊匹配
	if query.Name != nil && *query.Name != "" {
		db = db.Where("name LIKE ?", "%"+*query.Name+"%")
	}

	// 状态筛选
	if query.State != nil {
		db = db.Where("state = ?", *query.State)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 默认分页参数
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	// 分页查询
	offset := (page - 1) * pageSize
	result := db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&services)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	for _, s := range services {
		d.unmarshalJSONFields(s)
	}

	return services, total, nil
}

// List 获取所有启用的服务
func (d *GORMServiceDAO) List(ctx context.Context) ([]*model.HTTPService, error) {
	var services []*model.HTTPService

	result := d.db.WithContext(ctx).Where("state = ?", 1).Find(&services)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, s := range services {
		d.unmarshalJSONFields(s)
	}

	return services, nil
}

// Get 根据ID获取服务
func (d *GORMServiceDAO) Get(ctx context.Context, id uint) (*model.HTTPService, error) {
	var service model.HTTPService

	result := d.db.WithContext(ctx).Where("id = ? AND state = ?", id, 1).First(&service)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found")
		}
		return nil, result.Error
	}

	d.unmarshalJSONFields(&service)

	return &service, nil
}

// Save 保存服务（创建或更新）
func (d *GORMServiceDAO) Save(ctx context.Context, service *model.HTTPService) error {
	if service.Headers != nil {
		headersJSON, err := sonic.Marshal(service.Headers)
		if err != nil {
			return fmt.Errorf("marshal headers failed: %w", err)
		}
		service.HeadersJSON = string(headersJSON)
	}

	// 如果没有ID，创建新记录
	if service.ID == 0 {
		return d.db.WithContext(ctx).Create(service).Error
	}

	var existing model.HTTPService
	result := d.db.WithContext(ctx).Where("id = ?", service.ID).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return d.db.WithContext(ctx).Create(service).Error
		}
		return result.Error
	}

	return d.db.WithContext(ctx).Model(&model.HTTPService{}).Where("id = ?", service.ID).Updates(service).Error
}

// Delete 删除服务
func (d *GORMServiceDAO) Delete(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Delete(&model.HTTPService{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete service failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("service not found: %d", id)
	}
	return nil
}

// unmarshalJSONFields 反序列化服务对象中的JSON字段
func (d *GORMServiceDAO) unmarshalJSONFields(s *model.HTTPService) {
	if s.HeadersJSON != "" {
		if err := sonic.Unmarshal([]byte(s.HeadersJSON), &s.Headers); err != nil {
			d.logger.Warn("unmarshal headers failed", logger.Error(err))
		}
	}
	// InputSchema 和 OutputSchema 由 GORM 自动映射 (json.RawMessage -> text)
}
