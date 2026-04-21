package main

import (
	"fmt"
	"log"
	"time"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/infrastructure/database"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 连接数据库
	dsn := "root:1234qwer@tcp(127.0.0.1:3306)/mcp_server?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移表结构
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}

	// 密码加密
	password := "admin123"
	hashedPassword, err := database.HashPassword(password)
	if err != nil {
		log.Fatalf("密码加密失败: %v", err)
	}

	now := time.Now()

	// 创建用户
	user := &model.User{
		Username:     "admin",
		PasswordHash: hashedPassword,
		Name:         "Administrator",
		Email:        "admin@example.com",
		Role:         model.UserRoleAdmin,
		State:        1,
		LastLoginAt:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 检查用户是否已存在
	var existing model.User
	result := db.Where("username = ?", user.Username).First(&existing)
	if result.Error == nil {
		fmt.Printf("用户 '%s' 已存在，更新密码...\n", user.Username)
		user.ID = existing.ID
		result = db.Save(user)
	} else {
		result = db.Create(user)
	}

	if result.Error != nil {
		log.Fatalf("创建用户失败: %v", result.Error)
	}

	fmt.Printf("✓ 用户 '%s' 创建成功!\n", user.Username)
	fmt.Printf("  密码: %s\n", password)
}
