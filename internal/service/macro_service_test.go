package service

import (
	"testing"

	toolsv1 "trongcon-api/api/tools/v1"
)

func TestMacroCalculate_Balanced2000ThreeMeals(t *testing.T) {
	svc := NewMacroService()
	res, err := svc.Calculate(&toolsv1.MacroCalculateReq{
		Preset:        "balanced",
		DailyCalories: 2000,
		MealsPerDay:   3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.DailyCarbsG != 200 {
		t.Fatalf("expected 200g carbs, got %v", res.DailyCarbsG)
	}
	if res.DailyProteinG != 150 {
		t.Fatalf("expected 150g protein, got %v", res.DailyProteinG)
	}
	if res.DailyFatG != 67 && res.DailyFatG != 66 {
		t.Fatalf("expected ~66-67g fat, got %v", res.DailyFatG)
	}
	if res.PerMealProteinG != 50 {
		t.Fatalf("expected 50g protein per meal, got %v", res.PerMealProteinG)
	}
}
