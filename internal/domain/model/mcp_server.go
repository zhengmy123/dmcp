package model

import (
	"time"
)

type MCPServer struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	VAuthKey       string    `json:"vauth_key" gorm:"column:v_auth_key;size:128;not null;uniqueIndex"`
	Name           string    `json:"name" gorm:"column:name;size:128;not null"`
	Description    string    `json:"description" gorm:"column:description;size:512"`
	Type           string    `json:"type" gorm:"column:type;size:32;not null;default:http_service"`
	HTTPServerURL  string    `json:"http_server_url" gorm:"column:http_server_url;size:512"`
	AuthHeader     string    `json:"auth_header" gorm:"column:auth_header;size:256"`
	TimeoutSeconds int       `json:"timeout_seconds" gorm:"column:timeout_seconds;not null;default:30"`
	ExtraHeaders   string    `json:"extra_headers" gorm:"column:extra_headers;type:text"`
	State          int       `json:"state" gorm:"column:state;default:1;comment:状态 1-正常 0-删除"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (MCPServer) TableName() string {
	return "mcp_servers"
}
