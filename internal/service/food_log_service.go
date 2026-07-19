package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	foodlogv1 "trongcon-api/api/food_log/v1"
	toolsv1 "trongcon-api/api/tools/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrFoodLogEntryNotFound = errors.New("food log entry not found")
	ErrFoodLogMealNotFound  = errors.New("food log meal not found")
	ErrMealNotEmpty         = errors.New("meal has logged foods and cannot be removed")
	ErrInvalidLogDate       = errors.New("invalid date format, use YYYY-MM-DD")
)

const defaultMealCount = 3

type FoodLogService interface {
	GetGoals(ctx context.Context, userID uint) (*foodlogv1.NutritionGoalRes, error)
	UpdateGoals(ctx context.Context, userID uint, req *foodlogv1.UpdateGoalsReq) (*foodlogv1.NutritionGoalRes, error)
	GetDay(ctx context.Context, userID uint, dateStr string) (*foodlogv1.DayLogRes, error)
	CreateMeal(ctx context.Context, userID uint, req *foodlogv1.CreateMealReq) (*foodlogv1.MealOnlyRes, error)
	UpdateMeal(ctx context.Context, userID, mealID uint, req *foodlogv1.UpdateMealReq) (*foodlogv1.MealOnlyRes, error)
	DeleteMeal(ctx context.Context, userID, mealID uint) error
	CreateEntry(ctx context.Context, userID uint, req *foodlogv1.CreateEntryReq) (*foodlogv1.FoodLogEntryRes, error)
	UpdateEntry(ctx context.Context, userID, entryID uint, req *foodlogv1.UpdateEntryReq) (*foodlogv1.FoodLogEntryRes, error)
	DeleteEntry(ctx context.Context, userID, entryID uint) error
	ListRecent(ctx context.Context, userID uint) (*foodlogv1.RecentRes, error)
	SaveFromCalories(ctx context.Context, userID uint, req *foodlogv1.SaveFromCaloriesReq) (*foodlogv1.NutritionGoalRes, error)
	GetMemberStats(ctx context.Context, userID uint) (*foodlogv1.MemberStatsRes, error)
}

type foodLogService struct {
	repo     repository.FoodLogRepository
	foodRepo repository.FoodRepository
	macroSvc MacroService
}

func NewFoodLogService(repo repository.FoodLogRepository, foodRepo repository.FoodRepository, macroSvc MacroService) FoodLogService {
	return &foodLogService{repo: repo, foodRepo: foodRepo, macroSvc: macroSvc}
}

func defaultGoals() foodlogv1.NutritionGoalRes {
	return foodlogv1.NutritionGoalRes{
		DailyCalories: 2200,
		DailyProteinG: 165,
		DailyCarbG:    220,
		DailyFatG:     73,
	}
}

func parseLogDate(dateStr string) (time.Time, error) {
	d, err := time.Parse("2006-01-02", strings.TrimSpace(dateStr))
	if err != nil {
		return time.Time{}, ErrInvalidLogDate
	}
	return d, nil
}

func normalizeFoodQty(q float64) float64 {
	if q <= 0 {
		return 1
	}
	return q
}

func snapshotFromFood(food *entity.Food, qty float64) entity.FoodLogEntry {
	q := normalizeFoodQty(qty)
	return entity.FoodLogEntry{
		FoodID:       food.ID,
		FoodName:     food.Name,
		Quantity:     q,
		ServingSizeG: food.ServingSizeG * q,
		Protein:      food.Protein * q,
		Carb:         food.Carb * q,
		Fat:          food.Fat * q,
		Calories:     food.Calories * q,
	}
}

func toEntryRes(e *entity.FoodLogEntry) foodlogv1.FoodLogEntryRes {
	return foodlogv1.FoodLogEntryRes{
		ID:           e.ID,
		FoodID:       e.FoodID,
		FoodName:     e.FoodName,
		MealID:       e.MealID,
		Quantity:     e.Quantity,
		ServingSizeG: e.ServingSizeG,
		Protein:      e.Protein,
		Carb:         e.Carb,
		Fat:          e.Fat,
		Calories:     e.Calories,
		CreatedAt:    e.CreatedAt,
	}
}

func toMealOnlyRes(m *entity.FoodLogMeal) foodlogv1.MealOnlyRes {
	return foodlogv1.MealOnlyRes{
		ID:        m.ID,
		Name:      m.Name,
		SortOrder: m.SortOrder,
	}
}

func sumEntries(entries []entity.FoodLogEntry) foodlogv1.MacroTotals {
	var t foodlogv1.MacroTotals
	for _, e := range entries {
		t.Calories += e.Calories
		t.ProteinG += e.Protein
		t.CarbG += e.Carb
		t.FatG += e.Fat
	}
	return t
}

func subtractMacros(goal, total foodlogv1.MacroTotals) foodlogv1.MacroTotals {
	return foodlogv1.MacroTotals{
		Calories: goal.Calories - total.Calories,
		ProteinG: goal.ProteinG - total.ProteinG,
		CarbG:    goal.CarbG - total.CarbG,
		FatG:     goal.FatG - total.FatG,
	}
}

func goalsToMacros(g foodlogv1.NutritionGoalRes) foodlogv1.MacroTotals {
	return foodlogv1.MacroTotals{
		Calories: g.DailyCalories,
		ProteinG: g.DailyProteinG,
		CarbG:    g.DailyCarbG,
		FatG:     g.DailyFatG,
	}
}

func (s *foodLogService) resolveGoals(ctx context.Context, userID uint) foodlogv1.NutritionGoalRes {
	goal, err := s.repo.GetGoals(ctx, userID)
	if err != nil || goal == nil {
		return defaultGoals()
	}
	return foodlogv1.NutritionGoalRes{
		DailyCalories: goal.DailyCalories,
		DailyProteinG: goal.DailyProteinG,
		DailyCarbG:    goal.DailyCarbG,
		DailyFatG:     goal.DailyFatG,
	}
}

func (s *foodLogService) ensureDayMeals(ctx context.Context, userID uint, logDate time.Time) ([]entity.FoodLogMeal, error) {
	meals, err := s.repo.ListMealsByDate(ctx, userID, logDate)
	if err != nil {
		return nil, err
	}
	if len(meals) > 0 {
		return meals, nil
	}

	template, err := s.repo.LatestMealTemplate(ctx, userID, logDate)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, defaultMealCount)
	if len(template) > 0 {
		for _, m := range template {
			names = append(names, m.Name)
		}
	} else {
		for i := 1; i <= defaultMealCount; i++ {
			names = append(names, fmt.Sprintf("Meal %d", i))
		}
	}

	created := make([]entity.FoodLogMeal, 0, len(names))
	for i, name := range names {
		meal := entity.FoodLogMeal{
			UserID:    userID,
			LogDate:   logDate,
			Name:      name,
			SortOrder: i,
		}
		if err := s.repo.CreateMeal(ctx, &meal); err != nil {
			return nil, err
		}
		created = append(created, meal)
	}
	return created, nil
}

func buildNutritionHints(goals foodlogv1.NutritionGoalRes, totals, remaining foodlogv1.MacroTotals) []foodlogv1.NutritionHintRes {
	var hints []foodlogv1.NutritionHintRes

	if totals.Calories <= 0 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "info",
			Macro:   "general",
			Message: "Start logging your first meal — consistency builds your streak.",
		})
		return hints
	}

	if remaining.ProteinG > goals.DailyProteinG*0.12 && remaining.ProteinG >= 15 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "warning",
			Macro:   "protein",
			Message: fmt.Sprintf("You're short ~%dg protein today. Lean meat, eggs, Greek yogurt, or whey can close the gap.", int(remaining.ProteinG)),
		})
	}

	if remaining.CarbG > goals.DailyCarbG*0.15 && remaining.CarbG >= 20 && totals.Calories >= goals.DailyCalories*0.35 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "info",
			Macro:   "carbs",
			Message: fmt.Sprintf("Still need ~%dg carbs — rice, oats, or fruit work well around training.", int(remaining.CarbG)),
		})
	}

	if remaining.Calories > goals.DailyCalories*0.2 && remaining.Calories > 200 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "info",
			Macro:   "calories",
			Message: fmt.Sprintf("%d kcal remaining — room for another solid meal.", int(remaining.Calories)),
		})
	}

	if remaining.Calories < 0 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "warning",
			Macro:   "calories",
			Message: fmt.Sprintf("You're %d kcal over today's target. Lighter portions or lean protein may help balance the rest of the day.", int(-remaining.Calories)),
		})
	}

	if remaining.FatG < -8 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "warning",
			Macro:   "fat",
			Message: fmt.Sprintf("Fat is %dg over goal — consider leaner protein sources for remaining meals.", int(-remaining.FatG)),
		})
	}

	if totals.ProteinG >= goals.DailyProteinG && remaining.Calories > 80 && remaining.Calories < goals.DailyCalories*0.35 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "success",
			Macro:   "protein",
			Message: "Protein goal hit — nice work. Use remaining calories for carbs and fats as planned.",
		})
	}

	if len(hints) == 0 {
		hints = append(hints, foodlogv1.NutritionHintRes{
			Type:    "success",
			Macro:   "general",
			Message: "On track with today's targets. Keep logging to build your streak.",
		})
	}

	return hints
}

func (s *foodLogService) GetGoals(ctx context.Context, userID uint) (*foodlogv1.NutritionGoalRes, error) {
	g := s.resolveGoals(ctx, userID)
	return &g, nil
}

func (s *foodLogService) UpdateGoals(ctx context.Context, userID uint, req *foodlogv1.UpdateGoalsReq) (*foodlogv1.NutritionGoalRes, error) {
	goal := &entity.UserNutritionGoal{
		UserID:        userID,
		DailyCalories: req.DailyCalories,
		DailyProteinG: req.DailyProteinG,
		DailyCarbG:    req.DailyCarbG,
		DailyFatG:     req.DailyFatG,
	}
	if err := s.repo.UpsertGoals(ctx, goal); err != nil {
		return nil, err
	}
	res := foodlogv1.NutritionGoalRes{
		DailyCalories: goal.DailyCalories,
		DailyProteinG: goal.DailyProteinG,
		DailyCarbG:    goal.DailyCarbG,
		DailyFatG:     goal.DailyFatG,
	}
	return &res, nil
}

func (s *foodLogService) GetDay(ctx context.Context, userID uint, dateStr string) (*foodlogv1.DayLogRes, error) {
	logDate, err := parseLogDate(dateStr)
	if err != nil {
		return nil, err
	}

	meals, err := s.ensureDayMeals(ctx, userID, logDate)
	if err != nil {
		return nil, err
	}

	entries, err := s.repo.ListByDate(ctx, userID, logDate)
	if err != nil {
		return nil, err
	}

	entriesByMeal := make(map[uint][]entity.FoodLogEntry)
	for i := range entries {
		e := entries[i]
		entriesByMeal[e.MealID] = append(entriesByMeal[e.MealID], e)
	}

	goals := s.resolveGoals(ctx, userID)
	goalMacros := goalsToMacros(goals)
	totals := sumEntries(entries)
	remaining := subtractMacros(goalMacros, totals)

	mealRes := make([]foodlogv1.MealRes, 0, len(meals))
	for _, m := range meals {
		list := entriesByMeal[m.ID]
		resEntries := make([]foodlogv1.FoodLogEntryRes, 0, len(list))
		for i := range list {
			resEntries = append(resEntries, toEntryRes(&list[i]))
		}
		mealRes = append(mealRes, foodlogv1.MealRes{
			ID:        m.ID,
			Name:      m.Name,
			SortOrder: m.SortOrder,
			Entries:   resEntries,
			Totals:    sumEntries(list),
		})
	}

	sort.Slice(mealRes, func(i, j int) bool {
		return mealRes[i].SortOrder < mealRes[j].SortOrder
	})

	return &foodlogv1.DayLogRes{
		Date:        logDate.Format("2006-01-02"),
		Goals:       goals,
		Totals:      totals,
		Remaining:   remaining,
		Meals:       mealRes,
		Suggestions: buildNutritionHints(goals, totals, remaining),
	}, nil
}

func (s *foodLogService) CreateMeal(ctx context.Context, userID uint, req *foodlogv1.CreateMealReq) (*foodlogv1.MealOnlyRes, error) {
	logDate, err := parseLogDate(req.Date)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.ListMealsByDate(ctx, userID, logDate)
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		existing, err = s.ensureDayMeals(ctx, userID, logDate)
		if err != nil {
			return nil, err
		}
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("Meal %d", len(existing)+1)
	}

	meal := entity.FoodLogMeal{
		UserID:    userID,
		LogDate:   logDate,
		Name:      name,
		SortOrder: len(existing),
	}
	if err := s.repo.CreateMeal(ctx, &meal); err != nil {
		return nil, err
	}
	res := toMealOnlyRes(&meal)
	return &res, nil
}

func (s *foodLogService) UpdateMeal(ctx context.Context, userID, mealID uint, req *foodlogv1.UpdateMealReq) (*foodlogv1.MealOnlyRes, error) {
	meal, err := s.repo.GetMeal(ctx, userID, mealID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFoodLogMealNotFound
		}
		return nil, err
	}
	meal.Name = strings.TrimSpace(req.Name)
	if err := s.repo.UpdateMeal(ctx, meal); err != nil {
		return nil, err
	}
	res := toMealOnlyRes(meal)
	return &res, nil
}

func (s *foodLogService) DeleteMeal(ctx context.Context, userID, mealID uint) error {
	_, err := s.repo.GetMeal(ctx, userID, mealID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFoodLogMealNotFound
		}
		return err
	}
	count, err := s.repo.CountEntriesInMeal(ctx, userID, mealID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrMealNotEmpty
	}
	return s.repo.DeleteMeal(ctx, userID, mealID)
}

func (s *foodLogService) CreateEntry(ctx context.Context, userID uint, req *foodlogv1.CreateEntryReq) (*foodlogv1.FoodLogEntryRes, error) {
	logDate, err := parseLogDate(req.Date)
	if err != nil {
		return nil, err
	}

	meal, err := s.repo.GetMeal(ctx, userID, req.MealID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFoodLogMealNotFound
		}
		return nil, err
	}
	if meal.LogDate.Format("2006-01-02") != logDate.Format("2006-01-02") {
		return nil, ErrFoodLogMealNotFound
	}

	food, err := s.foodRepo.GetByID(ctx, req.FoodID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFoodNotFound
		}
		return nil, err
	}

	entry := snapshotFromFood(food, req.Quantity)
	entry.UserID = userID
	entry.LogDate = logDate
	entry.MealID = req.MealID

	if err := s.repo.CreateEntry(ctx, &entry); err != nil {
		return nil, err
	}
	res := toEntryRes(&entry)
	return &res, nil
}

func (s *foodLogService) UpdateEntry(ctx context.Context, userID, entryID uint, req *foodlogv1.UpdateEntryReq) (*foodlogv1.FoodLogEntryRes, error) {
	entry, err := s.repo.GetEntry(ctx, userID, entryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFoodLogEntryNotFound
		}
		return nil, err
	}

	if req.MealID != nil {
		meal, err := s.repo.GetMeal(ctx, userID, *req.MealID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrFoodLogMealNotFound
			}
			return nil, err
		}
		if meal.LogDate.Format("2006-01-02") != entry.LogDate.Format("2006-01-02") {
			return nil, ErrFoodLogMealNotFound
		}
		entry.MealID = *req.MealID
	}
	if req.Quantity != nil {
		food, err := s.foodRepo.GetByID(ctx, entry.FoodID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrFoodNotFound
			}
			return nil, err
		}
		snap := snapshotFromFood(food, *req.Quantity)
		entry.Quantity = snap.Quantity
		entry.ServingSizeG = snap.ServingSizeG
		entry.Protein = snap.Protein
		entry.Carb = snap.Carb
		entry.Fat = snap.Fat
		entry.Calories = snap.Calories
	}

	if err := s.repo.UpdateEntry(ctx, entry); err != nil {
		return nil, err
	}
	res := toEntryRes(entry)
	return &res, nil
}

func (s *foodLogService) DeleteEntry(ctx context.Context, userID, entryID uint) error {
	_, err := s.repo.GetEntry(ctx, userID, entryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFoodLogEntryNotFound
		}
		return err
	}
	return s.repo.DeleteEntry(ctx, userID, entryID)
}

func (s *foodLogService) ListRecent(ctx context.Context, userID uint) (*foodlogv1.RecentRes, error) {
	list, err := s.repo.ListRecentFoods(ctx, userID, 12)
	if err != nil {
		return nil, err
	}
	data := make([]foodlogv1.RecentFoodRes, 0, len(list))
	for i := range list {
		e := list[i]
		q := normalizeFoodQty(e.Quantity)
		data = append(data, foodlogv1.RecentFoodRes{
			FoodID:       e.FoodID,
			FoodName:     e.FoodName,
			ServingSizeG: e.ServingSizeG / q,
			Protein:      e.Protein / q,
			Carb:         e.Carb / q,
			Fat:          e.Fat / q,
			Calories:     e.Calories / q,
			LastLoggedAt: e.UpdatedAt,
		})
	}
	return &foodlogv1.RecentRes{Data: data}, nil
}

func (s *foodLogService) SaveFromCalories(ctx context.Context, userID uint, req *foodlogv1.SaveFromCaloriesReq) (*foodlogv1.NutritionGoalRes, error) {
	preset := req.Preset
	if preset == "" {
		preset = "balanced"
	}
	macros, err := s.macroSvc.Calculate(&toolsv1.MacroCalculateReq{
		Preset:        preset,
		DailyCalories: req.DailyCalories,
		MealsPerDay:   3,
	})
	if err != nil {
		return nil, err
	}
	return s.UpdateGoals(ctx, userID, &foodlogv1.UpdateGoalsReq{
		DailyCalories: macros.DailyCalories,
		DailyProteinG: macros.DailyProteinG,
		DailyCarbG:    macros.DailyCarbsG,
		DailyFatG:     macros.DailyFatG,
	})
}

func dateKey(t time.Time) string {
	return t.Format("2006-01-02")
}

func hasLoggedDate(dateSet map[string]bool, t time.Time) bool {
	return dateSet[dateKey(t)]
}

func computeCurrentStreak(dateSet map[string]bool, today time.Time) int {
	start := today
	if !hasLoggedDate(dateSet, today) {
		start = today.AddDate(0, 0, -1)
	}
	streak := 0
	for d := start; hasLoggedDate(dateSet, d); d = d.AddDate(0, 0, -1) {
		streak++
	}
	return streak
}

func computeLongestStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	asc := make([]time.Time, len(dates))
	copy(asc, dates)
	sort.Slice(asc, func(i, j int) bool {
		return asc[i].Before(asc[j])
	})
	longest := 1
	run := 1
	for i := 1; i < len(asc); i++ {
		prev := asc[i-1]
		cur := asc[i]
		if dateKey(cur) == dateKey(prev.AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}
	return longest
}

func countDaysThisWeek(dateSet map[string]bool, today time.Time) int {
	count := 0
	for i := 0; i < 7; i++ {
		d := today.AddDate(0, 0, -i)
		if hasLoggedDate(dateSet, d) {
			count++
		}
	}
	return count
}

func (s *foodLogService) GetMemberStats(ctx context.Context, userID uint) (*foodlogv1.MemberStatsRes, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	since := today.AddDate(0, 0, -365)

	savedGoal, _ := s.repo.GetGoals(ctx, userID)
	goals := s.resolveGoals(ctx, userID)
	goalsSaved := savedGoal != nil

	dates, err := s.repo.ListLoggedDates(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	dateSet := make(map[string]bool, len(dates))
	for _, d := range dates {
		dateSet[dateKey(d)] = true
	}

	current := computeCurrentStreak(dateSet, today)
	longest := computeLongestStreak(dates)
	if current > longest {
		longest = current
	}

	entries, err := s.repo.ListByDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	totals := sumEntries(entries)
	pct := 0
	if goals.DailyCalories > 0 {
		pct = int(totals.Calories / goals.DailyCalories * 100)
		if pct > 100 {
			pct = 100
		}
	}

	return &foodlogv1.MemberStatsRes{
		Goals:      goals,
		GoalsSaved: goalsSaved,
		Streak: foodlogv1.FoodLogStreakRes{
			Current:      current,
			Longest:      longest,
			LoggedToday:  hasLoggedDate(dateSet, today),
			DaysThisWeek: countDaysThisWeek(dateSet, today),
		},
		Today: foodlogv1.TodayProgressRes{
			Calories:     totals.Calories,
			GoalCalories: goals.DailyCalories,
			ProteinG:     totals.ProteinG,
			GoalProteinG: goals.DailyProteinG,
			Pct:          pct,
		},
	}, nil
}
