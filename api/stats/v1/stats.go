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

// TopTrainerRes ranks a PT by profile views + all attributed content views.
type TopTrainerRes struct {
	TrainerID     uint   `json:"trainer_id"`
	UserID        uint   `json:"user_id"`
	DisplayName   string `json:"display_name"`
	Title         string `json:"title,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	ProfileViews  int64  `json:"profile_views"`
	WorkoutViews  int64  `json:"workout_views"`
	RoutineViews  int64  `json:"routine_views"`
	MealPlanViews int64  `json:"meal_plan_views"`
	ArticleViews  int64  `json:"article_views"`
	TotalViews    int64  `json:"total_views"`
}

type TopTrainersRes struct {
	Data []TopTrainerRes `json:"data"`
}
