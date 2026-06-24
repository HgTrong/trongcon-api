package service

import (
	"errors"
	"math"

	toolsv1 "trongcon-api/api/tools/v1"
)

var (
	ErrInvalidActivityLevel = errors.New("invalid activity_level")
	ErrInvalidGoal          = errors.New("invalid goal")
)

const tdeeFormulaName = "Harris-Benedict (revised)"

var activityFactors = map[string]float64{
	"sedentary":   1.2,
	"light":       1.375,
	"moderate":    1.55,
	"active":      1.725,
	"very_active": 1.9,
}

// 1 kg/week ≈ 1100 kcal/day (7700 kcal per kg fat / 7 days).
var goalAdjustments = map[string]int{
	"extreme_weight_loss": -1100,
	"weight_loss":         -550,
	"mild_weight_loss":    -275,
	"maintain":            0,
	"mild_weight_gain":    275,
	"weight_gain":         550,
	"extreme_weight_gain": 1100,
}

type TDEEService interface {
	Calculate(req *toolsv1.CalculateReq) (*toolsv1.CalculateRes, error)
}

type tdeeService struct{}

func NewTDEEService() TDEEService {
	return &tdeeService{}
}

func (s *tdeeService) Calculate(req *toolsv1.CalculateReq) (*toolsv1.CalculateRes, error) {
	activityFactor, ok := activityFactors[req.ActivityLevel]
	if !ok {
		return nil, ErrInvalidActivityLevel
	}
	goalAdj, ok := goalAdjustments[req.Goal]
	if !ok {
		return nil, ErrInvalidGoal
	}

	heightCm, weightKg, err := normalizeBodyMetrics(req)
	if err != nil {
		return nil, err
	}

	bmr := harrisBenedictBMR(req.Gender, weightKg, heightCm, req.Age)
	tdee := bmr * activityFactor
	daily := tdee + float64(goalAdj)
	if daily < 800 {
		daily = 800
	}

	return &toolsv1.CalculateRes{
		BMR:            round1(bmr),
		TDEE:           round1(tdee),
		DailyCalories:  round1(daily),
		HeightCm:       round1(heightCm),
		WeightKg:       round1(weightKg),
		Formula:        tdeeFormulaName,
		ActivityFactor: activityFactor,
		GoalAdjustment: goalAdj,
		Goal:           req.Goal,
		ActivityLevel:  req.ActivityLevel,
	}, nil
}

func normalizeBodyMetrics(req *toolsv1.CalculateReq) (heightCm, weightKg float64, err error) {
	switch req.Units {
	case "metric":
		heightCm = req.Height
		weightKg = req.Weight
	case "imperial":
		heightIn := req.Height*12 + req.HeightInches
		if heightIn <= 0 {
			return 0, 0, errors.New("invalid height for imperial units")
		}
		heightCm = heightIn * 2.54
		weightKg = req.Weight * 0.45359237
	default:
		return 0, 0, errors.New("invalid units")
	}
	if heightCm <= 0 || weightKg <= 0 {
		return 0, 0, errors.New("height and weight must be greater than 0")
	}
	return heightCm, weightKg, nil
}

func harrisBenedictBMR(gender string, weightKg, heightCm float64, age int) float64 {
	switch gender {
	case "male":
		return 88.362 + (13.397 * weightKg) + (4.799 * heightCm) - (5.677 * float64(age))
	case "female":
		return 447.593 + (9.247 * weightKg) + (3.098 * heightCm) - (4.330 * float64(age))
	default:
		return 0
	}
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
