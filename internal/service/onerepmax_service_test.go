package service

import (
	"testing"

	toolsv1 "trongcon-api/api/tools/v1"
)

func TestOneRepMaxService_Brzycki(t *testing.T) {
	svc := NewOneRepMaxService()

	res, err := svc.Calculate(&toolsv1.OneRepMaxCalculateReq{
		Units:  "kg",
		Reps:   5,
		Weight: 45.36,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.OneRepMax != 51 {
		t.Fatalf("expected 1RM 51, got %v", res.OneRepMax)
	}
	if len(res.TrainingPercentages) != 11 {
		t.Fatalf("expected 11 percentages, got %d", len(res.TrainingPercentages))
	}
	if res.TrainingPercentages[0].Percent != 50 || res.TrainingPercentages[0].Weight != 26 {
		t.Fatalf("expected 50%% = 26 kg, got %v", res.TrainingPercentages[0])
	}
	if res.TrainingPercentages[5].Percent != 75 || res.TrainingPercentages[5].Weight != 38 {
		t.Fatalf("expected 75%% = 38 kg, got %v", res.TrainingPercentages[5])
	}
}

func TestBrzyckiOneRepMax_SingleRep(t *testing.T) {
	got := brzyckiOneRepMax(100, 1)
	if got != 100 {
		t.Fatalf("expected 100 for 1 rep, got %v", got)
	}
}
