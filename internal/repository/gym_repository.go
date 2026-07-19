package repository

import (
	"context"
	"errors"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type GymBranchRepository interface {
	Create(ctx context.Context, b *entity.GymBranch) error
	GetByID(ctx context.Context, id uint) (*entity.GymBranch, error)
	GetBySlug(ctx context.Context, slug string) (*entity.GymBranch, error)
	SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error)
	Update(ctx context.Context, b *entity.GymBranch) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q, city string, active *bool) ([]entity.GymBranch, int64, error)
}

type gymBranchRepository struct {
	db *gorm.DB
}

func NewGymBranchRepository(db *gorm.DB) GymBranchRepository {
	return &gymBranchRepository{db: db}
}

func (r *gymBranchRepository) Create(ctx context.Context, b *entity.GymBranch) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *gymBranchRepository) GetByID(ctx context.Context, id uint) (*entity.GymBranch, error) {
	var b entity.GymBranch
	if err := r.db.WithContext(ctx).First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gymBranchRepository) GetBySlug(ctx context.Context, slug string) (*entity.GymBranch, error) {
	var b entity.GymBranch
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gymBranchRepository) SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	q := r.db.WithContext(ctx).Model(&entity.GymBranch{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *gymBranchRepository) Update(ctx context.Context, b *entity.GymBranch) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *gymBranchRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.GymBranch{}, id).Error
}

func (r *gymBranchRepository) List(ctx context.Context, offset, limit int, order, q, city string, active *bool) ([]entity.GymBranch, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.GymBranch{})
	if q != "" {
		like := "%" + strings.TrimSpace(q) + "%"
		query = query.Where("name ILIKE ? OR address ILIKE ? OR city ILIKE ?", like, like, like)
	}
	if city != "" {
		query = query.Where("city ILIKE ?", "%"+strings.TrimSpace(city)+"%")
	}
	if active != nil {
		query = query.Where("is_active = ?", *active)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "sort_order ASC, id ASC"
	}
	var list []entity.GymBranch
	if err := query.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

type TrainerProfileRepository interface {
	Create(ctx context.Context, t *entity.TrainerProfile) error
	GetByID(ctx context.Context, id uint) (*entity.TrainerProfile, error)
	GetByUserID(ctx context.Context, userID uint) (*entity.TrainerProfile, error)
	Update(ctx context.Context, t *entity.TrainerProfile) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q string, branchID *uint, isPublic *bool) ([]entity.TrainerProfile, int64, error)
}

type trainerProfileRepository struct {
	db *gorm.DB
}

func NewTrainerProfileRepository(db *gorm.DB) TrainerProfileRepository {
	return &trainerProfileRepository{db: db}
}

func trainerPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("User").Preload("Branch")
}

func (r *trainerProfileRepository) Create(ctx context.Context, t *entity.TrainerProfile) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *trainerProfileRepository) GetByID(ctx context.Context, id uint) (*entity.TrainerProfile, error) {
	var t entity.TrainerProfile
	if err := trainerPreload(r.db.WithContext(ctx)).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *trainerProfileRepository) GetByUserID(ctx context.Context, userID uint) (*entity.TrainerProfile, error) {
	var t entity.TrainerProfile
	if err := trainerPreload(r.db.WithContext(ctx)).Where("user_id = ?", userID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *trainerProfileRepository) Update(ctx context.Context, t *entity.TrainerProfile) error {
	return r.db.WithContext(ctx).Omit("User", "Branch").Save(t).Error
}

func (r *trainerProfileRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.TrainerProfile{}, id).Error
}

func (r *trainerProfileRepository) List(ctx context.Context, offset, limit int, order, q string, branchID *uint, isPublic *bool) ([]entity.TrainerProfile, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.TrainerProfile{})
	if q != "" {
		like := "%" + strings.TrimSpace(q) + "%"
		query = query.Where("display_name ILIKE ? OR title ILIKE ? OR specialties ILIKE ?", like, like, like)
	}
	if branchID != nil && *branchID > 0 {
		query = query.Where("branch_id = ?", *branchID)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "id DESC"
	}
	var list []entity.TrainerProfile
	if err := trainerPreload(query).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, err
		}
		return nil, 0, err
	}
	return list, total, nil
}
