package database

import (
	"context"
	"fmt"
	"time"

	"dynamic_mcp_go_server/internal/domain/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserDAO 用户数据访问接口
type UserDAO interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	UpdatePassword(ctx context.Context, id uint, newPasswordHash string) error
	UpdateLastLogin(ctx context.Context, id uint) error
}

// GORMUserDAO GORM实现的UserDAO
type GORMUserDAO struct {
	db        *gorm.DB
	tableName string
}

// NewGORMUserDAO 创建GORM用户DAO
func NewGORMUserDAO(db *gorm.DB, tableName string) *GORMUserDAO {
	return &GORMUserDAO{
		db:        db,
		tableName: tableName,
	}
}

// Create 创建用户
func (d *GORMUserDAO) Create(ctx context.Context, user *model.User) error {
	result := d.db.WithContext(ctx).Table(d.tableName).Create(user)
	return result.Error
}

// GetByID 根据ID获取用户
func (d *GORMUserDAO) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	result := d.db.WithContext(ctx).Table(d.tableName).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (d *GORMUserDAO) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	result := d.db.WithContext(ctx).Table(d.tableName).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// List 获取所有用户
func (d *GORMUserDAO) List(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	result := d.db.WithContext(ctx).Table(d.tableName).Order("created_at DESC").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// Update 更新用户
func (d *GORMUserDAO) Update(ctx context.Context, user *model.User) error {
	result := d.db.WithContext(ctx).Table(d.tableName).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"state":      user.State,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %d", user.ID)
	}
	return nil
}

// Delete 删除用户
func (d *GORMUserDAO) Delete(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Table(d.tableName).Where("id = ?", id).Delete(nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %d", id)
	}
	return nil
}

// UpdatePassword 更新密码
func (d *GORMUserDAO) UpdatePassword(ctx context.Context, id uint, newPasswordHash string) error {
	result := d.db.WithContext(ctx).Table(d.tableName).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash": newPasswordHash,
		"updated_at":    time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %d", id)
	}
	return nil
}

// UpdateLastLogin 更新最后登录时间
func (d *GORMUserDAO) UpdateLastLogin(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Table(d.tableName).Where("id = ?", id).Update("last_login_at", time.Now())
	return result.Error
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ValidatePassword 验证密码
func ValidatePassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
