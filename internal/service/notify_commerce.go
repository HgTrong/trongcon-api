package service

import (
	"context"

	"trongcon-api/internal/entity"
	"trongcon-api/internal/money"
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
	var trainerUserID uint
	if up.PTPackage.Trainer.ID != 0 {
		trainerName = up.PTPackage.Trainer.DisplayName
		trainerUserID = up.PTPackage.Trainer.UserID
	}
	if trainerUserID == 0 {
		if t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID); err == nil && t != nil {
			trainerName = t.DisplayName
			trainerUserID = t.UserID
		}
	}
	s.notifyEmail(ctx, "pt_package_purchased", map[string]interface{}{
		"UserName":     name,
		"PackageTitle": title,
		"TrainerName":  trainerName,
	}, email)

	// Seller-facing copy — the PT should hear about a sale without having to
	// keep the Studio PT tab open all day.
	if trainerUserID != 0 {
		trainerEmail, trainerRealName := s.userEmail(ctx, trainerUserID)
		s.notifyEmail(ctx, "pt_package_sold", map[string]interface{}{
			"TrainerName":  trainerRealName,
			"StudentName":  name,
			"PackageTitle": title,
			"SessionCount": up.SessionTotal,
			"Price":        money.FormatVND(up.Price),
		}, trainerEmail)
	}
}
