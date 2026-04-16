package model

import (
	"time"
)

// User 用户信息
type User struct {
	ID           uint      `json:"id"`            // 主键ID
	Username     string    `json:"username" gorm:"size:64;uniqueIndex"` // 用户名
	PasswordHash string    `json:"-"`              // 密码哈希（JSON序列化时忽略）
	Name         string    `json:"name" gorm:"size:128"`         // 显示名称
	Email        string    `json:"email" gorm:"size:256"`         // 邮箱
	Role         string    `json:"role" gorm:"size:32"`          // 角色: admin, user
	Enabled      bool      `json:"enabled"`       // 是否启用
	LastLoginAt  time.Time `json:"last_login_at"` // 最后登录时间
	CreatedAt    time.Time `json:"created_at"`   // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

func (User) TableName() string {
	return "mcp_users"
}

// UserRole 用户角色常量
const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

// NewUser 创建新用户
func NewUser(username, passwordHash, name, email, role string) *User {
	now := time.Now()
	return &User{
		Username:     username,
		PasswordHash: passwordHash,
		Name:         name,
		Email:        email,
		Role:         role,
		Enabled:      true,
		LastLoginAt:  time.Time{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsValid 检查用户信息是否有效
func (u *User) IsValid() bool {
	return u.ID > 0 && u.Username != "" && u.Enabled
}
