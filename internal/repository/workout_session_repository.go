package repository

import (
	"context"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type WorkoutSessionRepository interface {
	Create(ctx context.Context, s *entity.WorkoutSession) error
	GetByID(ctx context.Context, userID, id uint) (*entity.WorkoutSession, error)
	Update(ctx context.Context, s *entity.WorkoutSession) error
	UpdateSessionFields(ctx context.Context, id uint, fields map[string]interface{}) error
	List(ctx context.Context, userID uint, offset, limit int, status string) ([]entity.WorkoutSession, int64, error)
	GetSet(ctx context.Context, userID, setID uint) (*entity.WorkoutSetLog, *entity.WorkoutSessionItem, error)
	UpdateSet(ctx context.Context, set *entity.WorkoutSetLog) error
	UpdateSetFields(ctx context.Context, setID uint, fields map[string]interface{}) error
	AddItem(ctx context.Context, item *entity.WorkoutSessionItem) error
	LastCompletedItemForExercise(ctx context.Context, userID, exerciseID, excludeSessionID uint) (*entity.WorkoutSessionItem, error)
	ListExerciseHistory(ctx context.Context, userID, exerciseID uint, limit int) ([]entity.WorkoutSessionItem, error)
}

type workoutSessionRepository struct {
	db *gorm.DB
}

func NewWorkoutSessionRepository(db *gorm.DB) WorkoutSessionRepository {
	return &workoutSessionRepository{db: db}
}

func sessionPreload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Items.Sets", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("set_number ASC, id ASC")
		})
}

func (r *workoutSessionRepository) Create(ctx context.Context, s *entity.WorkoutSession) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items := s.Items
		s.Items = nil
		if err := tx.Create(s).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].SessionID = s.ID
			items[i].SortOrder = i + 1
			sets := items[i].Sets
			items[i].Sets = nil
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
			for j := range sets {
				sets[j].SessionItemID = items[i].ID
				sets[j].SetNumber = j + 1
			}
			if len(sets) > 0 {
				if err := tx.Create(&sets).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *workoutSessionRepository) GetByID(ctx context.Context, userID, id uint) (*entity.WorkoutSession, error) {
	var s entity.WorkoutSession
	if err := sessionPreload(r.db.WithContext(ctx)).
		Where("user_id = ? AND id = ?", userID, id).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *workoutSessionRepository) Update(ctx context.Context, s *entity.WorkoutSession) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *workoutSessionRepository) UpdateSessionFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.WorkoutSession{}).Where("id = ?", id).Updates(fields).Error
}

func (r *workoutSessionRepository) List(ctx context.Context, userID uint, offset, limit int, status string) ([]entity.WorkoutSession, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.WorkoutSession{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.WorkoutSession
	if err := sessionPreload(query).Order("started_at DESC, id DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *workoutSessionRepository) GetSet(ctx context.Context, userID, setID uint) (*entity.WorkoutSetLog, *entity.WorkoutSessionItem, error) {
	var set entity.WorkoutSetLog
	if err := r.db.WithContext(ctx).First(&set, setID).Error; err != nil {
		return nil, nil, err
	}
	var item entity.WorkoutSessionItem
	if err := r.db.WithContext(ctx).First(&item, set.SessionItemID).Error; err != nil {
		return nil, nil, err
	}
	var session entity.WorkoutSession
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, item.SessionID).First(&session).Error; err != nil {
		return nil, nil, err
	}
	return &set, &item, nil
}

func (r *workoutSessionRepository) UpdateSet(ctx context.Context, set *entity.WorkoutSetLog) error {
	return r.db.WithContext(ctx).Save(set).Error
}

func (r *workoutSessionRepository) UpdateSetFields(ctx context.Context, setID uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.WorkoutSetLog{}).Where("id = ?", setID).Updates(fields).Error
}

func (r *workoutSessionRepository) AddItem(ctx context.Context, item *entity.WorkoutSessionItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sets := item.Sets
		item.Sets = nil
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		for j := range sets {
			sets[j].SessionItemID = item.ID
			sets[j].SetNumber = j + 1
		}
		if len(sets) > 0 {
			return tx.Create(&sets).Error
		}
		return nil
	})
}

func (r *workoutSessionRepository) LastCompletedItemForExercise(ctx context.Context, userID, exerciseID, excludeSessionID uint) (*entity.WorkoutSessionItem, error) {
	var item entity.WorkoutSessionItem
	q := r.db.WithContext(ctx).
		Joins("JOIN workout_sessions ON workout_sessions.id = workout_session_items.session_id").
		Where("workout_sessions.user_id = ? AND workout_session_items.exercise_id = ?", userID, exerciseID).
		Where("workout_sessions.status = ?", entity.SessionStatusCompleted)
	if excludeSessionID > 0 {
		q = q.Where("workout_sessions.id <> ?", excludeSessionID)
	}
	err := q.Preload("Sets", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("set_number ASC, id ASC")
	}).Order("workout_sessions.completed_at DESC NULLS LAST, workout_sessions.id DESC").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *workoutSessionRepository) ListExerciseHistory(ctx context.Context, userID, exerciseID uint, limit int) ([]entity.WorkoutSessionItem, error) {
	if limit < 1 {
		limit = 12
	}
	var items []entity.WorkoutSessionItem
	err := r.db.WithContext(ctx).
		Joins("JOIN workout_sessions ON workout_sessions.id = workout_session_items.session_id").
		Where("workout_sessions.user_id = ? AND workout_session_items.exercise_id = ?", userID, exerciseID).
		Where("workout_sessions.status = ?", entity.SessionStatusCompleted).
		Preload("Sets", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("set_number ASC, id ASC")
		}).
		Order("workout_sessions.completed_at DESC NULLS LAST, workout_sessions.id DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}
