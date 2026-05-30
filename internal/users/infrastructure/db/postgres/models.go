package postgres

import (
	"time"

	"gorm.io/gorm"
)

type userModel struct {
	ID              string         `gorm:"column:id;primaryKey"`
	Name            string         `gorm:"column:name;not null"`
	Email           string         `gorm:"column:email;not null;uniqueIndex"`
	PasswordHash    *string        `gorm:"column:password_hash"`
	EmailVerifiedAt *time.Time     `gorm:"column:email_verified_at"`
	AvatarURL       *string        `gorm:"column:avatar_url"`
	LastLoginAt     *time.Time     `gorm:"column:last_login_at"`
	Timezone        *string        `gorm:"column:timezone"`
	Locale          string         `gorm:"column:locale;not null;default:en"`
	CreatedAt       time.Time      `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (userModel) TableName() string {
	return "users"
}
