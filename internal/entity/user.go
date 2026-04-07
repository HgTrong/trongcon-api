package entity

import "time"

const (
	AccountFree    = "free"
	AccountPremium = "premium"
)

type User struct {
	BaseEntity
	Email        string `json:"email" gorm:"type:varchar(255);not null;uniqueIndex"`
	Name         string `json:"name" gorm:"type:varchar(255)"`
	FirstName    string `json:"first_name" gorm:"type:varchar(128)"`
	LastName     string `json:"last_name" gorm:"type:varchar(128)"`
	Gender       string `json:"gender" gorm:"type:varchar(32)"`
	Language     string `json:"language" gorm:"type:varchar(16);default:'en'"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	AccountType  string `json:"account_type" gorm:"type:varchar(32);default:'free'"`
	PasswordHash string `json:"-" gorm:"type:varchar(255);not null"`
	Roles        []Role `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}
