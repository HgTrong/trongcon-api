package v1

import "time"

// —— Membership plans ——

type MembershipPlanReq struct {
	Code            string  `json:"code" binding:"required"`
	Name            string  `json:"name" binding:"required"`
	Description     string  `json:"description"`
	Price           float64 `json:"price" binding:"required"`
	Currency        string  `json:"currency"`
	DurationMonths  int     `json:"duration_months" binding:"required,min=1"`
	BranchID        *uint   `json:"branch_id"`
	IncludesClasses *bool   `json:"includes_classes"`
	IsHighlighted   *bool   `json:"is_highlighted"`
	IsActive        *bool   `json:"is_active"`
	SortOrder       int     `json:"sort_order"`
}

type HighlightPlanReq struct {
	IsHighlighted bool `json:"is_highlighted"`
}

type MembershipPlanRes struct {
	ID              uint    `json:"id"`
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Price           float64 `json:"price"`
	Currency        string  `json:"currency"`
	DurationMonths  int     `json:"duration_months"`
	BranchID        *uint   `json:"branch_id"`
	BranchName      string  `json:"branch_name,omitempty"`
	IncludesClasses bool    `json:"includes_classes"`
	IsHighlighted   bool    `json:"is_highlighted"`
	IsActive        bool    `json:"is_active"`
	SortOrder       int     `json:"sort_order"`
}

type ListRes struct {
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

type CheckoutReq struct {
	PlanID uint `json:"plan_id" binding:"required"`
}

type PackageCheckoutReq struct {
	PackageID uint `json:"package_id" binding:"required"`
}

type CheckoutRes struct {
	ID          uint   `json:"id"`
	OrderID     string `json:"order_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	CheckoutURL string `json:"checkout_url,omitempty"`
	ApproveURL  string `json:"approve_url,omitempty"`
	Provider    string `json:"provider"`
}

type ConfirmVNPayReq struct {
	Params map[string]string `json:"params" binding:"required"`
}

type ConfirmStripeReq struct {
	SessionID string `json:"session_id" binding:"required"`
}

type GymMembershipRes struct {
	ID                  uint      `json:"id"`
	UserID              uint      `json:"user_id"`
	UserEmail           string    `json:"user_email,omitempty"`
	UserName            string    `json:"user_name,omitempty"`
	GymMembershipPlanID uint      `json:"gym_membership_plan_id"`
	PlanName            string    `json:"plan_name,omitempty"`
	IncludesClasses     bool      `json:"includes_classes"`
	BranchID            *uint     `json:"branch_id"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	DurationMonths      int       `json:"duration_months"`
	Price               float64   `json:"price"`
	Currency            string    `json:"currency"`
	Status              string    `json:"status"`
	PaymentProvider     string    `json:"payment_provider"`
	VnpTxnRef           string    `json:"vnp_txn_ref,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type GroupClassReq struct {
	BranchID    uint   `json:"branch_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Category    string `json:"category"`
	Description string `json:"description"`
	DurationMin int    `json:"duration_min"`
	Capacity    int    `json:"capacity"`
	TrainerID   *uint  `json:"trainer_id"`
	IsActive    *bool  `json:"is_active"`
}

type GroupClassRes struct {
	ID          uint   `json:"id"`
	BranchID    uint   `json:"branch_id"`
	BranchName  string `json:"branch_name,omitempty"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	DurationMin int    `json:"duration_min"`
	Capacity    int    `json:"capacity"`
	TrainerID   *uint  `json:"trainer_id"`
	IsActive    bool   `json:"is_active"`
}

type ClassSessionReq struct {
	GroupClassID uint      `json:"group_class_id" binding:"required"`
	StartsAt     time.Time `json:"starts_at" binding:"required"`
	EndsAt       time.Time `json:"ends_at" binding:"required"`
	Capacity     int       `json:"capacity"`
}

type ClassSessionRes struct {
	ID           uint      `json:"id"`
	GroupClassID uint      `json:"group_class_id"`
	ClassName    string    `json:"class_name,omitempty"`
	Category     string    `json:"category,omitempty"`
	BranchID     uint      `json:"branch_id,omitempty"`
	BranchName   string    `json:"branch_name,omitempty"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	Capacity     int       `json:"capacity"`
	BookedCount  int       `json:"booked_count"`
	IsCanceled   bool      `json:"is_canceled"`
}

type ClassBookingRes struct {
	ID             uint      `json:"id"`
	ClassSessionID uint      `json:"class_session_id"`
	Status         string    `json:"status"`
	ClassName      string    `json:"class_name,omitempty"`
	StartsAt       time.Time `json:"starts_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PTPackageReq struct {
	Title        string  `json:"title" binding:"required"`
	Description  string  `json:"description"`
	SessionCount int     `json:"session_count" binding:"required,min=1"`
	Price        float64 `json:"price" binding:"required"`
	Currency     string  `json:"currency"`
	ValidDays    int     `json:"valid_days"`
	IsPublic     *bool   `json:"is_public"`
	IsActive     *bool   `json:"is_active"`
}

type PTPackageRes struct {
	ID               uint    `json:"id"`
	TrainerProfileID uint    `json:"trainer_profile_id"`
	TrainerName      string  `json:"trainer_name,omitempty"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	SessionCount     int     `json:"session_count"`
	Price            float64 `json:"price"`
	Currency         string  `json:"currency"`
	ValidDays        int     `json:"valid_days"`
	IsPublic         bool    `json:"is_public"`
	IsActive         bool    `json:"is_active"`
}

type UserPTPackageRes struct {
	ID               uint       `json:"id"`
	UserID           uint       `json:"user_id"`
	StudentName      string     `json:"student_name,omitempty"`
	StudentEmail     string     `json:"student_email,omitempty"`
	PTPackageID      uint       `json:"pt_package_id"`
	PackageTitle     string     `json:"package_title,omitempty"`
	TrainerProfileID uint       `json:"trainer_profile_id"`
	TrainerName      string     `json:"trainer_name,omitempty"`
	SessionTotal     int        `json:"session_total"`
	SessionUsed      int        `json:"session_used"`
	SessionLeft      int        `json:"session_left"`
	Price            float64    `json:"price"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	StartsAt         time.Time  `json:"starts_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	PaymentProvider  string     `json:"payment_provider,omitempty"`
	VnpTxnRef        string     `json:"vnp_txn_ref,omitempty"`
	VnpTransactionNo string     `json:"vnp_transaction_no,omitempty"`
	UnreadCount      int64      `json:"unread_count,omitempty"`
	LastMessage      string     `json:"last_message,omitempty"`
	LastMessageAt    *time.Time `json:"last_message_at,omitempty"`
	PendingOffers    int64      `json:"pending_offers,omitempty"`
}

type LogPTSessionReq struct {
	ProofImageURL string     `json:"proof_image_url" binding:"required"`
	Note          string     `json:"note"`
	TaughtAt      *time.Time `json:"taught_at"`
}

type PTSessionLogRes struct {
	ID               uint      `json:"id"`
	UserPTPackageID  uint      `json:"user_pt_package_id"`
	TrainerProfileID uint      `json:"trainer_profile_id"`
	UserID           uint      `json:"user_id"`
	SessionIndex     int       `json:"session_index"`
	TaughtAt         time.Time `json:"taught_at"`
	Note             string    `json:"note,omitempty"`
	ProofImageURL    string    `json:"proof_image_url"`
	CreatedByUserID  uint      `json:"created_by_user_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type RevenueShareReq struct {
	PTPercent  float64 `json:"pt_percent" binding:"required"`
	GymPercent float64 `json:"gym_percent" binding:"required"`
	Notes      string  `json:"notes"`
}

type RevenueShareRes struct {
	ID         uint    `json:"id"`
	PTPercent  float64 `json:"pt_percent"`
	GymPercent float64 `json:"gym_percent"`
	Notes      string  `json:"notes"`
}

type PTEarningRes struct {
	ID               uint      `json:"id"`
	TrainerProfileID uint      `json:"trainer_profile_id"`
	TrainerName      string    `json:"trainer_name,omitempty"`
	UserPTPackageID  uint      `json:"user_pt_package_id"`
	SessionOfferID   *uint     `json:"session_offer_id,omitempty"`
	PackageTitle     string    `json:"package_title,omitempty"`
	StudentName      string    `json:"student_name,omitempty"`
	StudentEmail     string    `json:"student_email,omitempty"`
	GrossAmount      float64   `json:"gross_amount"`
	PTPercent        float64   `json:"pt_percent"`
	PTAmount         float64   `json:"pt_amount"`
	GymAmount        float64   `json:"gym_amount"`
	Currency         string    `json:"currency"`
	Note             string    `json:"note,omitempty"`
	PaidOut          bool      `json:"paid_out"`
	CreatedAt        time.Time `json:"created_at"`
}

type EarningsSummaryRes struct {
	TotalPTAmount float64        `json:"total_pt_amount"`
	Currency      string         `json:"currency"`
	Data          []PTEarningRes `json:"data"`
	Total         int64          `json:"total"`
}

// TodayPackageItemRes is one PT-package purchase, for the "today" and
// "unseen" lists on the PT dashboard.
type TodayPackageItemRes struct {
	ID           uint      `json:"id"`
	StudentName  string    `json:"student_name"`
	StudentEmail string    `json:"student_email"`
	PackageTitle string    `json:"package_title"`
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
}

// MyTodayActivityRes powers the PT-side "hoạt động hôm nay" panel — today's
// package sales + commission, and any new-student purchases the trainer
// hasn't acknowledged yet (see MarkStudentsSeen).
type MyTodayActivityRes struct {
	Date             string                `json:"date"`
	TodayRevenue     float64               `json:"today_revenue"`
	TodayPTEarnings  float64               `json:"today_pt_earnings"`
	TodayNewStudents int                   `json:"today_new_students"`
	TodayPackages    []TodayPackageItemRes `json:"today_packages"`
	Currency         string                `json:"currency"`
	UnseenCount      int                   `json:"unseen_count"`
	UnseenPackages   []TodayPackageItemRes `json:"unseen_packages"`
}

// MembershipMeRes wraps a user's current + recent gym memberships.
type MembershipMeRes struct {
	Active *GymMembershipRes  `json:"active"`
	Recent []GymMembershipRes `json:"recent"`
}

type CheckInTokenRes struct {
	Token        string    `json:"token"`
	ExpiresInSec int       `json:"expires_in_sec"`
	MembershipID uint      `json:"membership_id"`
	PlanName     string    `json:"plan_name,omitempty"`
	EndDate      time.Time `json:"end_date"`
}

type AdminCreateGymMembershipReq struct {
	UserID uint `json:"user_id" binding:"required"`
	PlanID uint `json:"plan_id" binding:"required"`
}

type AdminCreateUserPTPackageReq struct {
	UserID    uint `json:"user_id" binding:"required"`
	PackageID uint `json:"pt_package_id" binding:"required"`
}

type VerifyCheckInReq struct {
	Token    string `json:"token" binding:"required"`
	BranchID *uint  `json:"branch_id"`
	Note     string `json:"note"`
}

type GymCheckInRes struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	UserName     string    `json:"user_name,omitempty"`
	UserEmail    string    `json:"user_email,omitempty"`
	MembershipID uint      `json:"membership_id"`
	PlanName     string    `json:"plan_name,omitempty"`
	BranchID     *uint     `json:"branch_id"`
	CheckedInAt  time.Time `json:"checked_in_at"`
	Note         string    `json:"note,omitempty"`
}

type GymCheckInStatsRes struct {
	ActiveMembers      int64 `json:"active_members"`
	CheckedInToday     int64 `json:"checked_in_today"`
	UniqueMembersToday int64 `json:"unique_members_today"`
}

type GymCheckInListRes struct {
	Total int64              `json:"total"`
	Data  []GymCheckInRes    `json:"data"`
	Stats GymCheckInStatsRes `json:"stats"`
}

type SendChatMessageReq struct {
	Body string `json:"body" binding:"required"`
}

type CreateSessionOfferReq struct {
	StartsAt time.Time `json:"starts_at" binding:"required"`
	Note     string    `json:"note"`
}

type CompleteSessionOfferReq struct {
	ProofImageURL string `json:"proof_image_url" binding:"required"`
	Note          string `json:"note"`
}

type RescheduleSessionOfferReq struct {
	StartsAt time.Time `json:"starts_at" binding:"required"`
}

// LogSessionDirectReq lets a trainer record a session that was scheduled entirely
// outside the app (e.g. agreed over Zalo) — skips propose/accept and goes straight
// to "awaiting confirmation", same as CompleteSessionOfferReq but self-contained.
type LogSessionDirectReq struct {
	TaughtAt      time.Time `json:"taught_at"`
	ProofImageURL string    `json:"proof_image_url" binding:"required"`
	Note          string    `json:"note"`
}

type SessionOfferRes struct {
	ID                uint       `json:"id"`
	UserPTPackageID   uint       `json:"user_pt_package_id"`
	TrainerProfileID  uint       `json:"trainer_profile_id"`
	StudentUserID     uint       `json:"student_user_id"`
	StudentName       string     `json:"student_name,omitempty"`
	StudentEmail      string     `json:"student_email,omitempty"`
	StartsAt          time.Time  `json:"starts_at"`
	Note              string     `json:"note,omitempty"`
	ProposedByUserID  uint       `json:"proposed_by_user_id"`
	Status            string     `json:"status"`
	AcceptedByUserID  uint       `json:"accepted_by_user_id,omitempty"`
	AcceptedAt        *time.Time `json:"accepted_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	ProofImageURL     string     `json:"proof_image_url,omitempty"`
	SessionIndex      int        `json:"session_index,omitempty"`
	CompletedByUserID uint       `json:"completed_by_user_id,omitempty"`
	ConfirmedByUserID uint       `json:"confirmed_by_user_id,omitempty"`
	ConfirmedAt       *time.Time `json:"confirmed_at,omitempty"`
	ProofSubmittedAt  *time.Time `json:"proof_submitted_at,omitempty"`
	EndsAt            *time.Time `json:"ends_at,omitempty"`
	BookedViaSlot     bool       `json:"booked_via_slot"`
	CreatedAt         time.Time  `json:"created_at"`
}

// AdminSessionReviewRes is one row in the admin "pending session review" queue —
// staff use StudentCheckedInThatDay to sanity-check the proof against the gym's
// own QR check-in log before approving.
type AdminSessionReviewRes struct {
	SessionOfferRes
	TrainerName             string `json:"trainer_name,omitempty"`
	PackageTitle            string `json:"package_title,omitempty"`
	StudentCheckedInThatDay bool   `json:"student_checked_in_that_day"`
}

type WorkingHoursItem struct {
	Weekday     int  `json:"weekday" binding:"min=0,max=6"`
	StartMinute int  `json:"start_minute" binding:"min=0,max=1439"`
	EndMinute   int  `json:"end_minute" binding:"min=1,max=1440"`
	IsActive    bool `json:"is_active"`
}

type SetWorkingHoursReq struct {
	Hours []WorkingHoursItem `json:"hours" binding:"required"`
}

type BookingSettingsReq struct {
	SessionDurationMin  *int  `json:"session_duration_min"`
	AcceptingNewClients *bool `json:"accepting_new_clients"`
	BookingPaused       *bool `json:"booking_paused"`
	MaxActiveClients    *int  `json:"max_active_clients"`
}

type BookingSettingsRes struct {
	TrainerProfileID    uint `json:"trainer_profile_id"`
	SessionDurationMin  int  `json:"session_duration_min"`
	AcceptingNewClients bool `json:"accepting_new_clients"`
	BookingPaused       bool `json:"booking_paused"`
	MaxActiveClients    int  `json:"max_active_clients"`
	ActiveClients       int  `json:"active_clients"`
}

type AvailableSlotRes struct {
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

type BookSlotReq struct {
	UserPTPackageID   uint      `json:"user_pt_package_id" binding:"required"`
	StartsAt          time.Time `json:"starts_at" binding:"required"`
	Note              string    `json:"note"`
	SourceContentType string    `json:"source_content_type"`
	SourceContentID   uint      `json:"source_content_id"`
	SourceTitle       string    `json:"source_title"`
}

type CreateRecurringBookingReq struct {
	UserPTPackageID uint `json:"user_pt_package_id" binding:"required"`
	Weekday         int  `json:"weekday" binding:"min=0,max=6"`
	StartMinute     int  `json:"start_minute" binding:"min=0,max=1439"`
}

type RecurringBookingRes struct {
	ID                uint       `json:"id"`
	UserPTPackageID   uint       `json:"user_pt_package_id"`
	TrainerProfileID  uint       `json:"trainer_profile_id"`
	StudentUserID     uint       `json:"student_user_id"`
	StudentName       string     `json:"student_name,omitempty"`
	StudentEmail      string     `json:"student_email,omitempty"`
	PackageTitle      string     `json:"package_title,omitempty"`
	Weekday           int        `json:"weekday"`
	StartMinute       int        `json:"start_minute"`
	EndMinute         int        `json:"end_minute"`
	Status            string     `json:"status"`
	OccurrencesQueued int        `json:"occurrences_queued,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	LastGeneratedFor  *time.Time `json:"last_generated_for,omitempty"`
}

type BlockSlotReq struct {
	StartsAt time.Time `json:"starts_at" binding:"required"`
	EndsAt   time.Time `json:"ends_at" binding:"required"`
	Reason   string    `json:"reason"`
}

type BlockedSlotRes struct {
	ID               uint      `json:"id"`
	TrainerProfileID uint      `json:"trainer_profile_id"`
	StartsAt         time.Time `json:"starts_at"`
	EndsAt           time.Time `json:"ends_at"`
	Reason           string    `json:"reason,omitempty"`
}

type TrainerOpsOverviewRes struct {
	TrainerProfileID    uint   `json:"trainer_profile_id"`
	DisplayName         string `json:"display_name"`
	Status              string `json:"status"` // active | paused | private
	AcceptingNewClients bool   `json:"accepting_new_clients"`
	BookingPaused       bool   `json:"booking_paused"`
	ActiveClients       int    `json:"active_clients"`
	MaxActiveClients    int    `json:"max_active_clients"`
	AvailableSlotsWeek  int    `json:"available_slots_this_week"`
	TotalBookings       int64  `json:"total_bookings"`
	Completed           int64  `json:"completed"`
	Upcoming            int64  `json:"upcoming"`
	Cancelled           int64  `json:"cancelled"`
	AwaitingConfirm     int64  `json:"awaiting_confirmation"`
}

type TrainerClientRes struct {
	UserID          uint      `json:"user_id"`
	UserName        string    `json:"user_name"`
	UserEmail       string    `json:"user_email"`
	UserPTPackageID uint      `json:"user_pt_package_id"`
	PackageTitle    string    `json:"package_title"`
	JoinedAt        time.Time `json:"joined_at"`
	SessionTotal    int       `json:"session_total"`
	SessionUsed     int       `json:"session_used"`
	Status          string    `json:"status"`
}

type ContentFunnelItemRes struct {
	ContentType       string `json:"content_type"`
	ContentID         uint   `json:"content_id"`
	Title             string `json:"title"`
	Views             int64  `json:"views"`
	Likes             int64  `json:"likes"`
	Saves             int64  `json:"saves"`
	ProfileVisits     int64  `json:"profile_visits"`
	BookingsGenerated int64  `json:"bookings_generated"`
}

type ContentFunnelRes struct {
	TrainerProfileID    uint                   `json:"trainer_profile_id"`
	DisplayName         string                 `json:"display_name"`
	Articles            int                    `json:"articles"`
	Workouts            int                    `json:"workouts"`
	Routines            int                    `json:"routines"`
	MealPlans           int                    `json:"meal_plans"`
	TotalViews          int64                  `json:"total_views"`
	TotalLikes          int64                  `json:"total_likes"`
	TotalSaves          int64                  `json:"total_saves"`
	ProfileVisits       int64                  `json:"profile_visits"`
	BookingsFromContent int64                  `json:"bookings_from_content"`
	Items               []ContentFunnelItemRes `json:"items"`
}

type TrainerQualityRes struct {
	TrainerProfileID  uint    `json:"trainer_profile_id"`
	DisplayName       string  `json:"display_name"`
	Rating            float64 `json:"rating"`
	Reviews           int64   `json:"reviews"`
	CompletedSessions int64   `json:"completed_sessions"`
	CancellationRate  float64 `json:"cancellation_rate"`
	NoShowCount       int64   `json:"no_show_count"`
	NoShowRate        float64 `json:"no_show_rate"`
	ClientRetention   float64 `json:"client_retention"`
	ActiveClients     int     `json:"active_clients"`
	TotalClientsEver  int     `json:"total_clients_ever"`
}

type CreatePTReviewReq struct {
	SessionOfferID  uint   `json:"session_offer_id" binding:"required"`
	UserPTPackageID uint   `json:"user_pt_package_id" binding:"required"`
	Rating          int    `json:"rating" binding:"required,min=1,max=5"`
	Comment         string `json:"comment"`
}

type TouchPTFunnelReq struct {
	TrainerProfileID uint   `json:"trainer_profile_id" binding:"required"`
	Event            string `json:"event" binding:"required"` // content_view | profile_visit | booking | like
	ContentType      string `json:"content_type"`
	ContentID        uint   `json:"content_id"`
	Title            string `json:"title"`
}

type PTReviewRes struct {
	ID               uint      `json:"id"`
	TrainerProfileID uint      `json:"trainer_profile_id"`
	StudentUserID    uint      `json:"student_user_id"`
	StudentName      string    `json:"student_name,omitempty"`
	StudentEmail     string    `json:"student_email,omitempty"`
	UserPTPackageID  uint      `json:"user_pt_package_id"`
	SessionOfferID   uint      `json:"session_offer_id"`
	Rating           int       `json:"rating"`
	Comment          string    `json:"comment,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type CalendarDaySlotRes struct {
	Hour   int              `json:"hour"`
	Minute int              `json:"minute"`
	Status string           `json:"status"` // available | booked | blocked | empty
	Offer  *SessionOfferRes `json:"offer,omitempty"`
	Label  string           `json:"label,omitempty"`
}

type CalendarDayRes struct {
	Date  string               `json:"date"` // YYYY-MM-DD (VN)
	Slots []CalendarDaySlotRes `json:"slots"`
}

type TrainerCalendarRes struct {
	TrainerProfileID uint             `json:"trainer_profile_id"`
	From             string           `json:"from"`
	To               string           `json:"to"`
	Days             []CalendarDayRes `json:"days"`
}

type ChatMessageRes struct {
	ID                uint             `json:"id"`
	UserPTPackageID   uint             `json:"user_pt_package_id"`
	SenderUserID      uint             `json:"sender_user_id"`
	SenderName        string           `json:"sender_name,omitempty"`
	SenderAvatarURL   string           `json:"sender_avatar_url,omitempty"`
	Body              string           `json:"body"`
	MessageType       string           `json:"message_type"`
	SessionOfferID    *uint            `json:"session_offer_id,omitempty"`
	SessionOffer      *SessionOfferRes `json:"session_offer,omitempty"`
	SharedContentType string           `json:"shared_content_type,omitempty"`
	SharedContentID   *uint            `json:"shared_content_id,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
}

type ChatMessagesRes struct {
	Total     int64            `json:"total"`
	Data      []ChatMessageRes `json:"data"`
	CanSend   bool             `json:"can_send"`
	PackageID uint             `json:"user_pt_package_id"`
}
