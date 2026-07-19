package service

import (
	"context"
	"strings"

	authorv1 "trongcon-api/api/author/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"
)

func displayNameFromUser(u *entity.User) string {
	if u == nil {
		return ""
	}
	if n := strings.TrimSpace(u.Name); n != "" {
		return n
	}
	n := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	if n != "" {
		return n
	}
	return u.Email
}

// authorForUserID resolves a public author badge for any poster.
// Prefers a public TrainerProfile (links to /trainers/:id); otherwise falls
// back to the user's name/avatar with trainer_id=0 (no profile link).
func authorForUserID(ctx context.Context, trainerRepo repository.TrainerProfileRepository, userRepo repository.UserRepository, userID uint) *authorv1.AuthorRes {
	if userID == 0 {
		return nil
	}
	if trainerRepo != nil {
		tp, err := trainerRepo.GetByUserID(ctx, userID)
		if err == nil && tp != nil && tp.IsPublic {
			return &authorv1.AuthorRes{
				TrainerID:   tp.ID,
				UserID:      tp.UserID,
				DisplayName: tp.DisplayName,
				AvatarURL:   tp.User.ProfilePicture,
				Title:       tp.Title,
			}
		}
	}
	if userRepo == nil {
		return nil
	}
	u, err := userRepo.GetByID(ctx, userID)
	if err != nil || u == nil {
		return nil
	}
	return &authorv1.AuthorRes{
		TrainerID:   0,
		UserID:      u.ID,
		DisplayName: displayNameFromUser(u),
		AvatarURL:   u.ProfilePicture,
	}
}

// workoutAuthorID prefers explicit UserID (poster), then personal OwnerUserID.
func workoutAuthorID(w *entity.Workout) uint {
	if w == nil {
		return 0
	}
	if w.UserID > 0 {
		return w.UserID
	}
	if w.OwnerUserID != nil {
		return *w.OwnerUserID
	}
	return 0
}
