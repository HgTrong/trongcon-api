package v1

type ProfileUpdateReq struct {
	Email     string `json:"email" binding:"omitempty,email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender" binding:"omitempty,oneof=male female other prefer_not_to_say"`
	Language  string `json:"language"`
}

type ChangePasswordReq struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordRes struct {
	Status string `json:"status"`
}

type UpdateAvatarRes struct {
	User UserRes `json:"user"`
}
