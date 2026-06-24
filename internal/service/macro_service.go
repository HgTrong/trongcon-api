package service

import (
	"errors"
	"math"

	toolsv1 "trongcon-api/api/tools/v1"
)

var ErrInvalidMacroPreset = errors.New("invalid preset")

type macroPreset struct {
	Label   string
	CarbPct int
	Protein int
	FatPct  int
}

var macroPresets = map[string]macroPreset{
	"balanced": {
		Label:   "Balanced (40/30/30)",
		CarbPct: 40,
		Protein: 30,
		FatPct:  30,
	},
	"low_carb": {
		Label:   "Low Carb (20/40/40)",
		CarbPct: 20,
		Protein: 40,
		FatPct:  40,
	},
	"high_protein": {
		Label:   "High Protein (30/40/30)",
		CarbPct: 30,
		Protein: 40,
		FatPct:  30,
	},
	"ketogenic": {
		Label:   "Ketogenic (5/25/70)",
		CarbPct: 5,
		Protein: 25,
		FatPct:  70,
	},
}

type MacroService interface {
	Calculate(req *toolsv1.MacroCalculateReq) (*toolsv1.MacroCalculateRes, error)
}

type macroService struct{}

func NewMacroService() MacroService {
	return &macroService{}
}

func (s *macroService) Calculate(req *toolsv1.MacroCalculateReq) (*toolsv1.MacroCalculateRes, error) {
	preset, ok := macroPresets[req.Preset]
	if !ok {
		return nil, ErrInvalidMacroPreset
	}

	calories := req.DailyCalories
	dailyCarbs := macroGrams(calories, preset.CarbPct, 4)
	dailyProtein := macroGrams(calories, preset.Protein, 4)
	dailyFat := macroGrams(calories, preset.FatPct, 9)

	meals := float64(req.MealsPerDay)

	return &toolsv1.MacroCalculateRes{
		Preset:          req.Preset,
		PresetLabel:     preset.Label,
		DailyCalories:   round1(calories),
		MealsPerDay:     req.MealsPerDay,
		CarbPct:         preset.CarbPct,
		ProteinPct:      preset.Protein,
		FatPct:          preset.FatPct,
		DailyCarbsG:     round0(dailyCarbs),
		DailyProteinG:   round0(dailyProtein),
		DailyFatG:       round0(dailyFat),
		PerMealCarbsG:   round0(dailyCarbs / meals),
		PerMealProteinG: round0(dailyProtein / meals),
		PerMealFatG:     round0(dailyFat / meals),
	}, nil
}

func macroGrams(calories float64, pct int, kcalPerGram float64) float64 {
	return (calories * float64(pct) / 100) / kcalPerGram
}

func round0(v float64) float64 {
	return math.Round(v)
}
