package service

import (
	"context"

	statsv1 "trongcon-api/api/stats/v1"
	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type StatsService interface {
	Overview(ctx context.Context) (*statsv1.OverviewRes, error)
	TopTrainers(ctx context.Context, limit int) (*statsv1.TopTrainersRes, error)
}

type statsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) StatsService {
	return &statsService{db: db}
}

func (s *statsService) count(ctx context.Context, model interface{}) (int64, error) {
	var n int64
	if err := s.db.WithContext(ctx).Model(model).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (s *statsService) Overview(ctx context.Context) (*statsv1.OverviewRes, error) {
	res := &statsv1.OverviewRes{}
	var err error
	if res.Exercises, err = s.count(ctx, &entity.Exercise{}); err != nil {
		return nil, err
	}
	if res.Muscles, err = s.count(ctx, &entity.Muscle{}); err != nil {
		return nil, err
	}
	if res.Equipments, err = s.count(ctx, &entity.Equipment{}); err != nil {
		return nil, err
	}
	if res.Foods, err = s.count(ctx, &entity.Food{}); err != nil {
		return nil, err
	}
	if res.Workouts, err = s.count(ctx, &entity.Workout{}); err != nil {
		return nil, err
	}
	if res.Routines, err = s.count(ctx, &entity.Routine{}); err != nil {
		return nil, err
	}
	if res.MealPlans, err = s.count(ctx, &entity.MealPlan{}); err != nil {
		return nil, err
	}
	if res.Articles, err = s.count(ctx, &entity.Article{}); err != nil {
		return nil, err
	}
	if res.Categories, err = s.count(ctx, &entity.Category{}); err != nil {
		return nil, err
	}
	if res.Users, err = s.count(ctx, &entity.User{}); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *statsService) TopTrainers(ctx context.Context, limit int) (*statsv1.TopTrainersRes, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	type row struct {
		TrainerID     uint   `gorm:"column:trainer_id"`
		UserID        uint   `gorm:"column:user_id"`
		DisplayName   string `gorm:"column:display_name"`
		Title         string `gorm:"column:title"`
		AvatarURL     string `gorm:"column:avatar_url"`
		ProfileViews  int64  `gorm:"column:profile_views"`
		WorkoutViews  int64  `gorm:"column:workout_views"`
		RoutineViews  int64  `gorm:"column:routine_views"`
		MealPlanViews int64  `gorm:"column:meal_plan_views"`
		ArticleViews  int64  `gorm:"column:article_views"`
		TotalViews    int64  `gorm:"column:total_views"`
	}
	var rows []row
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			tp.id AS trainer_id,
			tp.user_id AS user_id,
			tp.display_name AS display_name,
			COALESCE(tp.title, '') AS title,
			COALESCE(u.profile_picture, '') AS avatar_url,
			COALESCE(tp.views, 0) AS profile_views,
			COALESCE(w.views, 0) AS workout_views,
			COALESCE(r.views, 0) AS routine_views,
			COALESCE(m.views, 0) AS meal_plan_views,
			COALESCE(a.views, 0) AS article_views,
			COALESCE(tp.views, 0)
				+ COALESCE(w.views, 0)
				+ COALESCE(r.views, 0)
				+ COALESCE(m.views, 0)
				+ COALESCE(a.views, 0) AS total_views
		FROM trainer_profiles tp
		LEFT JOIN users u ON u.id = tp.user_id AND u.deleted_at IS NULL
		LEFT JOIN (
			SELECT user_id, SUM(views) AS views
			FROM workouts
			WHERE deleted_at IS NULL AND user_id > 0
			GROUP BY user_id
		) w ON w.user_id = tp.user_id
		LEFT JOIN (
			SELECT user_id, SUM(views) AS views
			FROM routines
			WHERE deleted_at IS NULL AND user_id > 0
			GROUP BY user_id
		) r ON r.user_id = tp.user_id
		LEFT JOIN (
			SELECT user_id, SUM(views) AS views
			FROM meal_plans
			WHERE deleted_at IS NULL AND user_id > 0
			GROUP BY user_id
		) m ON m.user_id = tp.user_id
		LEFT JOIN (
			SELECT user_id, SUM(views) AS views
			FROM articles
			WHERE deleted_at IS NULL AND user_id > 0
			GROUP BY user_id
		) a ON a.user_id = tp.user_id
		WHERE tp.deleted_at IS NULL
		ORDER BY total_views DESC, tp.display_name ASC
		LIMIT ?
	`, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]statsv1.TopTrainerRes, 0, len(rows))
	for _, r := range rows {
		out = append(out, statsv1.TopTrainerRes{
			TrainerID:     r.TrainerID,
			UserID:        r.UserID,
			DisplayName:   r.DisplayName,
			Title:         r.Title,
			AvatarURL:     r.AvatarURL,
			ProfileViews:  r.ProfileViews,
			WorkoutViews:  r.WorkoutViews,
			RoutineViews:  r.RoutineViews,
			MealPlanViews: r.MealPlanViews,
			ArticleViews:  r.ArticleViews,
			TotalViews:    r.TotalViews,
		})
	}
	return &statsv1.TopTrainersRes{Data: out}, nil
}
