package service

import (
	"context"
	"fmt"
	"time"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/infrastructure/database"

	"gorm.io/gorm"
)

// UserService 用户管理服务
type UserService struct {
	dao *database.GORMUserDAO
}

// NewUserService 创建用户管理服务
func NewUserService(db *gorm.DB, tableName string) *UserService {
	return &UserService{
		dao: database.NewGORMUserDAO(db, tableName),
	}
}

// CreateUser 创建用户（加密密码）
func (s *UserService) CreateUser(ctx context.Context, username, password, name, email, role string) (*model.User, error) {
	// 密码加密
	hashedPassword, err := database.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Name:         name,
		Email:        email,
		Role:         role,
		State:        1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.dao.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// ValidatePassword 验证密码
func (s *UserService) ValidatePassword(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user.State != 1 {
		return nil, fmt.Errorf("user is disabled")
	}

	if !database.ValidatePassword(user.PasswordHash, password) {
		return nil, fmt.Errorf("invalid password")
	}

	// 更新最后登录时间
	s.dao.UpdateLastLogin(ctx, user.ID)

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.dao.GetByUsername(ctx, username)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	return s.dao.GetByID(ctx, id)
}

// ListUsers 列出所有用户
func (s *UserService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.dao.List(ctx)
}

// UpdatePassword 更新密码
func (s *UserService) UpdatePassword(ctx context.Context, id uint, newPassword string) error {
	hashedPassword, err := database.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.dao.UpdatePassword(ctx, id, hashedPassword)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.dao.Delete(ctx, id)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, id uint, name, email, role string, state int) error {
	user, err := s.dao.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Name = name
	user.Email = email
	user.Role = role
	user.State = state
	user.UpdatedAt = time.Now()

	return s.dao.Update(ctx, user)
}
