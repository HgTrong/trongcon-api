package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	aiv1 "trongcon-api/api/ai/v1"
	mealplanv1 "trongcon-api/api/meal_plan/v1"
	mytrainv1 "trongcon-api/api/my_train/v1"
	"trongcon-api/internal/entity"
	oaiclient "trongcon-api/internal/openai"
	"trongcon-api/internal/repository"

	goopenai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

var ErrAIUnavailable = errors.New("AI coach is not configured (missing OPENAI_API_KEY)")
var ErrAIChatThreadNotFound = errors.New("chat thread not found")
var ErrAIGenerationFailed = errors.New("AI generation failed — try again with clearer preferences")

type AICoachService interface {
	GenerateMealPlan(ctx context.Context, userID uint, req *aiv1.GenerateMealPlanReq) (*aiv1.GenerateMealPlanRes, error)
	GenerateRoutine(ctx context.Context, userID uint, req *aiv1.GenerateRoutineReq) (*aiv1.GenerateRoutineRes, error)
	Chat(ctx context.Context, userID uint, req *aiv1.ChatReq) (*aiv1.ChatRes, error)
	ListThreads(ctx context.Context, userID uint) (*aiv1.ListThreadsRes, error)
	ListMessages(ctx context.Context, userID, threadID uint) (*aiv1.ChatMessagesRes, error)
}

type aiCoachService struct {
	client       *oaiclient.Client
	foodRepo     repository.FoodRepository
	exerciseRepo repository.ExerciseRepository
	muscleRepo   repository.MuscleRepository
	mealPlanSvc  MealPlanService
	myTrainSvc   MyTrainService
	chatRepo     repository.AiChatRepository
}

func NewAICoachService(
	client *oaiclient.Client,
	foodRepo repository.FoodRepository,
	exerciseRepo repository.ExerciseRepository,
	muscleRepo repository.MuscleRepository,
	mealPlanSvc MealPlanService,
	myTrainSvc MyTrainService,
	chatRepo repository.AiChatRepository,
) AICoachService {
	return &aiCoachService{
		client:       client,
		foodRepo:     foodRepo,
		exerciseRepo: exerciseRepo,
		muscleRepo:   muscleRepo,
		mealPlanSvc:  mealPlanSvc,
		myTrainSvc:   myTrainSvc,
		chatRepo:     chatRepo,
	}
}

func (s *aiCoachService) requireClient() error {
	if s.client == nil || !s.client.Enabled() {
		return ErrAIUnavailable
	}
	return nil
}

type foodCandidate struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Calories     float64 `json:"calories"`
	Protein      float64 `json:"protein"`
	Carb         float64 `json:"carb"`
	Fat          float64 `json:"fat"`
	ServingSizeG float64 `json:"serving_size_g"`
}

type llmMealPlanOut struct {
	Title string `json:"title"`
	Meals []struct {
		Name  string `json:"name"`
		Items []struct {
			FoodID   uint    `json:"food_id"`
			Quantity float64 `json:"quantity"`
		} `json:"items"`
	} `json:"meals"`
}

func mealInputsCalories(meals []mealplanv1.MealPlanMealInput, calByFood map[uint]float64) float64 {
	var total float64
	for _, m := range meals {
		for _, it := range m.Items {
			total += calByFood[it.FoodID] * it.Quantity
		}
	}
	return total
}

// scaleMealInputsToCalories proportionally scales servings so day kcal ≈ target (± soft clamp).
func scaleMealInputsToCalories(meals []mealplanv1.MealPlanMealInput, calByFood map[uint]float64, target float64) []mealplanv1.MealPlanMealInput {
	if target <= 0 || len(meals) == 0 {
		return meals
	}
	total := mealInputsCalories(meals, calByFood)
	if total <= 0 {
		return meals
	}
	ratio := total / target
	// Within 8% — leave as-is
	if ratio >= 0.92 && ratio <= 1.08 {
		return meals
	}
	factor := target / total
	out := make([]mealplanv1.MealPlanMealInput, len(meals))
	for i, m := range meals {
		items := make([]mealplanv1.MealPlanItemInput, 0, len(m.Items))
		for _, it := range m.Items {
			qty := it.Quantity * factor
			// Round to 0.25 servings
			qty = float64(int(qty*4+0.5)) / 4
			if qty < 0.25 {
				qty = 0.25
			}
			if qty > 5 {
				qty = 5
			}
			items = append(items, mealplanv1.MealPlanItemInput{FoodID: it.FoodID, Quantity: qty})
		}
		out[i] = mealplanv1.MealPlanMealInput{Name: m.Name, Items: items}
	}
	return out
}

func foodLooksAllergic(name string, allergies []string) bool {
	n := strings.ToLower(name)
	for _, a := range allergies {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" {
			continue
		}
		if strings.Contains(n, a) {
			return true
		}
	}
	return false
}

func (s *aiCoachService) gatherFoodCandidates(ctx context.Context, req *aiv1.GenerateMealPlanReq) ([]foodCandidate, error) {
	queries := []string{""}
	if req.Diet != "" && req.Diet != "none" {
		queries = append(queries, req.Diet)
	}
	if req.Cuisine != "" {
		queries = append(queries, req.Cuisine)
	}
	for _, d := range req.DislikeFoods {
		_ = d // used only as negative filter
	}
	seed := []string{"chicken", "rice", "egg", "oat", "yogurt", "salmon", "beef", "potato", "broccoli", "banana", "milk", "tofu", "bean", "bread", "apple"}
	if req.Diet == "vegetarian" || req.Diet == "vegan" {
		seed = []string{"tofu", "bean", "lentil", "rice", "oat", "yogurt", "milk", "broccoli", "potato", "banana", "apple", "spinach", "chickpea", "quinoa"}
	}
	queries = append(queries, seed...)

	seen := map[uint]struct{}{}
	out := make([]foodCandidate, 0, 80)
	for _, q := range queries {
		if len(out) >= 80 {
			break
		}
		list, _, err := s.foodRepo.List(ctx, 0, 40, "name ASC", q)
		if err != nil {
			return nil, err
		}
		for _, f := range list {
			if _, ok := seen[f.ID]; ok {
				continue
			}
			if foodLooksAllergic(f.Name, req.Allergies) {
				continue
			}
			if foodLooksAllergic(f.Name, req.DislikeFoods) {
				continue
			}
			seen[f.ID] = struct{}{}
			out = append(out, foodCandidate{
				ID: f.ID, Name: f.Name, Calories: f.Calories, Protein: f.Protein,
				Carb: f.Carb, Fat: f.Fat, ServingSizeG: f.ServingSizeG,
			})
			if len(out) >= 80 {
				break
			}
		}
	}
	if len(out) < 10 {
		return nil, fmt.Errorf("%w: not enough foods in catalog (need at least 10)", ErrAIGenerationFailed)
	}
	return out, nil
}

func (s *aiCoachService) GenerateMealPlan(ctx context.Context, userID uint, req *aiv1.GenerateMealPlanReq) (*aiv1.GenerateMealPlanRes, error) {
	_ = userID
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	candidates, err := s.gatherFoodCandidates(ctx, req)
	if err != nil {
		return nil, err
	}
	candJSON, _ := json.Marshal(candidates)

	proteinHint := "derive from goal"
	if req.ProteinG != nil {
		proteinHint = fmt.Sprintf("%.0fg", *req.ProteinG)
	}
	carbHint, fatHint := "auto", "auto"
	if req.CarbG != nil {
		carbHint = fmt.Sprintf("%.0fg", *req.CarbG)
	}
	if req.FatG != nil {
		fatHint = fmt.Sprintf("%.0fg", *req.FatG)
	}

	system := `You are TrongCon nutrition coach. Build a 1-day meal plan using ONLY foods from the provided catalog.
HARD RULES:
- Every food_id MUST appear in the catalog.
- quantity = servings multiplier (use 0.5–3; avoid stacking calorie-dense foods).
- Compute running calories as sum(food.calories * quantity). FINAL DAY TOTAL MUST land within ±8% of the target calories. This is mandatory.
- Prefer lean proteins, vegetables, and moderate carbs. Avoid packing nuts, bacon, oil, cheese as large "free" volumes.
- Respect diet, allergies (already filtered), cuisine, and notes.
- Return JSON only: {"title":"...","meals":[{"name":"Breakfast","items":[{"food_id":1,"quantity":1.5}]}]}
- Exactly meals_per_day meals with clear names.
- Prefer ~calories/meals_per_day per meal (±20%).`

	userPrompt := fmt.Sprintf(`TARGET_CALORIES=%g (must hit this — NOT optional)
Protein target: %s
Carb target: %s
Fat target: %s
Meals per day: %d
Per-meal calorie guide: ~%.0f kcal
Goal: %s
Diet: %s
Cuisine: %s
Notes: %s
Each catalog row includes calories PER SERVING — multiply carefully.
Catalog (%d foods):
%s`,
		req.Calories, proteinHint, carbHint, fatHint, req.MealsPerDay, req.Calories/float64(req.MealsPerDay),
		req.Goal, req.Diet, req.Cuisine, req.Notes,
		len(candidates), string(candJSON))

	raw, err := s.client.ChatJSON(ctx, system, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIGenerationFailed, err)
	}
	var out llmMealPlanOut
	if err := oaiclient.DecodeJSON(raw, &out); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON", ErrAIGenerationFailed)
	}

	calByFood := map[uint]float64{}
	allowed := map[uint]struct{}{}
	for _, c := range candidates {
		allowed[c.ID] = struct{}{}
		calByFood[c.ID] = c.Calories
	}
	meals := make([]mealplanv1.MealPlanMealInput, 0, len(out.Meals))
	for _, m := range out.Meals {
		items := make([]mealplanv1.MealPlanItemInput, 0, len(m.Items))
		for _, it := range m.Items {
			if _, ok := allowed[it.FoodID]; !ok {
				continue
			}
			qty := it.Quantity
			if qty <= 0 {
				qty = 1
			}
			if qty > 4 {
				qty = 4
			}
			items = append(items, mealplanv1.MealPlanItemInput{FoodID: it.FoodID, Quantity: qty})
		}
		if len(items) == 0 {
			continue
		}
		name := strings.TrimSpace(m.Name)
		if name == "" {
			name = fmt.Sprintf("Meal %d", len(meals)+1)
		}
		meals = append(meals, mealplanv1.MealPlanMealInput{Name: name, Items: items})
	}
	if len(meals) == 0 {
		return nil, fmt.Errorf("%w: no valid foods mapped from catalog", ErrAIGenerationFailed)
	}

	meals = scaleMealInputsToCalories(meals, calByFood, req.Calories)

	title := strings.TrimSpace(out.Title)
	if req.Title != "" {
		title = strings.TrimSpace(req.Title)
	}
	if title == "" {
		title = fmt.Sprintf("AI plan · %.0f kcal · %s", req.Calories, req.Goal)
	}
	desc := fmt.Sprintf("Generated for %.0f kcal (%s). Diet: %s. %s", req.Calories, req.Goal, req.Diet, strings.TrimSpace(req.Notes))

	preview, err := s.mealPlanSvc.HydratePreview(ctx, title, desc, meals)
	if err != nil {
		return nil, err
	}
	return &aiv1.GenerateMealPlanRes{
		Preview:     *preview,
		CandidatesN: len(candidates),
		SaveHint:    "POST /user/my-meal-plans with title + meals from preview to persist",
	}, nil
}

type exerciseCandidate struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	Difficulty string   `json:"difficulty"`
	Equipment  string   `json:"equipment,omitempty"`
	Muscles    []string `json:"muscles,omitempty"`
}

type llmRoutineOut struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Days        []struct {
		WorkoutTitle string `json:"workout_title"`
		Items        []struct {
			ExerciseID uint   `json:"exercise_id"`
			Sets       int    `json:"sets"`
			Reps       string `json:"reps"`
		} `json:"items"`
	} `json:"days"`
}

// focusMuscleAliases expands casual labels (arms, shoulders…) into catalog-friendly tokens.
var focusMuscleAliases = map[string][]string{
	"arms":      {"bicep", "tricep", "forearm", "brachialis"},
	"arm":       {"bicep", "tricep", "forearm"},
	"shoulder":  {"shoulder", "delt"},
	"shoulders": {"shoulder", "delt"},
	"chest":     {"chest", "pec"},
	"back":      {"back", "lat", "trap", "rhomboid"},
	"legs":      {"quad", "hamstring", "glute", "calf", "leg"},
	"leg":       {"quad", "hamstring", "glute", "calf"},
	"core":      {"abs", "core", "oblique"},
	"abs":       {"abs", "core", "oblique"},
	"glutes":    {"glute"},
	"glute":     {"glute"},
}

func expandFocusTokens(raw []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(raw)*3)
	add := func(t string) {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	for _, want := range raw {
		w := strings.ToLower(strings.TrimSpace(want))
		add(w)
		if aliases, ok := focusMuscleAliases[w]; ok {
			for _, a := range aliases {
				add(a)
			}
		}
	}
	return out
}

func (s *aiCoachService) gatherExerciseCandidates(ctx context.Context, req *aiv1.GenerateRoutineReq) ([]exerciseCandidate, error) {
	tokens := expandFocusTokens(req.FocusMuscles)
	var muscleIDs []uint
	if len(tokens) > 0 {
		muscles, _, err := s.muscleRepo.List(ctx, 0, 200, "name ASC", "")
		if err != nil {
			return nil, err
		}
		for _, m := range muscles {
			name := strings.ToLower(m.Name)
			slug := strings.ToLower(m.Slug)
			region := strings.ToLower(m.Region)
			for _, t := range tokens {
				if strings.Contains(name, t) || strings.Contains(slug, t) || strings.Contains(region, t) {
					muscleIDs = append(muscleIDs, m.ID)
					break
				}
			}
		}
	}

	seen := map[uint]struct{}{}
	out := make([]exerciseCandidate, 0, 120)

	addList := func(list []entity.Exercise) {
		for _, ex := range list {
			if _, ok := seen[ex.ID]; ok {
				continue
			}
			if ex.Status != "" && ex.Status != "active" {
				continue
			}
			seen[ex.ID] = struct{}{}
			eqName := ""
			if ex.Equipment != nil {
				eqName = ex.Equipment.Name
			}
			muscles := make([]string, 0, len(ex.Muscles))
			for _, em := range ex.Muscles {
				if em.Muscle.Name != "" {
					muscles = append(muscles, em.Muscle.Name)
				}
			}
			out = append(out, exerciseCandidate{
				ID: ex.ID, Name: ex.Name, Difficulty: ex.Difficulty,
				Equipment: eqName, Muscles: muscles,
			})
		}
	}

	// Do NOT hard-filter by difficulty — prefer matching later, but always
	// pull a wide active pool so gen works on small catalogs.
	fetch := func(muscleID *uint, q, difficulty string, limit int) error {
		list, _, err := s.exerciseRepo.List(ctx, 0, limit, "name ASC", q, difficulty, "", "", "active", nil, muscleID)
		if err != nil {
			return err
		}
		addList(list)
		return nil
	}

	// 1) Preferred difficulty first (soft preference via order of insertion).
	if err := fetch(nil, "", req.Difficulty, 80); err != nil {
		return nil, err
	}
	// 2) Broader pool without difficulty.
	if err := fetch(nil, "", "", 120); err != nil {
		return nil, err
	}
	// 3) Focus muscles (any difficulty).
	for _, mid := range muscleIDs {
		m := mid
		if err := fetch(&m, "", "", 60); err != nil {
			return nil, err
		}
		if len(out) >= 150 {
			break
		}
	}
	// 4) Equipment keyword search (name/summary).
	for _, eq := range req.EquipmentPrefs {
		if err := fetch(nil, strings.TrimSpace(eq), "", 40); err != nil {
			return nil, err
		}
	}

	// Soft filter by equipment prefs — only apply when enough remain.
	if len(req.EquipmentPrefs) > 0 {
		filtered := make([]exerciseCandidate, 0, len(out))
		for _, c := range out {
			el := strings.ToLower(c.Equipment + " " + c.Name)
			ok := false
			for _, p := range req.EquipmentPrefs {
				if strings.Contains(el, strings.ToLower(strings.TrimSpace(p))) {
					ok = true
					break
				}
			}
			if ok || c.Equipment == "" || strings.Contains(el, "body") {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) >= 12 {
			out = filtered
		}
		// else keep the full pool — prefs are hints, not hard requirements
	}

	// Prefer requested difficulty near the front for the LLM context window.
	if req.Difficulty != "" && len(out) > 1 {
		pref := make([]exerciseCandidate, 0, len(out))
		rest := make([]exerciseCandidate, 0, len(out))
		for _, c := range out {
			if strings.EqualFold(c.Difficulty, req.Difficulty) {
				pref = append(pref, c)
			} else {
				rest = append(rest, c)
			}
		}
		out = append(pref, rest...)
	}

	if len(out) > 100 {
		out = out[:100]
	}
	if len(out) < 6 {
		return nil, fmt.Errorf("%w: catalog has too few active exercises (need at least 6)", ErrAIGenerationFailed)
	}
	return out, nil
}

func (s *aiCoachService) GenerateRoutine(ctx context.Context, userID uint, req *aiv1.GenerateRoutineReq) (*aiv1.GenerateRoutineRes, error) {
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	candidates, err := s.gatherExerciseCandidates(ctx, req)
	if err != nil {
		return nil, err
	}
	candJSON, _ := json.Marshal(candidates)
	minutes := req.SessionMinutes
	if minutes <= 0 {
		minutes = 60
	}

	system := `You are TrongCon strength coach. Build a weekly training split using ONLY exercises from the catalog.
Rules:
- Every exercise_id MUST be from the catalog JSON (use the id field exactly).
- days length MUST equal days_per_week.
- Each day: 4–8 exercises, sets 2–5, reps as string like "8-12" or "5".
- Prefer the requested difficulty when possible, but you MAY use nearby difficulties if needed.
- Prioritize focus muscles and equipment prefs as soft preferences — do not invent exercises.
- Session length is a hint for volume, not a hard stop.
- Return JSON: {"title":"...","description":"...","days":[{"workout_title":"Push","items":[{"exercise_id":1,"sets":3,"reps":"8-12"}]}]}`

	userPrompt := fmt.Sprintf(`Days/week: %d
Goal: %s
Difficulty: %s
Session minutes: %d
Equipment prefs: %v
Focus muscles: %v
Notes: %s
Catalog (%d exercises):
%s`,
		req.DaysPerWeek, req.Goal, req.Difficulty, minutes, req.EquipmentPrefs, req.FocusMuscles, req.Notes,
		len(candidates), string(candJSON))

	raw, err := s.client.ChatJSON(ctx, system, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIGenerationFailed, err)
	}
	var out llmRoutineOut
	if err := oaiclient.DecodeJSON(raw, &out); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON", ErrAIGenerationFailed)
	}
	if len(out.Days) == 0 {
		return nil, fmt.Errorf("%w: empty days", ErrAIGenerationFailed)
	}

	allowed := map[uint]struct{}{}
	for _, c := range candidates {
		allowed[c.ID] = struct{}{}
	}

	title := strings.TrimSpace(out.Title)
	if req.Title != "" {
		title = strings.TrimSpace(req.Title)
	}
	if title == "" {
		title = fmt.Sprintf("AI %d-day · %s", req.DaysPerWeek, req.Goal)
	}
	desc := strings.TrimSpace(out.Description)
	if desc == "" {
		desc = fmt.Sprintf("AI program · %s · %s", req.Goal, req.Difficulty)
	}

	workoutIDs := make([]uint, 0, len(out.Days))
	dayTitles := make([]string, 0, len(out.Days))
	routineItems := make([]mytrainv1.RoutineItemInput, 0, len(out.Days))

	for i, day := range out.Days {
		items := make([]mytrainv1.WorkoutItemInput, 0, len(day.Items))
		for _, it := range day.Items {
			if _, ok := allowed[it.ExerciseID]; !ok {
				continue
			}
			sets := it.Sets
			if sets < 1 {
				sets = 3
			}
			if sets > 6 {
				sets = 6
			}
			reps := strings.TrimSpace(it.Reps)
			if reps == "" {
				reps = "10"
			}
			items = append(items, mytrainv1.WorkoutItemInput{
				ExerciseID: it.ExerciseID,
				Sets:       sets,
				Reps:       reps,
			})
		}
		if len(items) == 0 {
			continue
		}
		wTitle := strings.TrimSpace(day.WorkoutTitle)
		if wTitle == "" {
			wTitle = fmt.Sprintf("%s · Day %d", title, i+1)
		}
		created, err := s.myTrainSvc.CreateWorkout(ctx, userID, &mytrainv1.CreateMyWorkoutReq{
			Title:      wTitle,
			Difficulty: req.Difficulty,
			Goal:       req.Goal,
			Items:      items,
		})
		if err != nil {
			return nil, err
		}
		workoutIDs = append(workoutIDs, created.Workout.ID)
		dayTitles = append(dayTitles, wTitle)
		routineItems = append(routineItems, mytrainv1.RoutineItemInput{WorkoutID: created.Workout.ID})
	}
	if len(routineItems) == 0 {
		return nil, fmt.Errorf("%w: no valid exercises mapped", ErrAIGenerationFailed)
	}

	createdR, err := s.myTrainSvc.CreateRoutine(ctx, userID, &mytrainv1.CreateRoutineReq{
		Title:       title,
		Description: desc,
		Difficulty:  req.Difficulty,
		IsPublic:    false,
		Items:       routineItems,
	})
	if err != nil {
		return nil, err
	}

	return &aiv1.GenerateRoutineRes{
		RoutineID:   createdR.Routine.ID,
		Title:       createdR.Routine.Title,
		Description: createdR.Routine.Description,
		Difficulty:  createdR.Routine.Difficulty,
		WorkoutIDs:  workoutIDs,
		DayTitles:   dayTitles,
	}, nil
}

func (s *aiCoachService) Chat(ctx context.Context, userID uint, req *aiv1.ChatReq) (*aiv1.ChatRes, error) {
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		return nil, errors.New("message is required")
	}

	var thread *entity.AiChatThread
	var err error
	if req.ThreadID != nil && *req.ThreadID > 0 {
		thread, err = s.chatRepo.GetThread(ctx, userID, *req.ThreadID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrAIChatThreadNotFound
			}
			return nil, err
		}
	} else {
		title := msg
		if utf8.RuneCountInString(title) > 80 {
			title = string([]rune(title)[:80]) + "…"
		}
		thread = &entity.AiChatThread{UserID: userID, Title: title}
		if err := s.chatRepo.CreateThread(ctx, thread); err != nil {
			return nil, err
		}
	}

	if err := s.chatRepo.AddMessage(ctx, &entity.AiChatMessage{
		ThreadID: thread.ID,
		Role:     "user",
		Content:  msg,
	}); err != nil {
		return nil, err
	}

	history, err := s.chatRepo.ListMessages(ctx, thread.ID, 30)
	if err != nil {
		return nil, err
	}

	messages := []goopenai.ChatCompletionMessage{
		{
			Role: goopenai.ChatMessageRoleSystem,
			Content: `You are TrongCon AI Coach — a practical training & nutrition assistant.
Use tools to look up real exercises, foods, and the user's routines from TrongCon catalog.
When you mention an exercise or food from tools, include its exact name and id like [exercise:#123 Name] or [food:#45 Name].
Be concise, actionable, and safety-aware. Reply in the user's language.`,
		},
	}
	for _, h := range history {
		role := goopenai.ChatMessageRoleUser
		if h.Role == "assistant" {
			role = goopenai.ChatMessageRoleAssistant
		} else if h.Role == "system" {
			role = goopenai.ChatMessageRoleSystem
		}
		messages = append(messages, goopenai.ChatCompletionMessage{Role: role, Content: h.Content})
	}

	tools := []goopenai.Tool{
		{
			Type: goopenai.ToolTypeFunction,
			Function: &goopenai.FunctionDefinition{
				Name:        "search_exercises",
				Description: "Search TrongCon exercise catalog by keyword / difficulty",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"q":          map[string]any{"type": "string"},
						"difficulty": map[string]any{"type": "string", "enum": []string{"novice", "intermediate", "advanced", ""}},
						"limit":      map[string]any{"type": "integer"},
					},
				},
			},
		},
		{
			Type: goopenai.ToolTypeFunction,
			Function: &goopenai.FunctionDefinition{
				Name:        "get_exercise",
				Description: "Get one exercise by id",
				Parameters: map[string]any{
					"type":     "object",
					"required": []string{"id"},
					"properties": map[string]any{
						"id": map[string]any{"type": "integer"},
					},
				},
			},
		},
		{
			Type: goopenai.ToolTypeFunction,
			Function: &goopenai.FunctionDefinition{
				Name:        "search_foods",
				Description: "Search TrongCon food catalog for macros",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"q":     map[string]any{"type": "string"},
						"limit": map[string]any{"type": "integer"},
					},
				},
			},
		},
		{
			Type: goopenai.ToolTypeFunction,
			Function: &goopenai.FunctionDefinition{
				Name:        "list_my_routines",
				Description: "List the current user's personal training routines",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}

	citations := make([]aiv1.ChatCitation, 0)
	citeSeen := map[string]struct{}{}
	addCite := func(typ string, id uint, name string) {
		key := fmt.Sprintf("%s:%d", typ, id)
		if _, ok := citeSeen[key]; ok {
			return
		}
		citeSeen[key] = struct{}{}
		citations = append(citations, aiv1.ChatCitation{Type: typ, ID: id, Name: name})
	}

	reply, err := s.client.ChatWithTools(ctx, messages, tools, func(name, args string) (string, error) {
		switch name {
		case "search_exercises":
			var p struct {
				Q          string `json:"q"`
				Difficulty string `json:"difficulty"`
				Limit      int    `json:"limit"`
			}
			_ = json.Unmarshal([]byte(args), &p)
			if p.Limit <= 0 || p.Limit > 15 {
				p.Limit = 10
			}
			list, _, err := s.exerciseRepo.List(ctx, 0, p.Limit, "views DESC", p.Q, p.Difficulty, "", "", "active", nil, nil)
			if err != nil {
				return "", err
			}
			type row struct {
				ID         uint   `json:"id"`
				Name       string `json:"name"`
				Difficulty string `json:"difficulty"`
				Summary    string `json:"summary"`
			}
			rows := make([]row, 0, len(list))
			for _, ex := range list {
				rows = append(rows, row{ID: ex.ID, Name: ex.Name, Difficulty: ex.Difficulty, Summary: ex.Summary})
				addCite("exercise", ex.ID, ex.Name)
			}
			b, _ := json.Marshal(rows)
			return string(b), nil

		case "get_exercise":
			var p struct {
				ID uint `json:"id"`
			}
			_ = json.Unmarshal([]byte(args), &p)
			ex, err := s.exerciseRepo.GetByID(ctx, p.ID)
			if err != nil {
				return `{"error":"not found"}`, nil
			}
			addCite("exercise", ex.ID, ex.Name)
			eq := ""
			if ex.Equipment != nil {
				eq = ex.Equipment.Name
			}
			muscles := make([]string, 0)
			for _, em := range ex.Muscles {
				if em.Muscle.Name != "" {
					muscles = append(muscles, em.Muscle.Name+" ("+em.Role+")")
				}
			}
			b, _ := json.Marshal(map[string]any{
				"id": ex.ID, "name": ex.Name, "difficulty": ex.Difficulty,
				"summary": ex.Summary, "equipment": eq, "muscles": muscles,
			})
			return string(b), nil

		case "search_foods":
			var p struct {
				Q     string `json:"q"`
				Limit int    `json:"limit"`
			}
			_ = json.Unmarshal([]byte(args), &p)
			if p.Limit <= 0 || p.Limit > 15 {
				p.Limit = 10
			}
			list, _, err := s.foodRepo.List(ctx, 0, p.Limit, "name ASC", p.Q)
			if err != nil {
				return "", err
			}
			type row struct {
				ID       uint    `json:"id"`
				Name     string  `json:"name"`
				Calories float64 `json:"calories"`
				Protein  float64 `json:"protein"`
				Carb     float64 `json:"carb"`
				Fat      float64 `json:"fat"`
			}
			rows := make([]row, 0, len(list))
			for _, f := range list {
				rows = append(rows, row{ID: f.ID, Name: f.Name, Calories: f.Calories, Protein: f.Protein, Carb: f.Carb, Fat: f.Fat})
				addCite("food", f.ID, f.Name)
			}
			b, _ := json.Marshal(rows)
			return string(b), nil

		case "list_my_routines":
			res, err := s.myTrainSvc.ListRoutines(ctx, userID, &mytrainv1.ListMyRoutinesReq{Page: 1, Limit: 10})
			if err != nil {
				return "", err
			}
			type row struct {
				ID    uint   `json:"id"`
				Title string `json:"title"`
			}
			rows := make([]row, 0, len(res.Data))
			for _, r := range res.Data {
				rows = append(rows, row{ID: r.ID, Title: r.Title})
				addCite("routine", r.ID, r.Title)
			}
			b, _ := json.Marshal(rows)
			return string(b), nil
		}
		return `{"error":"unknown tool"}`, nil
	}, 3)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIGenerationFailed, err)
	}
	if reply == "" {
		reply = "I couldn't find a clear answer — try asking about a specific exercise, food, or your routine."
	}

	if err := s.chatRepo.AddMessage(ctx, &entity.AiChatMessage{
		ThreadID: thread.ID,
		Role:     "assistant",
		Content:  reply,
	}); err != nil {
		return nil, err
	}

	return &aiv1.ChatRes{
		ThreadID:  thread.ID,
		Reply:     reply,
		Citations: citations,
	}, nil
}

func (s *aiCoachService) ListThreads(ctx context.Context, userID uint) (*aiv1.ListThreadsRes, error) {
	list, err := s.chatRepo.ListThreads(ctx, userID, 30)
	if err != nil {
		return nil, err
	}
	data := make([]aiv1.ChatThreadRes, 0, len(list))
	for _, t := range list {
		data = append(data, aiv1.ChatThreadRes{
			ID:        t.ID,
			Title:     t.Title,
			UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		})
	}
	return &aiv1.ListThreadsRes{Data: data}, nil
}

func (s *aiCoachService) ListMessages(ctx context.Context, userID, threadID uint) (*aiv1.ChatMessagesRes, error) {
	if _, err := s.chatRepo.GetThread(ctx, userID, threadID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAIChatThreadNotFound
		}
		return nil, err
	}
	list, err := s.chatRepo.ListMessages(ctx, threadID, 100)
	if err != nil {
		return nil, err
	}
	msgs := make([]aiv1.ChatMessageRes, 0, len(list))
	for _, m := range list {
		msgs = append(msgs, aiv1.ChatMessageRes{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}
	return &aiv1.ChatMessagesRes{ThreadID: threadID, Messages: msgs}, nil
}
