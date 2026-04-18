package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthKey 认证密钥配置
type AuthKey struct {
	ID        uint    `json:"id"`                                          // 主键ID
	KeyID     string  `json:"key_id"`                                      // 访问密钥ID
	Token     string  `json:"token"`                                       // 访问令牌(Token)
	Secret    string  `json:"secret"`                                      // 密钥Secret
	Name      string  `json:"name"`                                        // 密钥名称/描述
	State     int     `json:"state" gorm:"default:1;comment:状态 1-正常 0-删除"` // 状态
	LastUsed  *string `json:"last_used"`                                   // 最后使用时间
	ExpiresAt string  `json:"expires_at"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// AuthService 认证管理服务
type AuthService struct {
	adminToken string              // 后台管理 API Token
	authKeys   map[string]*AuthKey // token -> config
	mu         sync.RWMutex
	db         *gorm.DB
	tableName  string
	stopCh     chan struct{}
	refreshInt time.Duration
}

// NewAuthService 创建认证管理服务
func NewAuthService(adminToken string) *AuthService {
	return &AuthService{
		adminToken: adminToken,
		authKeys:   make(map[string]*AuthKey),
		stopCh:     make(chan struct{}),
		refreshInt: 5 * time.Minute,
	}
}

// InitWithGORM 初始化GORM存储
func (s *AuthService) InitWithGORM(db *gorm.DB, tableName string) {
	s.db = db
	s.tableName = tableName
}

// StartTokenRefresher 启动Token刷新协程
func (s *AuthService) StartTokenRefresher(ctx context.Context) {
	if s.db == nil {
		return
	}
	go s.refreshTokens(ctx)
}

func (s *AuthService) refreshTokens(ctx context.Context) {
	ticker := time.NewTicker(s.refreshInt)
	defer ticker.Stop()

	// 初始加载
	if err := s.loadTokensFromDB(); err != nil {
		fmt.Printf("initial token load failed: %v\n", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.loadTokensFromDB(); err != nil {
				fmt.Printf("refresh tokens failed: %v\n", err)
			}
		}
	}
}

// authKeyRow GORM扫描用的内部结构
type authKeyRow struct {
	ID        uint    `gorm:"column:id"`
	KeyID     string  `gorm:"column:key_id"`
	Token     string  `gorm:"column:token"`
	Secret    string  `gorm:"column:secret"`
	Name      *string `gorm:"column:name"`
	State     int     `gorm:"column:state"`
	LastUsed  *string `gorm:"column:last_used_at"`
	ExpiresAt *string `gorm:"column:expires_at"`
	CreatedAt *string `gorm:"column:created_at"`
	UpdatedAt *string `gorm:"column:updated_at"`
}

func (s *AuthService) loadTokensFromDB() error {
	if s.db == nil {
		return nil
	}

	var rows []authKeyRow
	result := s.db.Table(s.tableName).Where("state = ?", 1).Find(&rows)
	if result.Error != nil {
		return result.Error
	}

	newKeys := make(map[string]*AuthKey)
	for _, r := range rows {
		key := &AuthKey{
			ID:     r.ID,
			KeyID:  r.KeyID,
			Token:  r.Token,
			Secret: r.Secret,
			State:  r.State,
		}
		if r.Name != nil {
			key.Name = *r.Name
		}
		if r.LastUsed != nil {
			key.LastUsed = r.LastUsed
		}
		if r.ExpiresAt != nil {
			key.ExpiresAt = *r.ExpiresAt
		}
		if r.CreatedAt != nil {
			key.CreatedAt = *r.CreatedAt
		}
		if r.UpdatedAt != nil {
			key.UpdatedAt = *r.UpdatedAt
		}

		newKeys[key.Token] = key
	}

	s.mu.Lock()
	s.authKeys = newKeys
	s.mu.Unlock()

	return nil
}

// ValidateToken 验证Token，返回 (valid, expired)
func (s *AuthService) ValidateToken(token string) (bool, bool) {
	s.mu.RLock()
	key, exists := s.authKeys[token]
	s.mu.RUnlock()

	if !exists {
		return false, false
	}

	if key.State != 1 {
		return false, false
	}

	// 检查是否过期
	if key.ExpiresAt != "" {
		expTime, err := time.Parse(time.RFC3339, key.ExpiresAt)
		if err == nil && time.Now().After(expTime) {
			return false, true
		}
	}

	// 更新最后使用时间（异步）
	go s.updateLastUsed(key.Token)

	return true, false
}

// updateLastUsed 更新最后使用时间
func (s *AuthService) updateLastUsed(token string) {
	if s.db == nil {
		return
	}
	s.db.Table(s.tableName).Where("token = ?", token).Update("last_used_at", gorm.Expr("NOW()"))
}

// RegisterToken 注册 Token
func (s *AuthService) RegisterToken(ctx context.Context, keyID, token, secret, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if token == "" || secret == "" {
		return fmt.Errorf("token and secret are required")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	key := &AuthKey{
		KeyID:     keyID,
		Token:     token,
		Secret:    secret,
		Name:      name,
		State:     1,
		CreatedAt: nowStr,
		UpdatedAt: nowStr,
	}

	s.authKeys[token] = key

	// 如果有DB，也写入数据库
	if s.db != nil {
		return s.saveTokenToDB(ctx, key)
	}
	return nil
}

// saveTokenToDB 保存Token到数据库
func (s *AuthService) saveTokenToDB(ctx context.Context, key *AuthKey) error {
	row := authKeyRow{
		KeyID:  key.KeyID,
		Token:  key.Token,
		Secret: key.Secret,
		Name:   &key.Name,
		State:  key.State,
	}

	// 先查询记录是否存在
	var existing authKeyRow
	result := s.db.WithContext(ctx).Table(s.tableName).Where("id = ?", key.ID).First(&existing)

	if result.Error == nil {
		// 记录存在，执行更新
		row.ID = existing.ID
		return s.db.WithContext(ctx).Table(s.tableName).
			Where("id = ?", key.ID).
			Updates(map[string]interface{}{
				"key_id":     row.KeyID,
				"token":      row.Token,
				"secret":     row.Secret,
				"name":       row.Name,
				"state":      row.State,
				"updated_at": gorm.Expr("NOW()"),
			}).Error
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 记录不存在，执行创建
		return s.db.WithContext(ctx).Table(s.tableName).Create(&row).Error
	}

	return result.Error
}

// GetToken 获取 Token 配置
func (s *AuthService) GetToken(token string) (*AuthKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.authKeys[token]
	if !ok {
		return nil, false
	}
	return key, true
}

// ListTokens 列出所有 Token
func (s *AuthService) ListTokens(ctx context.Context) []*AuthKey {
	// 如果有数据库，直接从数据库读取所有 token
	if s.db != nil {
		return s.listTokensFromDB(ctx)
	}

	// 否则从内存缓存读取
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AuthKey, 0, len(s.authKeys))
	for _, key := range s.authKeys {
		result = append(result, key)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].KeyID < result[j].KeyID
	})
	return result
}

// listTokensFromDB 从数据库读取所有 Token
func (s *AuthService) listTokensFromDB(ctx context.Context) []*AuthKey {
	var rows []authKeyRow
	result := s.db.WithContext(ctx).Table(s.tableName).Where("state = ?", 1).Find(&rows)
	if result.Error != nil {
		return nil
	}

	var authKeys []*AuthKey
	for _, r := range rows {
		key := &AuthKey{
			ID:     r.ID,
			KeyID:  r.KeyID,
			Token:  r.Token,
			Secret: r.Secret,
			State:  r.State,
		}
		if r.Name != nil {
			key.Name = *r.Name
		}
		if r.LastUsed != nil {
			key.LastUsed = r.LastUsed
		}
		if r.ExpiresAt != nil {
			key.ExpiresAt = *r.ExpiresAt
		}
		if r.CreatedAt != nil {
			key.CreatedAt = *r.CreatedAt
		}
		if r.UpdatedAt != nil {
			key.UpdatedAt = *r.UpdatedAt
		}

		authKeys = append(authKeys, key)
	}

	sort.Slice(authKeys, func(i, j int) bool {
		return authKeys[i].KeyID < authKeys[j].KeyID
	})
	return authKeys
}

// DeleteToken 删除 Token
func (s *AuthService) DeleteToken(ctx context.Context, token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.authKeys[token]; !ok {
		return false
	}
	delete(s.authKeys, token)

	if s.db != nil {
		s.db.WithContext(ctx).Table(s.tableName).Where("token = ?", token).Delete(nil)
	}
	return true
}

// DisableToken 禁用 Token（软删除）
func (s *AuthService) DisableToken(ctx context.Context, token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, ok := s.authKeys[token]
	if !ok {
		return false
	}
	key.State = 0
	return true
}

// EnableToken 启用 Token
func (s *AuthService) EnableToken(ctx context.Context, token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, ok := s.authKeys[token]
	if !ok {
		return false
	}
	key.State = 1
	return true
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(ctx context.Context, token string) (newToken, newSecret string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, ok := s.authKeys[token]
	if !ok {
		return "", "", fmt.Errorf("token not found")
	}

	newToken = uuid.New().String()
	newSecret = uuid.New().String() + uuid.New().String()

	key.Token = newToken
	key.Secret = newSecret
	key.UpdatedAt = time.Now().Format(time.RFC3339)

	if s.db != nil {
		result := s.db.WithContext(ctx).Table(s.tableName).Where("id = ?", key.ID).Updates(map[string]interface{}{
			"token":      newToken,
			"secret":     newSecret,
			"updated_at": gorm.Expr("NOW()"),
		})
		if result.Error != nil {
			return "", "", result.Error
		}
	}

	return newToken, newSecret, nil
}

// VerifySignature 验证签名
func (s *AuthService) VerifySignature(token, timestamp, signature string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.authKeys[token]
	if !ok || key.State != 1 {
		return false
	}

	// 验证时间戳
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		var unixTs int64
		if _, err := fmt.Sscanf(timestamp, "%d", &unixTs); err != nil {
			return false
		}
		ts = time.Unix(unixTs, 0)
	}

	now := time.Now()
	if ts.Before(now.Add(-5*time.Minute)) || ts.After(now.Add(5*time.Minute)) {
		return false
	}

	// 计算签名
	signStr := buildAuthServiceSignString("", "", timestamp)
	expectedSig := calcAuthServiceHMACSHA256(key.Secret, signStr)
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

func buildAuthServiceSignString(method, path, timestamp string) string {
	return strings.ToUpper(method) + "\n" + path + "\n" + timestamp
}

func calcAuthServiceHMACSHA256(secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HasAuthKeys 检查是否有配置的认证密钥
func (s *AuthService) HasAuthKeys() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.authKeys) > 0
}

// HasDB 检查是否配置了数据库
func (s *AuthService) HasDB() bool {
	return s.db != nil
}

// AdminToken 获取管理Token
func (s *AuthService) AdminToken() string {
	return s.adminToken
}
