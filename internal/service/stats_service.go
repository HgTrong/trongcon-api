package service

import (
	"context"

	statsv1 "trongcon-api/api/stats/v1"
	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type StatsService interface {
	Overview(ctx context.Context) (*statsv1.OverviewRes, error)
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
