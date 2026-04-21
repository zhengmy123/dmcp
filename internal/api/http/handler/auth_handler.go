package handler

import (
	"strconv"
	"time"

	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/infrastructure/auth"
	"dynamic_mcp_go_server/internal/infrastructure/database"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
)

func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	uid, ok := userID.(uint)
	return uid, ok
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	User      *model.User `json:"user"`
}

func LoginHandler(userService *service.UserService, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request: "+err.Error())
			return
		}

		user, err := userService.ValidatePassword(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			response.Unauthorized(c, "invalid username or password")
			return
		}

		token, err := jwtManager.GenerateToken(user)
		if err != nil {
			response.InternalError(c, "failed to generate token")
			return
		}

		response.Success(c, LoginResponse{
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			User: &model.User{
				ID:          user.ID,
				Username:    user.Username,
				Name:        user.Name,
				Email:       user.Email,
				Role:        user.Role,
				State:       user.State,
				LastLoginAt: user.LastLoginAt,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			},
		})
	}
}

func GetCurrentUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			response.Unauthorized(c, "unauthorized")
			return
		}

		user, err := userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			response.NotFound(c, "user not found")
			return
		}

		response.Success(c, gin.H{
			"user": &model.User{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
				Email:    user.Email,
				Role:     user.Role,
				State:    user.State,
			},
		})
	}
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func ChangePasswordHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			response.Unauthorized(c, "unauthorized")
			return
		}

		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request: "+err.Error())
			return
		}

		user, err := userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			response.NotFound(c, "user not found")
			return
		}

		if !database.ValidatePassword(user.PasswordHash, req.OldPassword) {
			response.BadRequest(c, "invalid old password")
			return
		}

		err = userService.UpdatePassword(c.Request.Context(), user.ID, req.NewPassword)
		if err != nil {
			response.InternalError(c, "failed to update password")
			return
		}

		response.SuccessWithMessage(c, "password updated successfully", nil)
	}
}

func ListUsersHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := userService.ListUsers(c.Request.Context())
		if err != nil {
			response.InternalError(c, "failed to list users")
			return
		}

		safeUsers := make([]*model.User, len(users))
		for i, u := range users {
			safeUsers[i] = &model.User{
				ID:          u.ID,
				Username:    u.Username,
				Name:        u.Name,
				Email:       u.Email,
				Role:        u.Role,
				State:       u.State,
				LastLoginAt: u.LastLoginAt,
				CreatedAt:   u.CreatedAt,
				UpdatedAt:   u.UpdatedAt,
			}
		}

		response.Success(c, gin.H{
			"users": safeUsers,
			"count": len(safeUsers),
		})
	}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

func CreateUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request: "+err.Error())
			return
		}

		user, err := userService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Name, req.Email, req.Role)
		if err != nil {
			response.InternalError(c, "failed to create user: "+err.Error())
			return
		}

		response.Created(c, gin.H{
			"user": &model.User{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
				Email:    user.Email,
				Role:     user.Role,
				State:    user.State,
			},
		})
	}
}

type UpdateUserRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role" binding:"omitempty,oneof=admin user"`
	State   *int   `json:"state"`
}

func UpdateUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid user id")
			return
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request: "+err.Error())
			return
		}

		user, err := userService.GetUserByID(c.Request.Context(), uint(userID))
		if err != nil {
			response.NotFound(c, "user not found")
			return
		}

		if req.Name != "" {
			user.Name = req.Name
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		if req.Role != "" {
			user.Role = req.Role
		}
		if req.State != nil {
			user.State = *req.State
		}

		err = userService.UpdateUser(c.Request.Context(), uint(userID), user.Name, user.Email, user.Role, user.State)
		if err != nil {
			response.InternalError(c, "failed to update user")
			return
		}

		response.SuccessWithMessage(c, "user updated successfully", nil)
	}
}

func DeleteUserHandler(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid user id")
			return
		}

		err = userService.DeleteUser(c.Request.Context(), uint(userID))
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}

		response.SuccessWithMessage(c, "user deleted successfully", nil)
	}
}