package service

import (
	"testing"

	toolsv1 "trongcon-api/api/tools/v1"
)

func TestTDEECalculate_MaleMetricMaintain(t *testing.T) {
	svc := NewTDEEService()
	res, err := svc.Calculate(&toolsv1.CalculateReq{
		Gender:        "male",
		Units:         "metric",
		Age:           18,
		Height:        150,
		Weight:        100,
		ActivityLevel: "sedentary",
		Goal:          "maintain",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.DailyCalories != 2454.9 && res.DailyCalories != 2455 {
		t.Fatalf("expected ~2455 daily calories, got %v (bmr=%v tdee=%v)", res.DailyCalories, res.BMR, res.TDEE)
	}
}

func TestTDEECalculate_WeightLoss(t *testing.T) {
	svc := NewTDEEService()
	res, err := svc.Calculate(&toolsv1.CalculateReq{
		Gender:        "male",
		Units:         "metric",
		Age:           18,
		Height:        150,
		Weight:        100,
		ActivityLevel: "sedentary",
		Goal:          "weight_loss",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := res.TDEE - 550
	if res.DailyCalories != round1(want) {
		t.Fatalf("expected %v, got %v", round1(want), res.DailyCalories)
	}
}
