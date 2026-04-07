package entity

import (
	"time"

	"gorm.io/gorm"
)

// BaseEntity giống pattern strongbody-api (id, timestamp, soft delete).
type BaseEntity struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

func (b *BaseEntity) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}
