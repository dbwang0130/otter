package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Username  string `json:"username" gorm:"not null;size:100"`
	Email     string `json:"email" gorm:"not null;size:255"`
	Password  string `json:"-" gorm:"not null;size:255"` // 密码不序列化到JSON
	FirstName string `json:"first_name" gorm:"size:100"`
	LastName  string `json:"last_name" gorm:"size:100"`
	Phone     string `json:"phone" gorm:"size:20"`
	Status    string `json:"status" gorm:"default:active;size:20"`
	IsAdmin   bool   `json:"is_admin" gorm:"default:false"` // 管理员标识
}

func (User) TableName() string {
	return "users"
}
