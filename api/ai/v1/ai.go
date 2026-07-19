package v1

import mealplanv1 "trongcon-api/api/meal_plan/v1"

type GenerateMealPlanReq struct {
	Calories     float64  `json:"calories" binding:"required,min=800,max=6000"`
	ProteinG     *float64 `json:"protein_g"`
	CarbG        *float64 `json:"carb_g"`
	FatG         *float64 `json:"fat_g"`
	MealsPerDay  int      `json:"meals_per_day" binding:"required,min=2,max=6"`
	Goal         string   `json:"goal" binding:"required,oneof=cut maintain bulk"`
	Diet         string   `json:"diet"` // none, vegetarian, vegan, high_protein
	Allergies    []string `json:"allergies"`
	DislikeFoods []string `json:"dislike_foods"`
	Cuisine      string   `json:"cuisine"`
	Notes        string   `json:"notes"`
	Title        string   `json:"title"`
}

type GenerateMealPlanRes struct {
	Preview      mealplanv1.MealPlanRes `json:"preview"`
	CandidatesN  int                    `json:"candidates_n"`
	SaveHint     string                 `json:"save_hint"`
}

type CreateMyMealPlanReq struct {
	Title       string                       `json:"title" binding:"required,min=1,max=200"`
	Description string                       `json:"description"`
	IsPublic    bool                         `json:"is_public"`
	Meals       []mealplanv1.MealPlanMealInput `json:"meals" binding:"required,min=1"`
}

type UpdateMyMealPlanReq struct {
	Title       *string                        `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string                        `json:"description"`
	IsPublic    *bool                          `json:"is_public"`
	Meals       *[]mealplanv1.MealPlanMealInput  `json:"meals"`
}

type ListMyMealPlansReq struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Q     string `form:"q"`
}

type GenerateRoutineReq struct {
	DaysPerWeek    int      `json:"days_per_week" binding:"required,min=3,max=6"`
	Goal           string   `json:"goal" binding:"required,oneof=gain_muscle gain_strength lose_weight"`
	Difficulty     string   `json:"difficulty" binding:"required,oneof=novice intermediate advanced"`
	EquipmentPrefs []string `json:"equipment_prefs"`
	FocusMuscles   []string `json:"focus_muscles"`
	SessionMinutes int      `json:"session_minutes"`
	Notes          string   `json:"notes"`
	Title          string   `json:"title"`
}

type GenerateRoutineRes struct {
	RoutineID   uint   `json:"routine_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	WorkoutIDs  []uint `json:"workout_ids"`
	DayTitles   []string `json:"day_titles"`
}

type ChatReq struct {
	Message  string `json:"message" binding:"required,min=1,max=4000"`
	ThreadID *uint  `json:"thread_id"`
}

type ChatCitation struct {
	Type string `json:"type"` // exercise | food | routine
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ChatRes struct {
	ThreadID  uint           `json:"thread_id"`
	Reply     string         `json:"reply"`
	Citations []ChatCitation `json:"citations,omitempty"`
}

type ListThreadsRes struct {
	Data []ChatThreadRes `json:"data"`
}

type ChatThreadRes struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
}

type ChatMessagesRes struct {
	ThreadID uint              `json:"thread_id"`
	Messages []ChatMessageRes  `json:"messages"`
}

type ChatMessageRes struct {
	ID        uint   `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
