package v1

import userv1 "trongcon-api/api/user/v1"

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SignupReq struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender" binding:"omitempty,oneof=male female other prefer_not_to_say"`
	Language  string `json:"language"`
	OTP       string `json:"otp" binding:"required,min=4,max=16"`
}

type SignupRequestOTPReq struct {
	Email string `json:"email" binding:"required,email"`
}

type SignupRequestOTPRes struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type LoginRes struct {
	Token string         `json:"token"`
	User  userv1.UserRes `json:"user"`
}

type ForgotPasswordReq struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordRes struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type VerifyForgotOTPReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,min=4,max=16"`
}

type VerifyForgotOTPRes struct {
	SecretToken string `json:"secret_token"`
}

type ResetPasswordReq struct {
	SecretToken string `json:"secret_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ResetPasswordRes struct {
	Status string `json:"status"`
}
