package model

import (
	"time"
)

type MCPServer struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	VAuthKey       string    `json:"vauth_key" gorm:"size:128;not null;uniqueIndex"`
	Name           string    `json:"name" gorm:"size:128;not null"`
	Description    string    `json:"description" gorm:"size:512"`
	Type           string    `json:"type" gorm:"size:32;not null;default:http_service"`
	HTTPServerURL  string    `json:"http_server_url" gorm:"size:512"`
	AuthHeader     string    `json:"auth_header" gorm:"size:256"`
	TimeoutSeconds int       `json:"timeout_seconds" gorm:"not null;default:30"`
	ExtraHeaders   string    `json:"extra_headers" gorm:"type:text"`
	State          int       `json:"state" gorm:"default:1;comment:状态 1-正常 0-删除"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (MCPServer) TableName() string {
	return "mcp_servers"
}
