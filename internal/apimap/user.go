package apimap

import (
	userv1 "trongcon-api/api/user/v1"
	"trongcon-api/internal/entity"
)

func RoleNames(u *entity.User) []string {
	out := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		out = append(out, r.Name)
	}
	return out
}

func UserToRes(u *entity.User) userv1.UserRes {
	return userv1.UserRes{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Gender:      u.Gender,
		Language:    u.Language,
		MemberSince: u.CreatedAt,
		LastLogin:   u.LastLoginAt,
		AccountType: u.AccountType,
		Roles:       RoleNames(u),
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
