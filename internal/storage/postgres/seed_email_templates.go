package postgres

import (
	"log"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

func seedTransactionalEmailTemplates(db *gorm.DB) error {
	templates := []entity.EmailTemplate{
		{
			Key: "gym_membership_purchased", Name: "Mua thẻ hội viên", IsActive: true,
			Subject: "TrongCon — Thẻ hội viên đã kích hoạt",
			Body: `<p>Xin chào {{.UserName}},</p>
<p>Thẻ hội viên <strong>{{.PlanName}}</strong> của bạn đã được kích hoạt.</p>
<p>Hiệu lực: {{.StartDate}} → {{.EndDate}}</p>
<p>Bạn có thể đặt lớp nhóm và dùng QR check-in tại phòng tập. Premium app cũng được mở theo thời hạn thẻ.</p>`,
		},
		{
			Key: "pt_package_purchased", Name: "Mua gói PT", IsActive: true,
			Subject: "TrongCon — Gói PT đã kích hoạt",
			Body: `<p>Xin chào {{.UserName}},</p>
<p>Gói <strong>{{.PackageTitle}}</strong> với HLV {{.TrainerName}} đã sẵn sàng.</p>
<p>Mở chat gói để đặt lịch / đề xuất buổi tập.</p>`,
		},
		{
			Key: "pt_session_proposed", Name: "Đề xuất buổi PT", IsActive: true,
			Subject: "TrongCon — Có đề xuất buổi tập mới",
			Body: `<p>Xin chào {{.UserName}},</p>
<p>{{.FromName}} vừa đề xuất buổi tập lúc <strong>{{.StartsAt}}</strong>.</p>
<p>Vào chat gói để chấp nhận hoặc từ chối.</p>`,
		},
		{
			Key: "pt_session_confirmed", Name: "Xác nhận buổi PT", IsActive: true,
			Subject: "TrongCon — Buổi tập đã được xác nhận",
			Body: `<p>Xin chào {{.UserName}},</p>
<p>Buổi tập lúc <strong>{{.StartsAt}}</strong> đã được xác nhận hoàn thành.</p>`,
		},
	}
	for _, t := range templates {
		var n int64
		if err := db.Model(&entity.EmailTemplate{}).Where("key = ?", t.Key).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		if err := db.Create(&t).Error; err != nil {
			return err
		}
		log.Printf("seed: email template %s", t.Key)
	}
	return nil
}
