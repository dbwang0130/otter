package auth

import (
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	GetByToken(token string) (*RefreshToken, error)
	RevokeByToken(token string) error
	RevokeByUserID(userID uint) error
	DeleteExpired() error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) GetByToken(token string) (*RefreshToken, error) {
	var refreshToken RefreshToken
	if err := r.db.Where("token = ? AND is_revoked = ?", token, false).First(&refreshToken).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) RevokeByToken(token string) error {
	return r.db.Model(&RefreshToken{}).Where("token = ?", token).Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) RevokeByUserID(userID uint) error {
	return r.db.Model(&RefreshToken{}).Where("user_id = ?", userID).Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", "now()").Delete(&RefreshToken{}).Error
}
