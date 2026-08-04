package repository

import (
	"context"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type WorkoutRepository interface {
	Create(ctx context.Context, w *entity.Workout) error
	GetByID(ctx context.Context, id uint) (*entity.Workout, error)
	Update(ctx context.Context, w *entity.Workout) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q, difficulty, goal string) ([]entity.Workout, int64, error)
	ListCatalog(ctx context.Context, offset, limit int, order, q, difficulty, goal string) ([]entity.Workout, int64, error)
	ListAll(ctx context.Context, offset, limit int, order, q, difficulty, goal string) ([]entity.Workout, int64, error)
	ListByOwner(ctx context.Context, ownerID uint, offset, limit int, order, q string) ([]entity.Workout, int64, error)
	ListPublicByOwner(ctx context.Context, ownerID uint, offset, limit int, order string) ([]entity.Workout, int64, error)
	ReplaceItems(ctx context.Context, workoutID uint, items []entity.WorkoutItem) error
	IncrementViews(ctx context.Context, id uint) (int64, error)
}

type workoutRepository struct {
	db *gorm.DB
}

func NewWorkoutRepository(db *gorm.DB) WorkoutRepository {
	return &workoutRepository{db: db}
}

func workoutItemPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("User").Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order ASC, id ASC")
	})
}

func (r *workoutRepository) Create(ctx context.Context, w *entity.Workout) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items := w.Items
		w.Items = nil
		if err := tx.Create(w).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].WorkoutID = w.ID
			items[i].SortOrder = i + 1
		}
		return tx.Create(&items).Error
	})
}

func (r *workoutRepository) GetByID(ctx context.Context, id uint) (*entity.Workout, error) {
	var w entity.Workout
	if err := workoutItemPreload(r.db.WithContext(ctx)).First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *workoutRepository) Update(ctx context.Context, w *entity.Workout) error {
	return r.db.WithContext(ctx).Omit("User").Save(w).Error
}

func (r *workoutRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Workout{}, id).Error
}

func (r *workoutRepository) listFiltered(ctx context.Context, query *gorm.DB, offset, limit int, order string) ([]entity.Workout, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "id DESC"
	}
	order = strings.TrimSpace(order)

	var list []entity.Workout
	if err := workoutItemPreload(query).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *workoutRepository) List(ctx context.Context, offset, limit int, order, q, difficulty, goal string) ([]entity.Workout, int64, error) {
	return r.ListCatalog(ctx, offset, limit, order, q, difficulty, goal)
}

func (r *workoutRepository) ListCatalog(ctx context.Context, offset, limit int, order, q, difficulty, goal string) ([]entity.Workout, int64, error) {
	// Catalog = admin-owned (owner null) + published PT workouts.
	query := r.db.WithContext(ctx).Model(&entity.Workout{}).Where("owner_user_id IS NULL OR is_public = ?", true)
	if q != "" {
		query = query.Where("title ILIKE ?", "%"+q+"%")
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if goal != "" {
		query = query.Where("goal = ?", goal)
	}
	return r.listFiltered(ctx, query, offset, limit, order)
}

func (r *workoutRepository) ListAll(ctx context.Context, offset, limit int, order, q, difficulty, goal string) ([]entity.Workout, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Workout{})
	if q != "" {
		query = query.Where("title ILIKE ?", "%"+q+"%")
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if goal != "" {
		query = query.Where("goal = ?", goal)
	}
	return r.listFiltered(ctx, query, offset, limit, order)
}

func (r *workoutRepository) ListPublicByOwner(ctx context.Context, ownerID uint, offset, limit int, order string) ([]entity.Workout, int64, error) {
	// Published personal copies OR catalog items authored by this user.
	query := r.db.WithContext(ctx).Model(&entity.Workout{}).Where(
		"user_id = ? AND (is_public = ? OR owner_user_id IS NULL)",
		ownerID, true,
	)
	return r.listFiltered(ctx, query, offset, limit, order)
}

func (r *workoutRepository) ListByOwner(ctx context.Context, ownerID uint, offset, limit int, order, q string) ([]entity.Workout, int64, error) {
	// Personal copies (owner) + anything authored/posted by this user (admin "Posted by" or PT create).
	query := r.db.WithContext(ctx).Model(&entity.Workout{}).Where(
		"owner_user_id = ? OR user_id = ?",
		ownerID, ownerID,
	)
	if q != "" {
		query = query.Where("title ILIKE ?", "%"+q+"%")
	}
	return r.listFiltered(ctx, query, offset, limit, order)
}

func (r *workoutRepository) ReplaceItems(ctx context.Context, workoutID uint, items []entity.WorkoutItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workout_id = ?", workoutID).Delete(&entity.WorkoutItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].WorkoutID = workoutID
			items[i].SortOrder = i + 1
		}
		return tx.Create(&items).Error
	})
}

func (r *workoutRepository) IncrementViews(ctx context.Context, id uint) (int64, error) {
	res := r.db.WithContext(ctx).Model(&entity.Workout{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1"))
	if res.Error != nil {
		return 0, res.Error
	}
	var views int64
	if err := r.db.WithContext(ctx).Model(&entity.Workout{}).Where("id = ?", id).Select("views").Scan(&views).Error; err != nil {
		return 0, err
	}
	return views, nil
}
