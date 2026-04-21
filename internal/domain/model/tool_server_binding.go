package model

import (
	"time"
)

type ToolServerBinding struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ToolID    uint      `json:"tool_id" gorm:"not null;index:idx_tool_server,unique"`
	ServerID  uint      `json:"server_id" gorm:"not null;index:idx_tool_server,unique"`
	State     int       `json:"state" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ToolServerBinding) TableName() string {
	return "tool_mcp_server_bindings"
}
