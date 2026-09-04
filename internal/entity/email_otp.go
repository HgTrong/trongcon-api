package entity

import "time"

// EmailOTP stores one-time codes (e.g. forgot-password).
type EmailOTP struct {
	BaseEntity
	Email    string    `json:"email" gorm:"type:varchar(255);not null;index"`
	OTP      string    `json:"otp" gorm:"type:varchar(16);not null"`
	Purpose  string    `json:"purpose" gorm:"type:varchar(64);not null;default:'forgot_password';index"`
	Used     bool      `json:"used" gorm:"not null;default:false;index"`
	ExpireAt time.Time `json:"expire_at" gorm:"not null;index"`
}

func (EmailOTP) TableName() string { return "email_otps" }

const EmailOTPPurposeForgotPassword = "forgot_password"
const EmailOTPPurposeSignup = "signup"
