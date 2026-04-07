package repository

import (
	"context"
	"errors"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, u *entity.User) error
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, u *entity.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order string) ([]entity.User, int64, error)
	UpdateLastLoginAt(ctx context.Context, id uint, t time.Time) error
	AppendRole(ctx context.Context, u *entity.User, role *entity.Role) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *entity.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, u *entity.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *userRepository) List(ctx context.Context, offset, limit int, order string) ([]entity.User, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.User
	if order == "" {
		order = "id ASC"
	}
	if err := r.db.WithContext(ctx).Preload("Roles").Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *userRepository) UpdateLastLoginAt(ctx context.Context, id uint, t time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("last_login_at", t).Error
}

func (r *userRepository) AppendRole(ctx context.Context, u *entity.User, role *entity.Role) error {
	return r.db.WithContext(ctx).Model(u).Association("Roles").Append(role)
}
