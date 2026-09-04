package postgres

import (
	"log"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

// ptPackageSoldBody is a self-contained, inline-styled HTML email — SES SMTP
// sends this string verbatim as the message's HTML body, with no surrounding
// layout, so every template needs its own full styling (email clients strip
// <style> blocks and external CSS, hence inline styles + a table-based shell
// for cross-client layout stability).
const ptPackageSoldBody = `<div style="background:#f4efe6;padding:32px 16px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;margin:0 auto;">
<tr><td style="background:#ffffff;border-radius:14px;overflow:hidden;box-shadow:0 1px 3px rgba(20,15,10,.08);">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0">
<tr><td style="background:#da1f27;padding:26px 32px;">
<span style="color:#ffffff;font-size:19px;font-weight:700;letter-spacing:.02em;">TrongCon</span>
</td></tr>
<tr><td style="padding:32px;">
<p style="margin:0 0 6px;color:#a1927a;font-size:12px;font-weight:700;letter-spacing:.1em;text-transform:uppercase;">Đơn mới</p>
<h1 style="margin:0 0 18px;color:#211c15;font-size:22px;font-weight:800;line-height:1.3;">Bạn vừa bán được 1 gói PT 🎉</h1>
<p style="margin:0 0 24px;color:#4a4239;font-size:15px;line-height:1.65;">Xin chào {{.TrainerName}}, học viên <strong style="color:#211c15;">{{.StudentName}}</strong> vừa thanh toán thành công một gói tập với bạn.</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#faf6ef;border:1px solid #eee3d2;border-radius:12px;">
<tr><td style="padding:22px 24px;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0">
<tr>
<td style="color:#a1927a;font-size:13px;padding-bottom:12px;">Gói tập</td>
<td style="color:#211c15;font-size:14px;font-weight:700;text-align:right;padding-bottom:12px;">{{.PackageTitle}}</td>
</tr>
<tr>
<td style="color:#a1927a;font-size:13px;padding:12px 0;border-top:1px solid #eee3d2;">Số buổi</td>
<td style="color:#211c15;font-size:14px;font-weight:700;text-align:right;padding:12px 0;border-top:1px solid #eee3d2;">{{.SessionCount}} buổi</td>
</tr>
<tr>
<td style="color:#a1927a;font-size:13px;padding-top:12px;border-top:1px solid #eee3d2;">Giá trị đơn</td>
<td style="color:#da1f27;font-size:21px;font-weight:800;text-align:right;padding-top:12px;border-top:1px solid #eee3d2;">{{.Price}}</td>
</tr>
</table>
</td></tr>
</table>
<p style="margin:24px 0 0;color:#4a4239;font-size:14px;line-height:1.6;">Mở <strong>Studio PT</strong> trong app để bắt đầu trao đổi lịch tập với học viên nhé.</p>
</td></tr>
<tr><td style="padding:18px 32px;background:#faf6ef;border-top:1px solid #eee3d2;">
<p style="margin:0;color:#b0a494;font-size:12px;">Email tự động từ TrongCon — không cần trả lời email này.</p>
</td></tr>
</table>
</td></tr>
</table>
</div>`

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
			Key: "pt_package_sold", Name: "Bán được gói PT (báo HLV)", IsActive: true,
			Subject: "TrongCon — Bạn vừa bán được 1 gói PT",
			Body:    ptPackageSoldBody,
		},
		{
			Key: "pt_recurring_booking_created", Name: "Học viên đăng ký lịch cố định", IsActive: true,
			Subject: "TrongCon — Học viên vừa đăng ký lịch cố định hàng tuần",
			Body: `<p>Xin chào {{.TrainerName}},</p>
<p>Học viên <strong>{{.StudentName}}</strong> vừa đăng ký lịch tập cố định <strong>mỗi {{.Weekday}} lúc {{.TimeRange}}</strong>
cho gói <strong>{{.PackageTitle}}</strong>. Hệ thống đã tự xếp lịch những tuần tới — bạn không cần xác nhận từng tuần.</p>
<p>Nếu không phù hợp, vào Studio PT → Lịch nhận khách → mục "Lịch cố định của học viên" để hủy — nên xử lý sớm trong 48 giờ đầu,
trước khi buổi tập đầu tiên diễn ra.</p>`,
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
		{
			Key: "pt_session_proof_rejected", Name: "Bằng chứng buổi PT bị từ chối", IsActive: true,
			Subject: "TrongCon — Cần gửi lại bằng chứng buổi tập",
			Body: `<p>Xin chào {{.UserName}},</p>
<p>Bằng chứng buổi tập lúc <strong>{{.StartsAt}}</strong> chưa được duyệt.</p>
<p>Vào chat gói để gửi lại ảnh bằng chứng cho buổi này.</p>`,
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

// upgradeDefaultEmailTemplates re-ships an improved default Subject/Body for
// an existing transactional template — but only if nobody has customized it
// via the admin UI yet (updated_at == created_at, i.e. untouched since it was
// first seeded). This lets us polish a template's design later without
// clobbering an admin's edits, and runs itself out of a job the first time it
// touches a row (updated_at then differs from created_at for good).
func upgradeDefaultEmailTemplates(db *gorm.DB) error {
	upgrades := []struct {
		Key     string
		Subject string
		Body    string
	}{
		{Key: "pt_package_sold", Subject: "TrongCon — Bạn vừa bán được 1 gói PT", Body: ptPackageSoldBody},
	}
	for _, u := range upgrades {
		res := db.Exec(
			`UPDATE email_templates SET subject = ?, body = ?, updated_at = NOW() WHERE key = ? AND updated_at = created_at`,
			u.Subject, u.Body, u.Key,
		)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			log.Printf("migrate: refreshed default email template %s", u.Key)
		}
	}
	return nil
}
