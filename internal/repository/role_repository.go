package repository

import (
	"context"
	"errors"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetByName(ctx context.Context, name string) (*entity.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &role, nil
}
