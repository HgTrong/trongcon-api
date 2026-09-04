package v1

import "time"

type MoneySnap struct {
	Gross     float64 `json:"gross"`
	Platform  float64 `json:"platform"`
	PTShare   float64 `json:"pt_share"`
	Orders    int64   `json:"orders"`
	Premium   float64 `json:"premium"`
	GymPass   float64 `json:"gym_pass"`
	PTPackage float64 `json:"pt_package"`
}

// DailyPoint is one UTC calendar day of revenue (for charts).
type DailyPoint struct {
	Date     string  `json:"date"` // YYYY-MM-DD
	Gross    float64 `json:"gross"`
	Platform float64 `json:"platform"`
	PTShare  float64 `json:"pt_share"`
	Orders   int64   `json:"orders"`
}

type ChangePct struct {
	Gross    float64 `json:"gross"`
	Platform float64 `json:"platform"`
	PTShare  float64 `json:"pt_share"`
	Orders   float64 `json:"orders"`
}

type SourceBreakdown struct {
	Source   string  `json:"source"`
	Label    string  `json:"label"`
	Gross    float64 `json:"gross"`
	Platform float64 `json:"platform"`
	PTShare  float64 `json:"pt_share"`
	Orders   int64   `json:"orders"`
}

type PTLeaderboardRow struct {
	TrainerProfileID uint    `json:"trainer_profile_id"`
	DisplayName      string  `json:"display_name"`
	Title            string  `json:"title,omitempty"`
	AvatarURL        string  `json:"avatar_url,omitempty"`
	Gross            float64 `json:"gross"`
	PTShare          float64 `json:"pt_share"`
	PlatformShare    float64 `json:"platform_share"`
	Orders           int64   `json:"orders"`
	Currency         string  `json:"currency"`
}

type OverviewRes struct {
	Currency      string             `json:"currency"`
	Today         MoneySnap          `json:"today"`
	Yesterday     MoneySnap          `json:"yesterday"`
	ChangePct     ChangePct          `json:"change_pct"`
	AllTime       MoneySnap          `json:"all_time"`
	Last7Days     MoneySnap          `json:"last_7_days"`
	Last30Days    MoneySnap          `json:"last_30_days"`
	OwedToPT      float64            `json:"owed_to_pt"`
	PlatformTake  float64            `json:"platform_take"`
	BySource      []SourceBreakdown  `json:"by_source"`
	PTLeaderboard []PTLeaderboardRow `json:"pt_leaderboard"`
	DailySeries   []DailyPoint       `json:"daily_series"` // last 30 UTC days inclusive
	GeneratedAt   time.Time          `json:"generated_at"`

	// Selected is the same MoneySnap shape but scoped to whatever [from, to)
	// range the caller asked for (?from=&to=, defaults to "today" so old
	// clients that ignore these fields still see today's numbers here too).
	// PreviousPeriod is the immediately preceding window of equal length —
	// generalizes the fixed "today vs yesterday" comparison to any range.
	RangeFrom           string             `json:"range_from"`
	RangeTo             string             `json:"range_to"`
	Selected            MoneySnap          `json:"selected"`
	PreviousPeriod      MoneySnap          `json:"previous_period"`
	SelectedChangePct   ChangePct          `json:"selected_change_pct"`
	SelectedBySource    []SourceBreakdown  `json:"selected_by_source"`
	SelectedLeaderboard []PTLeaderboardRow `json:"selected_leaderboard"`
}

type PaymentRow struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Label       string    `json:"label"`
	UserID      uint      `json:"user_id"`
	UserEmail   string    `json:"user_email,omitempty"`
	UserName    string    `json:"user_name,omitempty"`
	TrainerName string    `json:"trainer_name,omitempty"`
	Amount      float64   `json:"amount"`
	Platform    float64   `json:"platform"`
	PTShare     float64   `json:"pt_share"`
	Currency    string    `json:"currency"`
	Provider    string    `json:"provider,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaymentsRes struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	Limit    int          `json:"limit"`
	Currency string       `json:"currency"`
	Data     []PaymentRow `json:"data"`
}

type PTLeaderboardRes struct {
	Total    int64              `json:"total"`
	Currency string             `json:"currency"`
	SortBy   string             `json:"sort_by"`
	Data     []PTLeaderboardRow `json:"data"`
}

// TodayActivityRes powers the admin dashboard's "hoạt động hôm nay" panel —
// who signed up, who bought what, at a glance without opening every list page.
type TodayActivityRes struct {
	Date              string                   `json:"date"`
	NewUsers          int64                    `json:"new_users"`
	Revenue           MoneySnap                `json:"revenue"`
	NewGymMemberships []TodayGymMembershipItem `json:"new_gym_memberships"`
	NewPTPackages     []TodayPTPackageItem     `json:"new_pt_packages"`
	NewPremium        []TodayPremiumItem       `json:"new_premium"`
}

type TodayGymMembershipItem struct {
	ID        uint      `json:"id"`
	UserName  string    `json:"user_name"`
	UserEmail string    `json:"user_email"`
	PlanName  string    `json:"plan_name"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type TodayPTPackageItem struct {
	ID           uint      `json:"id"`
	StudentName  string    `json:"student_name"`
	StudentEmail string    `json:"student_email"`
	TrainerName  string    `json:"trainer_name"`
	PackageTitle string    `json:"package_title"`
	SessionCount int       `json:"session_count"`
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
}

type TodayPremiumItem struct {
	ID        uint      `json:"id"`
	UserName  string    `json:"user_name"`
	UserEmail string    `json:"user_email"`
	PlanName  string    `json:"plan_name"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}
