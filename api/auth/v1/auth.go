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
}

type LoginRes struct {
	Token string       `json:"token"`
	User  userv1.UserRes `json:"user"`
}
