package repository

import (
	"context"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PTContentStatRepository interface {
	UpsertTouch(ctx context.Context, trainerProfileID uint, contentType string, contentID uint, title string, viewsDelta, likesDelta, savesDelta, profileDelta, bookingDelta int64) error
	ListByTrainer(ctx context.Context, trainerProfileID uint) ([]entity.PTContentStat, error)
	SumByTrainer(ctx context.Context, trainerProfileID uint) (views, likes, saves, profileVisits, bookings int64, err error)
	BackfillFromCatalog(ctx context.Context, trainerProfileID, authorUserID uint) error
}

type ptContentStatRepository struct{ db *gorm.DB }

func NewPTContentStatRepository(db *gorm.DB) PTContentStatRepository {
	return &ptContentStatRepository{db: db}
}

func (r *ptContentStatRepository) UpsertTouch(ctx context.Context, trainerProfileID uint, contentType string, contentID uint, title string, viewsDelta, likesDelta, savesDelta, profileDelta, bookingDelta int64) error {
	row := entity.PTContentStat{
		TrainerProfileID:  trainerProfileID,
		ContentType:       contentType,
		ContentID:         contentID,
		Title:             title,
		Views:             viewsDelta,
		Likes:             likesDelta,
		Saves:             savesDelta,
		ProfileVisits:     profileDelta,
		BookingsGenerated: bookingDelta,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "trainer_profile_id"}, {Name: "content_type"}, {Name: "content_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":              gorm.Expr("CASE WHEN EXCLUDED.title <> '' THEN EXCLUDED.title ELSE pt_content_stats.title END"),
			"views":              gorm.Expr("pt_content_stats.views + EXCLUDED.views"),
			"likes":              gorm.Expr("pt_content_stats.likes + EXCLUDED.likes"),
			"saves":              gorm.Expr("pt_content_stats.saves + EXCLUDED.saves"),
			"profile_visits":     gorm.Expr("pt_content_stats.profile_visits + EXCLUDED.profile_visits"),
			"bookings_generated": gorm.Expr("pt_content_stats.bookings_generated + EXCLUDED.bookings_generated"),
			"updated_at":         gorm.Expr("NOW()"),
		}),
	}).Create(&row).Error
}

func (r *ptContentStatRepository) ListByTrainer(ctx context.Context, trainerProfileID uint) ([]entity.PTContentStat, error) {
	var rows []entity.PTContentStat
	err := r.db.WithContext(ctx).
		Where("trainer_profile_id = ?", trainerProfileID).
		Order("bookings_generated DESC, views DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *ptContentStatRepository) SumByTrainer(ctx context.Context, trainerProfileID uint) (views, likes, saves, profileVisits, bookings int64, err error) {
	type agg struct {
		Views         int64
		Likes         int64
		Saves         int64
		ProfileVisits int64
		Bookings      int64
	}
	var a agg
	err = r.db.WithContext(ctx).Model(&entity.PTContentStat{}).
		Select(`COALESCE(SUM(views),0) as views,
			COALESCE(SUM(likes),0) as likes,
			COALESCE(SUM(saves),0) as saves,
			COALESCE(SUM(profile_visits),0) as profile_visits,
			COALESCE(SUM(bookings_generated),0) as bookings`).
		Where("trainer_profile_id = ?", trainerProfileID).
		Scan(&a).Error
	return a.Views, a.Likes, a.Saves, a.ProfileVisits, a.Bookings, err
}

// BackfillFromCatalog copies current catalog view counters into pt_content_stats (idempotent bump via upsert of absolute views when zero-row).
func (r *ptContentStatRepository) BackfillFromCatalog(ctx context.Context, trainerProfileID, authorUserID uint) error {
	type item struct {
		ContentType string
		ContentID   uint
		Title       string
		Views       int64
	}
	var items []item
	queries := []string{
		`SELECT 'article' as content_type, id as content_id, title, views FROM articles WHERE user_id = ? AND deleted_at IS NULL`,
		`SELECT 'workout' as content_type, id as content_id, title, views FROM workouts WHERE user_id = ? AND deleted_at IS NULL AND is_public = true`,
		`SELECT 'routine' as content_type, id as content_id, title, views FROM routines WHERE user_id = ? AND deleted_at IS NULL AND is_public = true`,
		`SELECT 'meal_plan' as content_type, id as content_id, title, views FROM meal_plans WHERE user_id = ? AND deleted_at IS NULL AND is_public = true`,
	}
	for _, q := range queries {
		var rows []item
		if err := r.db.WithContext(ctx).Raw(q, authorUserID).Scan(&rows).Error; err != nil {
			continue
		}
		items = append(items, rows...)
	}
	for _, it := range items {
		var existing entity.PTContentStat
		err := r.db.WithContext(ctx).
			Where("trainer_profile_id = ? AND content_type = ? AND content_id = ?", trainerProfileID, it.ContentType, it.ContentID).
			First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			_ = r.db.WithContext(ctx).Create(&entity.PTContentStat{
				TrainerProfileID: trainerProfileID,
				ContentType:      it.ContentType,
				ContentID:        it.ContentID,
				Title:            it.Title,
				Views:            it.Views,
			}).Error
			continue
		}
		if err == nil && existing.Title == "" && it.Title != "" {
			_ = r.db.WithContext(ctx).Model(&existing).Update("title", it.Title).Error
		}
	}
	return nil
}

type PTAttributionRepository interface {
	Upsert(ctx context.Context, a *entity.PTAttribution) error
	Get(ctx context.Context, userID, trainerProfileID uint) (*entity.PTAttribution, error)
	Clear(ctx context.Context, userID, trainerProfileID uint) error
}

type ptAttributionRepository struct{ db *gorm.DB }

func NewPTAttributionRepository(db *gorm.DB) PTAttributionRepository {
	return &ptAttributionRepository{db: db}
}

func (r *ptAttributionRepository) Upsert(ctx context.Context, a *entity.PTAttribution) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "trainer_profile_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"content_type", "content_id", "title", "updated_at"}),
	}).Create(a).Error
}

func (r *ptAttributionRepository) Get(ctx context.Context, userID, trainerProfileID uint) (*entity.PTAttribution, error) {
	var a entity.PTAttribution
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND trainer_profile_id = ?", userID, trainerProfileID).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ptAttributionRepository) Clear(ctx context.Context, userID, trainerProfileID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND trainer_profile_id = ?", userID, trainerProfileID).
		Delete(&entity.PTAttribution{}).Error
}

type PTReviewRepository interface {
	Create(ctx context.Context, rev *entity.PTReview) error
	GetBySessionOfferID(ctx context.Context, offerID uint) (*entity.PTReview, error)
	ListByTrainer(ctx context.Context, trainerProfileID uint, offset, limit int) ([]entity.PTReview, int64, error)
	AvgRating(ctx context.Context, trainerProfileID uint) (avg float64, count int64, err error)
}

type ptReviewRepository struct{ db *gorm.DB }

func NewPTReviewRepository(db *gorm.DB) PTReviewRepository {
	return &ptReviewRepository{db: db}
}

func (r *ptReviewRepository) Create(ctx context.Context, rev *entity.PTReview) error {
	return r.db.WithContext(ctx).Create(rev).Error
}

func (r *ptReviewRepository) GetBySessionOfferID(ctx context.Context, offerID uint) (*entity.PTReview, error) {
	var rev entity.PTReview
	err := r.db.WithContext(ctx).Where("session_offer_id = ?", offerID).First(&rev).Error
	if err != nil {
		return nil, err
	}
	return &rev, nil
}

func (r *ptReviewRepository) ListByTrainer(ctx context.Context, trainerProfileID uint, offset, limit int) ([]entity.PTReview, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.PTReview{}).Where("trainer_profile_id = ?", trainerProfileID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.PTReview
	err := r.db.WithContext(ctx).Where("trainer_profile_id = ?", trainerProfileID).
		Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ptReviewRepository) AvgRating(ctx context.Context, trainerProfileID uint) (float64, int64, error) {
	type row struct {
		Avg   float64
		Count int64
	}
	var out row
	err := r.db.WithContext(ctx).Model(&entity.PTReview{}).
		Select("COALESCE(AVG(rating),0) as avg, COUNT(*) as count").
		Where("trainer_profile_id = ?", trainerProfileID).
		Scan(&out).Error
	return out.Avg, out.Count, err
}
