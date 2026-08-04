package service

import (
	"context"

	"trongcon-api/internal/entity"
)

func (s *gymCommerceService) notifyPTPackagePurchased(ctx context.Context, up *entity.UserPTPackage) {
	if up == nil {
		return
	}
	email, name := s.userEmail(ctx, up.UserID)
	title, trainerName := "", ""
	if up.PTPackage.ID != 0 {
		title = up.PTPackage.Title
	}
	if up.PTPackage.Trainer.ID != 0 {
		trainerName = up.PTPackage.Trainer.DisplayName
	} else if t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID); err == nil && t != nil {
		trainerName = t.DisplayName
	}
	s.notifyEmail(ctx, "pt_package_purchased", map[string]interface{}{
		"UserName":     name,
		"PackageTitle": title,
		"TrainerName":  trainerName,
	}, email)
}
