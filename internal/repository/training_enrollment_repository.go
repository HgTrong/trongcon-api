package repository

import (
	"context"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type TrainingEnrollmentRepository interface {
	Create(ctx context.Context, e *entity.TrainingEnrollment) error
	GetByID(ctx context.Context, userID, id uint) (*entity.TrainingEnrollment, error)
	Update(ctx context.Context, e *entity.TrainingEnrollment) error
	List(ctx context.Context, userID uint, offset, limit int, status string) ([]entity.TrainingEnrollment, int64, error)
	GetSlot(ctx context.Context, userID, slotID uint) (*entity.EnrollmentSlot, *entity.TrainingEnrollment, error)
	UpdateSlot(ctx context.Context, slot *entity.EnrollmentSlot) error
	ListCompletedSessionsForEnrollment(ctx context.Context, userID, enrollmentID uint) ([]entity.WorkoutSession, error)
}

type trainingEnrollmentRepository struct {
	db *gorm.DB
}

func NewTrainingEnrollmentRepository(db *gorm.DB) TrainingEnrollmentRepository {
	return &trainingEnrollmentRepository{db: db}
}

func enrollmentPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("Slots", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("week_index ASC, day_index ASC, id ASC")
	})
}

func (r *trainingEnrollmentRepository) Create(ctx context.Context, e *entity.TrainingEnrollment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		slots := e.Slots
		e.Slots = nil
		if err := tx.Create(e).Error; err != nil {
			return err
		}
		for i := range slots {
			slots[i].EnrollmentID = e.ID
		}
		if len(slots) > 0 {
			return tx.Create(&slots).Error
		}
		return nil
	})
}

func (r *trainingEnrollmentRepository) GetByID(ctx context.Context, userID, id uint) (*entity.TrainingEnrollment, error) {
	var e entity.TrainingEnrollment
	if err := enrollmentPreload(r.db.WithContext(ctx)).
		Where("user_id = ? AND id = ?", userID, id).
		First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *trainingEnrollmentRepository) Update(ctx context.Context, e *entity.TrainingEnrollment) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *trainingEnrollmentRepository) List(ctx context.Context, userID uint, offset, limit int, status string) ([]entity.TrainingEnrollment, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.TrainingEnrollment{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.TrainingEnrollment
	if err := enrollmentPreload(query).Order("start_date DESC, id DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *trainingEnrollmentRepository) GetSlot(ctx context.Context, userID, slotID uint) (*entity.EnrollmentSlot, *entity.TrainingEnrollment, error) {
	var slot entity.EnrollmentSlot
	if err := r.db.WithContext(ctx).First(&slot, slotID).Error; err != nil {
		return nil, nil, err
	}
	var en entity.TrainingEnrollment
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, slot.EnrollmentID).First(&en).Error; err != nil {
		return nil, nil, err
	}
	return &slot, &en, nil
}

func (r *trainingEnrollmentRepository) UpdateSlot(ctx context.Context, slot *entity.EnrollmentSlot) error {
	return r.db.WithContext(ctx).Save(slot).Error
}

func (r *trainingEnrollmentRepository) ListCompletedSessionsForEnrollment(ctx context.Context, userID, enrollmentID uint) ([]entity.WorkoutSession, error) {
	var slots []entity.EnrollmentSlot
	if err := r.db.WithContext(ctx).Where("enrollment_id = ? AND session_id IS NOT NULL", enrollmentID).Find(&slots).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(slots))
	for _, s := range slots {
		if s.SessionID != nil {
			ids = append(ids, *s.SessionID)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var sessions []entity.WorkoutSession
	err := r.db.WithContext(ctx).
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Items.Sets", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("set_number ASC, id ASC")
		}).
		Where("user_id = ? AND id IN ? AND status = ?", userID, ids, entity.SessionStatusCompleted).
		Find(&sessions).Error
	return sessions, err
}
