package model

import (
	"time"
)

type TokenServerBinding struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TokenID   uint      `json:"token_id" gorm:"not null;index"`
	ServerID  uint      `json:"server_id" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (TokenServerBinding) TableName() string {
	return "token_mcp_server_bindings"
}
