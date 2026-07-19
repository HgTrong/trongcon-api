package repository

import (
	"context"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type ExerciseRepository interface {
	Create(ctx context.Context, ex *entity.Exercise) error
	GetByID(ctx context.Context, id uint) (*entity.Exercise, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Exercise, error)
	Update(ctx context.Context, ex *entity.Exercise) error
	Delete(ctx context.Context, id uint) error
	IncrementViews(ctx context.Context, id uint) (int, error)
	List(ctx context.Context, offset, limit int, order, q, difficulty, force, mechanic, status string, equipmentID, muscleID *uint) ([]entity.Exercise, int64, error)
	SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error)
	ReplaceSteps(ctx context.Context, exerciseID uint, steps []entity.ExerciseStep) error
	ReplaceMuscles(ctx context.Context, exerciseID uint, muscles []entity.ExerciseMuscle) error
}

type exerciseRepository struct {
	db *gorm.DB
}

func NewExerciseRepository(db *gorm.DB) ExerciseRepository {
	return &exerciseRepository{db: db}
}

func (r *exerciseRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("Equipment").Preload("Steps").Preload("Muscles.Muscle")
}

func (r *exerciseRepository) Create(ctx context.Context, ex *entity.Exercise) error {
	return r.db.WithContext(ctx).Create(ex).Error
}

func (r *exerciseRepository) GetByID(ctx context.Context, id uint) (*entity.Exercise, error) {
	var ex entity.Exercise
	if err := r.preload(r.db.WithContext(ctx)).First(&ex, id).Error; err != nil {
		return nil, err
	}
	return &ex, nil
}

func (r *exerciseRepository) GetBySlug(ctx context.Context, slug string) (*entity.Exercise, error) {
	var ex entity.Exercise
	if err := r.preload(r.db.WithContext(ctx)).Where("slug = ?", slug).First(&ex).Error; err != nil {
		return nil, err
	}
	return &ex, nil
}

func (r *exerciseRepository) Update(ctx context.Context, ex *entity.Exercise) error {
	return r.db.WithContext(ctx).Save(ex).Error
}

func (r *exerciseRepository) IncrementViews(ctx context.Context, id uint) (int, error) {
	res := r.db.WithContext(ctx).Model(&entity.Exercise{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1"))
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	var views int
	if err := r.db.WithContext(ctx).Model(&entity.Exercise{}).Where("id = ?", id).Select("views").Scan(&views).Error; err != nil {
		return 0, err
	}
	return views, nil
}

func (r *exerciseRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Exercise{}, id).Error
}

func (r *exerciseRepository) SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&entity.Exercise{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *exerciseRepository) List(ctx context.Context, offset, limit int, order, q, difficulty, force, mechanic, status string, equipmentID, muscleID *uint) ([]entity.Exercise, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Exercise{})
	if q != "" {
		query = query.Where("name ILIKE ? OR summary ILIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if force != "" {
		query = query.Where("force = ?", force)
	}
	if mechanic != "" {
		query = query.Where("mechanic = ?", mechanic)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if equipmentID != nil && *equipmentID > 0 {
		query = query.Where("equipment_id = ?", *equipmentID)
	}
	if muscleID != nil && *muscleID > 0 {
		query = query.Where("id IN (?)",
			r.db.Model(&entity.ExerciseMuscle{}).Select("exercise_id").Where("muscle_id = ?", *muscleID),
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "id DESC"
	}
	order = strings.TrimSpace(order)

	var list []entity.Exercise
	if err := r.preload(query).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *exerciseRepository) ReplaceSteps(ctx context.Context, exerciseID uint, steps []entity.ExerciseStep) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("exercise_id = ?", exerciseID).Delete(&entity.ExerciseStep{}).Error; err != nil {
			return err
		}
		if len(steps) == 0 {
			return nil
		}
		for i := range steps {
			steps[i].ExerciseID = exerciseID
		}
		return tx.Create(&steps).Error
	})
}

func (r *exerciseRepository) ReplaceMuscles(ctx context.Context, exerciseID uint, muscles []entity.ExerciseMuscle) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("exercise_id = ?", exerciseID).Delete(&entity.ExerciseMuscle{}).Error; err != nil {
			return err
		}
		if len(muscles) == 0 {
			return nil
		}
		for i := range muscles {
			muscles[i].ExerciseID = exerciseID
		}
		return tx.Create(&muscles).Error
	})
}
