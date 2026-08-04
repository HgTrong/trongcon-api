package entity

import "time"

// GymCheckIn is a staff-verified floor entry against an active gym membership.
type GymCheckIn struct {
	BaseEntity
	UserID              uint      `json:"user_id" gorm:"not null;index"`
	UserGymMembershipID uint      `json:"user_gym_membership_id" gorm:"not null;index"`
	BranchID            *uint     `json:"branch_id" gorm:"index"`
	CheckedInAt         time.Time `json:"checked_in_at" gorm:"not null;index"`
	VerifiedByUserID    uint      `json:"verified_by_user_id" gorm:"index"`
	Note                string    `json:"note" gorm:"type:varchar(255)"`
}

func (GymCheckIn) TableName() string { return "gym_check_ins" }
