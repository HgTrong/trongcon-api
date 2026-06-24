package v1

type OneRepMaxCalculateReq struct {
	Units  string  `json:"units" binding:"required,oneof=kg lbs"`
	Reps   int     `json:"reps" binding:"required,min=1,max=12"`
	Weight float64 `json:"weight" binding:"required,gt=0"`
}

type TrainingPercentage struct {
	Percent int     `json:"percent"`
	Weight  float64 `json:"weight"`
}

type OneRepMaxCalculateRes struct {
	Units               string               `json:"units"`
	Reps                int                  `json:"reps"`
	Weight              float64              `json:"weight"`
	OneRepMax           float64              `json:"one_rep_max"`
	Formula             string               `json:"formula"`
	TrainingPercentages []TrainingPercentage `json:"training_percentages"`
}
