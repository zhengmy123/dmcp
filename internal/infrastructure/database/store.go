package database

import (
	"context"
	"fmt"
	"sort"

	"dynamic_mcp_go_server/internal/domain/model"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"
)

// ToolDAO 工具定义数据访问接口
type ToolDAO interface {
	List(ctx context.Context) ([]model.ToolDefinition, error)
	ListByServerID(ctx context.Context, serverID uint) ([]model.ToolDefinition, error)
}

// toolRow 工具定义行
type toolRow struct {
	ID          uint   `gorm:"column:id"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Parameters  []byte `gorm:"column:parameters"`
}

// GORMToolDAO GORM实现的ToolDAO
type GORMToolDAO struct {
	db    *gorm.DB
	table string
}

// NewGORMToolDAO 创建GORM工具DAO
func NewGORMToolDAO(db *gorm.DB, table string) *GORMToolDAO {
	return &GORMToolDAO{
		db:    db,
		table: table,
	}
}

// List 获取所有启用的工具定义
func (d *GORMToolDAO) List(ctx context.Context) ([]model.ToolDefinition, error) {
	var rows []toolRow
	result := d.db.WithContext(ctx).Table(d.table).
		Select("name, description, parameters").
		Where("state = ?", 1).
		Order("updated_at DESC").
		Find(&rows)
	if result.Error != nil {
		return nil, result.Error
	}

	return scanToolRows(rows)
}

// ListByServerID 根据服务器ID获取工具定义
func (d *GORMToolDAO) ListByServerID(ctx context.Context, serverID uint) ([]model.ToolDefinition, error) {
	var rows []toolRow
	result := d.db.WithContext(ctx).Table(d.table).
		Select("name, description, parameters").
		Where("service_id = ? AND state = ?", serverID, 1).
		Order("updated_at DESC").
		Find(&rows)
	if result.Error != nil {
		return nil, result.Error
	}

	return scanToolRows(rows)
}

// scanToolRows 扫描工具定义行
func scanToolRows(rows []toolRow) ([]model.ToolDefinition, error) {
	defs := make([]model.ToolDefinition, 0, len(rows))
	for _, r := range rows {
		defs = append(defs, model.ToolDefinition{
			Name:        r.Name,
			Description: r.Description,
			Parameters:  r.Parameters,
			State:       1,
		})
	}
	return defs, nil
}

// parseParametersJSON 解析 parameters_json（JSON Schema object 格式）
// 格式: {"type":"object","properties":{...},"required":[...]}
func parseParametersJSON(data []byte) ([]model.ParameterDefinition, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var schema jsonSchemaObject
	if err := sonic.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("invalid JSON Schema: %w", err)
	}

	return schemaToParams(schema), nil
}

// jsonSchemaObject 表示 JSON Schema 的 object 格式
type jsonSchemaObject struct {
	Type       string                       `json:"type"`
	Properties map[string]jsonSchemaProperty `json:"properties"`
	Required   []string                     `json:"required"`
}

// jsonSchemaProperty 表示 JSON Schema 中的单个属性
type jsonSchemaProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Default     any      `json:"default"`
	Enum        []any    `json:"enum"`
	Minimum     *float64 `json:"minimum"`
	Maximum     *float64 `json:"maximum"`
}

// schemaToParams 将 JSON Schema object 转换为 []ParameterDefinition
func schemaToParams(schema jsonSchemaObject) []model.ParameterDefinition {
	if len(schema.Properties) == 0 {
		return nil
	}

	requiredSet := make(map[string]bool, len(schema.Required))
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	params := make([]model.ParameterDefinition, 0, len(schema.Properties))
	for name, prop := range schema.Properties {
		pd := model.ParameterDefinition{
			Name:        name,
			Type:        model.ParameterType(prop.Type),
			Required:    requiredSet[name],
			Description: prop.Description,
			Default:     prop.Default,
			Enum:        prop.Enum,
			Minimum:     prop.Minimum,
			Maximum:     prop.Maximum,
		}
		params = append(params, pd)
	}

	// 按名称排序保证顺序稳定
	sort.Slice(params, func(i, j int) bool {
		// required 排前面
		if params[i].Required != params[j].Required {
			return params[i].Required
		}
		return params[i].Name < params[j].Name
	})

	return params
}
