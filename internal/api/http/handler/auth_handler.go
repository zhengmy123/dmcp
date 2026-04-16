package handler

import (
	"net/http"
	"strconv"
	"time"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/infrastructure/auth"
	"dynamic_mcp_go_server/internal/infrastructure/database"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	uid, ok := userID.(uint)
	return uid, ok
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	User      *model.User `json:"user"`
}

// LoginHandler 处理登录请求
func LoginHandler(userService *service.UserService, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		// 验证用户
		user, err := userService.ValidatePassword(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		// 生成JWT token
		token, err := jwtManager.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, LoginResponse{
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			User: &model.User{
				ID:          user.ID,
				Username:    user.Username,
				Name:        user.Name,
				Email:       user.Email,
				Role:        user.Role,
				Enabled:     user.Enabled,
				LastLoginAt: user.LastLoginAt,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			},
		})
	}
}

// GetCurrentUserHandler 获取当前用户信息
func GetCurrentUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": &model.User{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
				Email:    user.Email,
				Role:     user.Role,
				Enabled:  user.Enabled,
			},
		})
	}
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePasswordHandler 修改密码
func ChangePasswordHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		// 获取当前用户验证旧密码
		user, err := userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		if !database.ValidatePassword(user.PasswordHash, req.OldPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid old password"})
			return
		}

		// 更新密码
		err = userService.UpdatePassword(c.Request.Context(), user.ID, req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
	}
}

// ListUsersHandler 列出所有用户
func ListUsersHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := userService.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
			return
		}

		// 隐藏密码
		safeUsers := make([]*model.User, len(users))
		for i, u := range users {
			safeUsers[i] = &model.User{
				ID:          u.ID,
				Username:    u.Username,
				Name:        u.Name,
				Email:       u.Email,
				Role:        u.Role,
				Enabled:     u.Enabled,
				LastLoginAt: u.LastLoginAt,
				CreatedAt:   u.CreatedAt,
				UpdatedAt:   u.UpdatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"users": safeUsers,
			"count": len(safeUsers),
		})
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// CreateUserHandler 创建用户
func CreateUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		user, err := userService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Name, req.Email, req.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "user created successfully",
			"user": &model.User{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
				Email:    user.Email,
				Role:     user.Role,
				Enabled:  user.Enabled,
			},
		})
	}
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role" binding:"omitempty,oneof=admin user"`
	Enabled *bool  `json:"enabled"`
}

// UpdateUserHandler 更新用户
func UpdateUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		// 获取当前用户信息
		user, err := userService.GetUserByID(c.Request.Context(), uint(userID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// 更新字段
		if req.Name != "" {
			user.Name = req.Name
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		if req.Role != "" {
			user.Role = req.Role
		}
		if req.Enabled != nil {
			user.Enabled = *req.Enabled
		}

		err = userService.UpdateUser(c.Request.Context(), uint(userID), user.Name, user.Email, user.Role, user.Enabled)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
	}
}

// DeleteUserHandler 删除用户
func DeleteUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		err = userService.DeleteUser(c.Request.Context(), uint(userID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
	}
}
