package user

import (
	"gorm.io/gorm"
)

type Repository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
	List(offset, limit int) ([]*User, int64, error)

	// UserProfile 相关方法
	GetProfileByUserID(userID uint) (*UserProfile, error)
	CreateOrUpdateProfile(profile *UserProfile) error
	DeleteProfile(userID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *repository) GetByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetByUsername(username string) (*User, error) {
	var user User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Update(user *User) error {
	return r.db.Save(user).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

func (r *repository) List(offset, limit int) ([]*User, int64, error) {
	var users []*User
	var total int64

	if err := r.db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetProfileByUserID 根据用户ID获取用户配置
func (r *repository) GetProfileByUserID(userID uint) (*UserProfile, error) {
	var profile UserProfile
	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// CreateOrUpdateProfile 创建或更新用户配置
func (r *repository) CreateOrUpdateProfile(profile *UserProfile) error {
	var existing UserProfile
	err := r.db.Where("user_id = ?", profile.UserID).First(&existing).Error

	if err != nil {
		// 不存在则创建
		return r.db.Create(profile).Error
	}

	// 存在则更新
	profile.ID = existing.ID
	return r.db.Save(profile).Error
}

// DeleteProfile 删除用户配置
func (r *repository) DeleteProfile(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&UserProfile{}).Error
}
