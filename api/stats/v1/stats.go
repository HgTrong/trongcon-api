package v1

type OverviewRes struct {
	Exercises  int64 `json:"exercises"`
	Muscles    int64 `json:"muscles"`
	Equipments int64 `json:"equipments"`
	Foods      int64 `json:"foods"`
	Workouts   int64 `json:"workouts"`
	Routines   int64 `json:"routines"`
	MealPlans  int64 `json:"meal_plans"`
	Articles   int64 `json:"articles"`
	Categories int64 `json:"categories"`
	Users      int64 `json:"users"`
}
