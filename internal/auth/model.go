package auth

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken 刷新token模型
type RefreshToken struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Token     string    `gorm:"uniqueIndex;not null;size:255"` // refresh token
	UserID    uint      `gorm:"not null;index"`                // 用户ID
	ExpiresAt time.Time `gorm:"not null;index"`                // 过期时间
	IsRevoked bool      `gorm:"default:false"`                 // 是否已撤销
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
