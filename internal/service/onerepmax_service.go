package service

import (
	toolsv1 "trongcon-api/api/tools/v1"
)

const oneRepMaxFormulaName = "Brzycki"

var trainingPercents = []int{50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100}

type OneRepMaxService interface {
	Calculate(req *toolsv1.OneRepMaxCalculateReq) (*toolsv1.OneRepMaxCalculateRes, error)
}

type oneRepMaxService struct{}

func NewOneRepMaxService() OneRepMaxService {
	return &oneRepMaxService{}
}

func (s *oneRepMaxService) Calculate(req *toolsv1.OneRepMaxCalculateReq) (*toolsv1.OneRepMaxCalculateRes, error) {
	oneRM := brzyckiOneRepMax(req.Weight, req.Reps)
	rounded := round0(oneRM)

	percentages := make([]toolsv1.TrainingPercentage, 0, len(trainingPercents))
	for _, pct := range trainingPercents {
		percentages = append(percentages, toolsv1.TrainingPercentage{
			Percent: pct,
			Weight:  round0(oneRM * float64(pct) / 100),
		})
	}

	return &toolsv1.OneRepMaxCalculateRes{
		Units:               req.Units,
		Reps:                req.Reps,
		Weight:              round1(req.Weight),
		OneRepMax:           rounded,
		Formula:             oneRepMaxFormulaName,
		TrainingPercentages: percentages,
	}, nil
}

func brzyckiOneRepMax(weight float64, reps int) float64 {
	if reps <= 1 {
		return weight
	}
	return weight / (1.0278 - 0.0278*float64(reps))
}
