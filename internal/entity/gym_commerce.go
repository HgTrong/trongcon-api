package entity

import "time"

const (
	PHTypeGymMembership = "gym_membership"
	PHTypePTPackage     = "pt_package"

	GymMemStatusPending  = "pending"
	GymMemStatusActive   = "active"
	GymMemStatusExpired  = "expired"
	GymMemStatusCanceled = "canceled"

	PTPkgStatusPending  = "pending"
	PTPkgStatusActive   = "active"
	PTPkgStatusExpired  = "expired"
	PTPkgStatusCanceled = "canceled"

	ClassBookingBooked   = "booked"
	ClassBookingCanceled = "canceled"

	ChatMsgTypeText         = "text"
	ChatMsgTypeSessionOffer = "session_offer"
	ChatMsgTypeContentShare = "content_share"

	SessionOfferPending              = "pending"
	SessionOfferScheduled            = "scheduled"
	SessionOfferAwaitingConfirmation = "awaiting_confirmation"
	SessionOfferDeclined             = "declined"
	SessionOfferCancelled            = "cancelled"
	SessionOfferCompleted            = "completed"
	SessionOfferNoShow               = "no_show"

	RecurringBookingStatusActive   = "active"
	RecurringBookingStatusPaused   = "paused"
	RecurringBookingStatusCanceled = "canceled"
)

// GymMembershipPlan is a club pass (floor + group classes).
type GymMembershipPlan struct {
	BaseEntity
	Code            string     `json:"code" gorm:"type:varchar(64);uniqueIndex"`
	Name            string     `json:"name" gorm:"type:varchar(255);not null"`
	Description     string     `json:"description" gorm:"type:text"`
	Price           float64    `json:"price" gorm:"type:decimal(12,2);not null;default:0"`
	Currency        string     `json:"currency" gorm:"type:varchar(10);not null;default:'VND'"`
	DurationMonths  int        `json:"duration_months" gorm:"not null;default:1"`
	BranchID        *uint      `json:"branch_id" gorm:"index"` // nil = all branches
	Branch          *GymBranch `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	IncludesClasses bool       `json:"includes_classes" gorm:"not null;default:true"`
	IsHighlighted   bool       `json:"is_highlighted" gorm:"not null;default:false;index"` // show on marketing home
	IsActive        bool       `json:"is_active" gorm:"not null;default:true;index"`
	SortOrder       int        `json:"sort_order" gorm:"not null;default:0"`
}

func (GymMembershipPlan) TableName() string { return "gym_membership_plans" }

// UserGymMembership is a purchased club pass window.
type UserGymMembership struct {
	BaseEntity
	UserID                  uint              `json:"user_id" gorm:"not null;index"`
	User                    User              `json:"-" gorm:"foreignKey:UserID"`
	GymMembershipPlanID     uint              `json:"gym_membership_plan_id" gorm:"not null;index"`
	GymMembershipPlan       GymMembershipPlan `json:"plan,omitempty" gorm:"foreignKey:GymMembershipPlanID"`
	BranchID                *uint             `json:"branch_id" gorm:"index"`
	StartDate               time.Time         `json:"start_date"`
	EndDate                 time.Time         `json:"end_date"`
	DurationMonths          int               `json:"duration_months"`
	Price                   float64           `json:"price" gorm:"type:decimal(12,2);not null;default:0"`
	Currency                string            `json:"currency" gorm:"type:varchar(10);default:'VND'"`
	Status                  string            `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	PaymentProvider         string            `json:"payment_provider" gorm:"type:varchar(20);not null;default:'vnpay'"`
	VnpTxnRef               string            `json:"vnp_txn_ref" gorm:"type:varchar(100);index"`
	VnpTransactionNo        string            `json:"vnp_transaction_no" gorm:"type:varchar(100);index"`
	StripeCheckoutSessionID string            `json:"stripe_checkout_session_id" gorm:"type:varchar(255);index"`
	StripePaymentIntentID   string            `json:"stripe_payment_intent_id" gorm:"type:varchar(255);index"`
}

func (UserGymMembership) TableName() string { return "user_gym_memberships" }

// GroupClass is a recurring class type (yoga, zumba, …) at a branch.
type GroupClass struct {
	BaseEntity
	BranchID    uint            `json:"branch_id" gorm:"not null;index"`
	Branch      GymBranch       `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	Name        string          `json:"name" gorm:"type:varchar(200);not null"`
	Category    string          `json:"category" gorm:"type:varchar(64);index"` // yoga, zumba, hiit, …
	Description string          `json:"description" gorm:"type:text"`
	DurationMin int             `json:"duration_min" gorm:"not null;default:60"`
	Capacity    int             `json:"capacity" gorm:"not null;default:20"`
	TrainerID   *uint           `json:"trainer_id" gorm:"index"` // optional coach for the class
	Trainer     *TrainerProfile `json:"trainer,omitempty" gorm:"foreignKey:TrainerID"`
	IsActive    bool            `json:"is_active" gorm:"not null;default:true;index"`
}

func (GroupClass) TableName() string { return "group_classes" }

// ClassSession is one scheduled occurrence of a group class.
type ClassSession struct {
	BaseEntity
	GroupClassID uint       `json:"group_class_id" gorm:"not null;index"`
	GroupClass   GroupClass `json:"group_class,omitempty" gorm:"foreignKey:GroupClassID"`
	StartsAt     time.Time  `json:"starts_at" gorm:"not null;index"`
	EndsAt       time.Time  `json:"ends_at" gorm:"not null"`
	Capacity     int        `json:"capacity" gorm:"not null;default:20"`
	BookedCount  int        `json:"booked_count" gorm:"not null;default:0"`
	IsCanceled   bool       `json:"is_canceled" gorm:"not null;default:false"`
}

func (ClassSession) TableName() string { return "class_sessions" }

// ClassBooking is a member reservation for a class session.
type ClassBooking struct {
	BaseEntity
	UserID         uint         `json:"user_id" gorm:"not null;index"`
	User           User         `json:"-" gorm:"foreignKey:UserID"`
	ClassSessionID uint         `json:"class_session_id" gorm:"not null;index"`
	ClassSession   ClassSession `json:"session,omitempty" gorm:"foreignKey:ClassSessionID"`
	Status         string       `json:"status" gorm:"type:varchar(20);not null;default:'booked';index"`
}

func (ClassBooking) TableName() string { return "class_bookings" }

// PTPackage is a session bundle priced by a trainer.
type PTPackage struct {
	BaseEntity
	TrainerProfileID uint           `json:"trainer_profile_id" gorm:"not null;index"`
	Trainer          TrainerProfile `json:"trainer,omitempty" gorm:"foreignKey:TrainerProfileID"`
	Title            string         `json:"title" gorm:"type:varchar(255);not null"`
	Description      string         `json:"description" gorm:"type:text"`
	SessionCount     int            `json:"session_count" gorm:"not null;default:4"`
	Price            float64        `json:"price" gorm:"type:decimal(12,2);not null;default:0"`
	Currency         string         `json:"currency" gorm:"type:varchar(10);not null;default:'VND'"`
	ValidDays        int            `json:"valid_days" gorm:"not null;default:90"`
	IsPublic         bool           `json:"is_public" gorm:"not null;default:false;index"`
	IsActive         bool           `json:"is_active" gorm:"not null;default:true;index"`
}

func (PTPackage) TableName() string { return "pt_packages" }

// UserPTPackage is a purchased PT session bundle.
type UserPTPackage struct {
	BaseEntity
	UserID                  uint      `json:"user_id" gorm:"not null;index"`
	User                    User      `json:"-" gorm:"foreignKey:UserID"`
	PTPackageID             uint      `json:"pt_package_id" gorm:"not null;index"`
	PTPackage               PTPackage `json:"package,omitempty" gorm:"foreignKey:PTPackageID"`
	TrainerProfileID        uint      `json:"trainer_profile_id" gorm:"not null;index"`
	SessionTotal            int       `json:"session_total" gorm:"not null"`
	SessionUsed             int       `json:"session_used" gorm:"not null;default:0"`
	Price                   float64   `json:"price" gorm:"type:decimal(12,2);not null;default:0"`
	Currency                string    `json:"currency" gorm:"type:varchar(10);default:'VND'"`
	Status                  string    `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	StartsAt                time.Time `json:"starts_at"`
	ExpiresAt               time.Time `json:"expires_at"`
	PaymentProvider         string    `json:"payment_provider" gorm:"type:varchar(20);not null;default:'vnpay'"`
	VnpTxnRef               string    `json:"vnp_txn_ref" gorm:"type:varchar(100);index"`
	VnpTransactionNo        string    `json:"vnp_transaction_no" gorm:"type:varchar(100);index"`
	StripeCheckoutSessionID string    `json:"stripe_checkout_session_id" gorm:"type:varchar(255);index"`
	StripePaymentIntentID   string    `json:"stripe_payment_intent_id" gorm:"type:varchar(255);index"`
}

func (UserPTPackage) TableName() string { return "user_pt_packages" }

// RevenueShareSetting is a singleton row (id=1) for PT / gym cut.
type RevenueShareSetting struct {
	BaseEntity
	PTPercent  float64 `json:"pt_percent" gorm:"type:decimal(5,2);not null;default:40"`
	GymPercent float64 `json:"gym_percent" gorm:"type:decimal(5,2);not null;default:60"`
	Notes      string  `json:"notes" gorm:"type:text"`
}

func (RevenueShareSetting) TableName() string { return "revenue_share_settings" }

// PTEarning is a ledger row recorded once per taught (or no-show) session,
// not at purchase time — a PT is only paid for sessions actually delivered.
type PTEarning struct {
	BaseEntity
	TrainerProfileID uint           `json:"trainer_profile_id" gorm:"not null;index"`
	Trainer          TrainerProfile `json:"-" gorm:"foreignKey:TrainerProfileID"`
	UserPTPackageID  uint           `json:"user_pt_package_id" gorm:"not null;index"`
	UserPTPackage    UserPTPackage  `json:"-" gorm:"foreignKey:UserPTPackageID"`
	SessionOfferID   *uint          `json:"session_offer_id" gorm:"index"`
	SessionIndex     int            `json:"session_index" gorm:"not null;default:0"`
	GrossAmount      float64        `json:"gross_amount" gorm:"type:decimal(12,2);not null;default:0"`
	PTPercent        float64        `json:"pt_percent" gorm:"type:decimal(5,2);not null;default:0"`
	PTAmount         float64        `json:"pt_amount" gorm:"type:decimal(12,2);not null;default:0"`
	GymAmount        float64        `json:"gym_amount" gorm:"type:decimal(12,2);not null;default:0"`
	Currency         string         `json:"currency" gorm:"type:varchar(10);default:'VND'"`
	Note             string         `json:"note" gorm:"type:varchar(255)"`
	PaidOut          bool           `json:"paid_out" gorm:"not null;default:false"`
}

func (PTEarning) TableName() string { return "pt_earnings" }

// PTSessionLog is one taught session against a purchased PT package, with photo proof.
type PTSessionLog struct {
	BaseEntity
	UserPTPackageID  uint          `json:"user_pt_package_id" gorm:"not null;index"`
	UserPTPackage    UserPTPackage `json:"-" gorm:"foreignKey:UserPTPackageID"`
	TrainerProfileID uint          `json:"trainer_profile_id" gorm:"not null;index"`
	UserID           uint          `json:"user_id" gorm:"not null;index"` // student
	SessionIndex     int           `json:"session_index" gorm:"not null"` // 1..SessionTotal
	TaughtAt         time.Time     `json:"taught_at" gorm:"not null;index"`
	Note             string        `json:"note" gorm:"type:text"`
	ProofImageURL    string        `json:"proof_image_url" gorm:"type:text;not null"`
	CreatedByUserID  uint          `json:"created_by_user_id" gorm:"not null"`
}

func (PTSessionLog) TableName() string { return "pt_session_logs" }

// PTSessionOffer is a proposed 1-1 session time that both parties must agree on.
type PTSessionOffer struct {
	BaseEntity
	UserPTPackageID   uint          `json:"user_pt_package_id" gorm:"not null;index"`
	UserPTPackage     UserPTPackage `json:"-" gorm:"foreignKey:UserPTPackageID"`
	TrainerProfileID  uint          `json:"trainer_profile_id" gorm:"not null;index"`
	StudentUserID     uint          `json:"student_user_id" gorm:"not null;index"`
	StartsAt          time.Time     `json:"starts_at" gorm:"not null;index"`
	Note              string        `json:"note" gorm:"type:text"`
	ProposedByUserID  uint          `json:"proposed_by_user_id" gorm:"not null;index"`
	Status            string        `json:"status" gorm:"type:varchar(32);not null;default:'pending';index"`
	AcceptedByUserID  uint          `json:"accepted_by_user_id" gorm:"index"`
	AcceptedAt        *time.Time    `json:"accepted_at"`
	CompletedAt       *time.Time    `json:"completed_at"`
	ProofImageURL     string        `json:"proof_image_url" gorm:"type:text"`
	SessionIndex      int           `json:"session_index"`
	CompletedByUserID uint          `json:"completed_by_user_id"`
	ConfirmedByUserID uint          `json:"confirmed_by_user_id" gorm:"index"`
	ConfirmedAt       *time.Time    `json:"confirmed_at"`
	// ProofSubmittedAt is when PT uploaded proof (awaiting student confirm / auto-confirm cron).
	ProofSubmittedAt *time.Time `json:"proof_submitted_at"`
	// EndsAt is session end; used for slot locking / overlap checks.
	EndsAt *time.Time `json:"ends_at" gorm:"index"`
	// BookedViaSlot marks calendar bookings (skip pending negotiate).
	BookedViaSlot bool `json:"booked_via_slot" gorm:"not null;default:false;index"`
	// RecurringBookingID links an auto-materialized occurrence back to its
	// standing weekly reservation (nil for one-off/negotiated bookings).
	RecurringBookingID *uint `json:"recurring_booking_id,omitempty" gorm:"index"`
}

func (PTSessionOffer) TableName() string { return "pt_session_offers" }

// PTRecurringBooking is a student's standing weekly reservation with a
// trainer (e.g. "every Tuesday 14:00") — auto-materialized into dated
// PTSessionOffer rows on a rolling horizon (see MaterializeRecurringBookings)
// until paused/canceled or the package runs out of session credits.
type PTRecurringBooking struct {
	BaseEntity
	UserPTPackageID  uint          `json:"user_pt_package_id" gorm:"not null;index"`
	UserPTPackage    UserPTPackage `json:"-" gorm:"foreignKey:UserPTPackageID"`
	TrainerProfileID uint          `json:"trainer_profile_id" gorm:"not null;index"`
	StudentUserID    uint          `json:"student_user_id" gorm:"not null;index"`
	Weekday          int           `json:"weekday" gorm:"not null"` // 0=Sun..6=Sat
	StartMinute      int           `json:"start_minute" gorm:"not null"`
	EndMinute        int           `json:"end_minute" gorm:"not null"`
	Status           string        `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	// LastGeneratedFor is the VN-local date of the furthest-out occurrence
	// already materialized, so the rolling generator doesn't re-scan from scratch.
	LastGeneratedFor *time.Time `json:"last_generated_for,omitempty"`
}

func (PTRecurringBooking) TableName() string { return "pt_recurring_bookings" }

// PTContentStat is denormalized funnel metrics for one PT-authored content item.
type PTContentStat struct {
	BaseEntity
	TrainerProfileID  uint   `json:"trainer_profile_id" gorm:"not null;uniqueIndex:idx_pt_content_stat;index"`
	ContentType       string `json:"content_type" gorm:"type:varchar(32);not null;uniqueIndex:idx_pt_content_stat"` // article|workout|routine|meal_plan
	ContentID         uint   `json:"content_id" gorm:"not null;uniqueIndex:idx_pt_content_stat"`
	Title             string `json:"title" gorm:"type:varchar(500)"`
	Views             int64  `json:"views" gorm:"not null;default:0"`
	Likes             int64  `json:"likes" gorm:"not null;default:0"`
	Saves             int64  `json:"saves" gorm:"not null;default:0"`
	ProfileVisits     int64  `json:"profile_visits" gorm:"not null;default:0"`
	BookingsGenerated int64  `json:"bookings_generated" gorm:"not null;default:0"`
}

func (PTContentStat) TableName() string { return "pt_content_stats" }

// PTAttribution remembers the last PT content a user engaged with (for funnel → booking).
type PTAttribution struct {
	BaseEntity
	UserID           uint   `json:"user_id" gorm:"not null;uniqueIndex:idx_pt_attr_user_trainer"`
	TrainerProfileID uint   `json:"trainer_profile_id" gorm:"not null;uniqueIndex:idx_pt_attr_user_trainer;index"`
	ContentType      string `json:"content_type" gorm:"type:varchar(32);not null"`
	ContentID        uint   `json:"content_id" gorm:"not null"`
	Title            string `json:"title" gorm:"type:varchar(500)"`
}

func (PTAttribution) TableName() string { return "pt_attributions" }

// PTReview is a 1–5 star rating after a completed session.
type PTReview struct {
	BaseEntity
	TrainerProfileID uint   `json:"trainer_profile_id" gorm:"not null;index"`
	StudentUserID    uint   `json:"student_user_id" gorm:"not null;index"`
	UserPTPackageID  uint   `json:"user_pt_package_id" gorm:"not null;index"`
	SessionOfferID   uint   `json:"session_offer_id" gorm:"not null;uniqueIndex"`
	Rating           int    `json:"rating" gorm:"not null"` // 1..5
	Comment          string `json:"comment" gorm:"type:text"`
}

func (PTReview) TableName() string { return "pt_reviews" }

// PTWorkingHours is one recurring weekday window when a PT accepts bookings.
// A trainer may have multiple windows per weekday (e.g. 08:00–12:00 and 14:00–20:00).
// Weekday: 0=Sunday … 6=Saturday (Go time.Weekday). Start/End are minutes from midnight local VN.
type PTWorkingHours struct {
	BaseEntity
	TrainerProfileID uint `json:"trainer_profile_id" gorm:"not null;index:idx_pt_hours_trainer_weekday;index"`
	Weekday          int  `json:"weekday" gorm:"not null;index:idx_pt_hours_trainer_weekday"`
	StartMinute      int  `json:"start_minute" gorm:"not null;default:480"` // 08:00
	EndMinute        int  `json:"end_minute" gorm:"not null;default:1200"`  // 20:00
	IsActive         bool `json:"is_active" gorm:"not null;default:true"`
}

func (PTWorkingHours) TableName() string { return "pt_working_hours" }

// PTBlockedSlot is a one-off closed window (PT closes a slot / day off).
type PTBlockedSlot struct {
	BaseEntity
	TrainerProfileID uint      `json:"trainer_profile_id" gorm:"not null;index"`
	StartsAt         time.Time `json:"starts_at" gorm:"not null;index"`
	EndsAt           time.Time `json:"ends_at" gorm:"not null;index"`
	Reason           string    `json:"reason" gorm:"type:varchar(255)"`
}

func (PTBlockedSlot) TableName() string { return "pt_blocked_slots" }

// PTPackageChatMessage is a text message between PT and student on a purchased package.
type PTPackageChatMessage struct {
	BaseEntity
	UserPTPackageID uint   `json:"user_pt_package_id" gorm:"not null;index"`
	SenderUserID    uint   `json:"sender_user_id" gorm:"not null;index"`
	Body            string `json:"body" gorm:"type:text;not null"`
	MessageType     string `json:"message_type" gorm:"type:varchar(32);not null;default:'text';index"`
	SessionOfferID  *uint  `json:"session_offer_id" gorm:"index"`
	// SharedContentType/SharedContentID point a "content_share" message at the
	// workout/routine/meal_plan the trainer just sent — enough for the client
	// to build a click-through link, no extra join needed.
	SharedContentType string `json:"shared_content_type,omitempty" gorm:"type:varchar(20)"`
	SharedContentID   *uint  `json:"shared_content_id,omitempty"`
}

func (PTPackageChatMessage) TableName() string { return "pt_package_chat_messages" }

// PTPackageChatRead tracks the last chat message a user has seen on a package thread.
type PTPackageChatRead struct {
	BaseEntity
	UserID            uint `json:"user_id" gorm:"not null;uniqueIndex:idx_pt_chat_read_user_pkg"`
	UserPTPackageID   uint `json:"user_pt_package_id" gorm:"not null;uniqueIndex:idx_pt_chat_read_user_pkg;index"`
	LastReadMessageID uint `json:"last_read_message_id" gorm:"not null;default:0"`
}

func (PTPackageChatRead) TableName() string { return "pt_package_chat_reads" }
