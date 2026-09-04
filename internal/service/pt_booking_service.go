package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	gcv1 "trongcon-api/api/gym_commerce/v1"
	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

// errSlotTaken signals the transactional re-check inside BookSlot /
// materializeOccurrence found a conflict that appeared after the initial
// (non-transactional) availability check — i.e. someone else won the race.
var errSlotTaken = errors.New("slot taken")

func vnLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return time.FixedZone("ICT", 7*3600)
	}
	return loc
}

func (s *gymCommerceService) sessionDuration(t *entity.TrainerProfile) time.Duration {
	min := t.SessionDurationMin
	if min < 15 {
		min = 60
	}
	if min > 240 {
		min = 240
	}
	return time.Duration(min) * time.Minute
}

func offerEnd(o *entity.PTSessionOffer, fallback time.Duration) time.Time {
	if o.EndsAt != nil && !o.EndsAt.IsZero() {
		return o.EndsAt.UTC()
	}
	return o.StartsAt.UTC().Add(fallback)
}

func rangesOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

// assertStudentNotBusy blocks double-booking on the student's side: without this,
// nothing stops a student from accepting overlapping session times with two
// different trainers (the existing busy-check only guards one trainer's calendar).
// ignoreOfferID excludes the offer being accepted/booked itself from the check.
func (s *gymCommerceService) assertStudentNotBusy(ctx context.Context, studentUserID, ignoreOfferID uint, startsAt, endsAt time.Time) error {
	busy, err := s.offerRepo.ListBusyInRangeForStudent(ctx, studentUserID, startsAt.Add(-time.Minute), endsAt.Add(time.Minute))
	if err != nil {
		return err
	}
	for i := range busy {
		if busy[i].ID == ignoreOfferID {
			continue
		}
		if rangesOverlap(startsAt.UTC(), endsAt.UTC(), busy[i].StartsAt.UTC(), offerEnd(&busy[i], endsAt.Sub(startsAt))) {
			return fmt.Errorf("bạn đã có một buổi tập khác vào thời gian này")
		}
	}
	return nil
}

func (s *gymCommerceService) assertCanPurchasePTPackage(ctx context.Context, userID uint, trainerProfileID uint) error {
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return notFoundOr(err, "không tìm thấy huấn luyện viên")
	}
	// BookingPaused means the trainer stopped taking any new session bookings
	// (e.g. on leave) — it applies to existing clients too, unlike AcceptingNewClients.
	if t.BookingPaused {
		return fmt.Errorf("huấn luyện viên đang tạm dừng nhận lịch")
	}
	existing, err := s.userPtPkgRepo.HasActivePackage(ctx, trainerProfileID, userID)
	if err != nil {
		return err
	}
	if existing {
		return nil
	}
	if !t.AcceptingNewClients {
		return fmt.Errorf("huấn luyện viên hiện không nhận học viên mới")
	}
	if t.MaxActiveClients > 0 {
		active, err := s.userPtPkgRepo.CountActiveClients(ctx, t.ID)
		if err != nil {
			return err
		}
		if int(active) >= t.MaxActiveClients {
			return fmt.Errorf("huấn luyện viên đã đạt số lượng học viên tối đa")
		}
	}
	return nil
}

func (s *gymCommerceService) bookingSettingsRes(ctx context.Context, t *entity.TrainerProfile) (*gcv1.BookingSettingsRes, error) {
	active, err := s.userPtPkgRepo.CountActiveClients(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	dur := t.SessionDurationMin
	if dur <= 0 {
		dur = 60
	}
	return &gcv1.BookingSettingsRes{
		TrainerProfileID:    t.ID,
		SessionDurationMin:  dur,
		AcceptingNewClients: t.AcceptingNewClients,
		BookingPaused:       t.BookingPaused,
		ActiveClients:       active,
	}, nil
}

func (s *gymCommerceService) GetMyBookingSettings(ctx context.Context, trainerUserID uint) (*gcv1.BookingSettingsRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	return s.bookingSettingsRes(ctx, t)
}

func (s *gymCommerceService) UpdateMyBookingSettings(ctx context.Context, trainerUserID uint, req *gcv1.BookingSettingsReq) (*gcv1.BookingSettingsRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	if req.SessionDurationMin != nil {
		d := *req.SessionDurationMin
		if d < 15 || d > 240 {
			return nil, fmt.Errorf("thời gian buổi tập phải trong khoảng 15 đến 240 phút")
		}
		t.SessionDurationMin = d
	}
	if req.AcceptingNewClients != nil {
		t.AcceptingNewClients = *req.AcceptingNewClients
	}
	if req.BookingPaused != nil {
		t.BookingPaused = *req.BookingPaused
	}
	if err := s.trainerRepo.Update(ctx, t); err != nil {
		return nil, err
	}
	return s.bookingSettingsRes(ctx, t)
}

func (s *gymCommerceService) GetMyWorkingHours(ctx context.Context, trainerUserID uint) (*gcv1.ListRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	rows, err := s.hoursRepo.ListAllByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.WorkingHoursItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, gcv1.WorkingHoursItem{
			Weekday:     r.Weekday,
			StartMinute: r.StartMinute,
			EndMinute:   r.EndMinute,
			IsActive:    r.IsActive,
		})
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) SetMyWorkingHours(ctx context.Context, trainerUserID uint, req *gcv1.SetWorkingHoursReq) (*gcv1.ListRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	rows := make([]entity.PTWorkingHours, 0, len(req.Hours))
	byDay := map[int][][2]int{}
	for _, h := range req.Hours {
		if h.Weekday < 0 || h.Weekday > 6 {
			return nil, fmt.Errorf("thứ trong tuần không hợp lệ")
		}
		if h.StartMinute < 0 || h.EndMinute > 24*60 || h.EndMinute <= h.StartMinute {
			return nil, fmt.Errorf("khung giờ không hợp lệ cho thứ %d trong tuần", h.Weekday)
		}
		rows = append(rows, entity.PTWorkingHours{
			TrainerProfileID: t.ID,
			Weekday:          h.Weekday,
			StartMinute:      h.StartMinute,
			EndMinute:        h.EndMinute,
			IsActive:         h.IsActive,
		})
		if h.IsActive {
			byDay[h.Weekday] = append(byDay[h.Weekday], [2]int{h.StartMinute, h.EndMinute})
		}
	}
	for day, ranges := range byDay {
		sort.Slice(ranges, func(i, j int) bool {
			if ranges[i][0] == ranges[j][0] {
				return ranges[i][1] < ranges[j][1]
			}
			return ranges[i][0] < ranges[j][0]
		})
		for i := 1; i < len(ranges); i++ {
			if ranges[i][0] < ranges[i-1][1] {
				return nil, fmt.Errorf("các khung giờ bị trùng nhau vào thứ %d trong tuần", day)
			}
		}
	}
	if err := s.hoursRepo.ReplaceForTrainer(ctx, t.ID, rows); err != nil {
		return nil, err
	}
	return s.GetMyWorkingHours(ctx, trainerUserID)
}

func (s *gymCommerceService) ListMyBlockedSlots(ctx context.Context, trainerUserID uint, from, to time.Time) (*gcv1.ListRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("cần chọn khoảng thời gian từ/đến")
	}
	rows, err := s.blockedRepo.ListInRange(ctx, t.ID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.BlockedSlotRes, 0, len(rows))
	for _, b := range rows {
		out = append(out, gcv1.BlockedSlotRes{
			ID: b.ID, TrainerProfileID: b.TrainerProfileID,
			StartsAt: b.StartsAt, EndsAt: b.EndsAt, Reason: b.Reason,
		})
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) BlockMySlot(ctx context.Context, trainerUserID uint, req *gcv1.BlockSlotReq) (*gcv1.BlockedSlotRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	start, end := req.StartsAt.UTC(), req.EndsAt.UTC()
	if !end.After(start) {
		return nil, fmt.Errorf("giờ kết thúc phải sau giờ bắt đầu")
	}
	b := &entity.PTBlockedSlot{
		TrainerProfileID: t.ID,
		StartsAt:         start,
		EndsAt:           end,
		Reason:           strings.TrimSpace(req.Reason),
	}
	if err := s.blockedRepo.Create(ctx, b); err != nil {
		return nil, err
	}
	return &gcv1.BlockedSlotRes{
		ID: b.ID, TrainerProfileID: b.TrainerProfileID,
		StartsAt: b.StartsAt, EndsAt: b.EndsAt, Reason: b.Reason,
	}, nil
}

func (s *gymCommerceService) UnblockMySlot(ctx context.Context, trainerUserID, blockID uint) error {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return err
	}
	return s.blockedRepo.Delete(ctx, blockID, t.ID)
}

func (s *gymCommerceService) ListAvailableSlots(ctx context.Context, trainerProfileID uint, from, to time.Time) (*gcv1.ListRes, error) {
	slots, err := s.generateAvailableSlots(ctx, trainerProfileID, from, to)
	if err != nil {
		return nil, err
	}
	return &gcv1.ListRes{Total: int64(len(slots)), Data: slots}, nil
}

func (s *gymCommerceService) generateAvailableSlots(ctx context.Context, trainerProfileID uint, from, to time.Time) ([]gcv1.AvailableSlotRes, error) {
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("cần chọn khoảng thời gian từ/đến")
	}
	if to.Sub(from) > 14*24*time.Hour {
		return nil, fmt.Errorf("khoảng thời gian không được vượt quá 14 ngày")
	}
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "không tìm thấy huấn luyện viên")
	}
	if t.BookingPaused {
		return []gcv1.AvailableSlotRes{}, nil
	}
	hours, err := s.hoursRepo.ListByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	if len(hours) == 0 {
		return []gcv1.AvailableSlotRes{}, nil
	}
	byDay := map[int][]entity.PTWorkingHours{}
	for _, h := range hours {
		byDay[h.Weekday] = append(byDay[h.Weekday], h)
	}
	dur := s.sessionDuration(t)
	fromUTC, toUTC := from.UTC(), to.UTC()
	busy, err := s.offerRepo.ListBusyInRange(ctx, t.ID, fromUTC, toUTC)
	if err != nil {
		return nil, err
	}
	blocked, err := s.blockedRepo.ListInRange(ctx, t.ID, fromUTC, toUTC)
	if err != nil {
		return nil, err
	}
	loc := vnLocation()
	now := time.Now().UTC()
	out := make([]gcv1.AvailableSlotRes, 0)
	day := time.Date(from.In(loc).Year(), from.In(loc).Month(), from.In(loc).Day(), 0, 0, 0, 0, loc)
	endDay := time.Date(to.In(loc).Year(), to.In(loc).Month(), to.In(loc).Day(), 0, 0, 0, 0, loc)
	for d := day; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		windows := byDay[int(d.Weekday())]
		for _, h := range windows {
			if !h.IsActive {
				continue
			}
			windowEnd := d.Add(time.Duration(h.EndMinute) * time.Minute)
			for slotStart := d.Add(time.Duration(h.StartMinute) * time.Minute); !slotStart.Add(dur).After(windowEnd); slotStart = slotStart.Add(dur) {
				slotEnd := slotStart.Add(dur)
				sUTC, eUTC := slotStart.UTC(), slotEnd.UTC()
				if eUTC.Before(fromUTC) || !sUTC.Before(toUTC) {
					continue
				}
				if !sUTC.After(now) {
					continue
				}
				conflict := false
				for i := range busy {
					bStart := busy[i].StartsAt.UTC()
					bEnd := offerEnd(&busy[i], dur)
					if rangesOverlap(sUTC, eUTC, bStart, bEnd) {
						conflict = true
						break
					}
				}
				if conflict {
					continue
				}
				for i := range blocked {
					if rangesOverlap(sUTC, eUTC, blocked[i].StartsAt.UTC(), blocked[i].EndsAt.UTC()) {
						conflict = true
						break
					}
				}
				if conflict {
					continue
				}
				out = append(out, gcv1.AvailableSlotRes{StartsAt: sUTC, EndsAt: eUTC})
			}
		}
	}
	return out, nil
}

// tryReserveSlot re-checks, inside an advisory-locked transaction, that no
// busy offer or blocked window overlaps [startsAt,endsAt) for this trainer.
// Callers must already hold the pg_advisory_xact_lock for trainerProfileID
// on tx — this only does the read; it never locks or inserts on its own.
func (s *gymCommerceService) tryReserveSlot(ctx context.Context, tx *gorm.DB, trainerProfileID uint, startsAt, endsAt time.Time, dur time.Duration) (bool, error) {
	var busy []entity.PTSessionOffer
	if err := tx.WithContext(ctx).
		Where("trainer_profile_id = ? AND status IN ? AND starts_at < ?",
			trainerProfileID,
			[]string{entity.SessionOfferPending, entity.SessionOfferScheduled, entity.SessionOfferAwaitingConfirmation},
			endsAt,
		).Find(&busy).Error; err != nil {
		return false, err
	}
	for i := range busy {
		if rangesOverlap(startsAt, endsAt, busy[i].StartsAt.UTC(), offerEnd(&busy[i], dur)) {
			return false, nil
		}
	}
	var blockedCount int64
	if err := tx.WithContext(ctx).Model(&entity.PTBlockedSlot{}).
		Where("trainer_profile_id = ? AND starts_at < ? AND ends_at > ?", trainerProfileID, endsAt, startsAt).
		Count(&blockedCount).Error; err != nil {
		return false, err
	}
	return blockedCount == 0, nil
}

func (s *gymCommerceService) BookSlot(ctx context.Context, studentUserID uint, req *gcv1.BookSlotReq) (*gcv1.SessionOfferRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, req.UserPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "không tìm thấy gói tập của người dùng")
	}
	if up.UserID != studentUserID {
		return nil, fmt.Errorf("gói tập không thuộc về người dùng này")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("gói tập không còn hoạt động")
	}
	if err := s.assertSessionsAvailable(ctx, up); err != nil {
		return nil, err
	}
	t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "không tìm thấy huấn luyện viên")
	}
	if t.BookingPaused {
		return nil, fmt.Errorf("huấn luyện viên đang tạm dừng nhận lịch")
	}
	startsAt := req.StartsAt.UTC().Truncate(time.Minute)
	if !startsAt.After(time.Now().UTC()) {
		return nil, fmt.Errorf("giờ đặt phải ở trong tương lai")
	}
	dur := s.sessionDuration(t)
	endsAt := startsAt.Add(dur)

	slots, err := s.generateAvailableSlots(ctx, t.ID, startsAt.Add(-time.Minute), endsAt.Add(time.Minute))
	if err != nil {
		return nil, err
	}
	ok := false
	for _, sl := range slots {
		if sl.StartsAt.Equal(startsAt) {
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("khung giờ này không còn trống")
	}
	if err := s.assertStudentNotBusy(ctx, studentUserID, 0, startsAt, endsAt); err != nil {
		return nil, err
	}

	// The check above ran against a snapshot — re-verify inside an
	// advisory-locked transaction right before inserting, so two students
	// racing for the same slot can't both win (see tryReserveSlot).
	now := time.Now().UTC()
	offer := &entity.PTSessionOffer{
		UserPTPackageID:  up.ID,
		TrainerProfileID: up.TrainerProfileID,
		StudentUserID:    up.UserID,
		StartsAt:         startsAt,
		EndsAt:           &endsAt,
		Note:             strings.TrimSpace(req.Note),
		ProposedByUserID: studentUserID,
		Status:           entity.SessionOfferScheduled,
		AcceptedByUserID: studentUserID,
		AcceptedAt:       &now,
		BookedViaSlot:    true,
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(t.ID)).Error; err != nil {
			return err
		}
		free, err := s.tryReserveSlot(ctx, tx, t.ID, startsAt, endsAt, dur)
		if err != nil {
			return err
		}
		if !free {
			return errSlotTaken
		}
		return tx.Create(offer).Error
	})
	if err != nil {
		if errors.Is(err, errSlotTaken) {
			return nil, fmt.Errorf("rất tiếc, khung giờ này vừa có người đặt trước bạn — vui lòng chọn khung giờ khác")
		}
		return nil, err
	}
	offerID := offer.ID
	body := fmt.Sprintf("Booked session slot: %s – %s", startsAt.Format(time.RFC3339), endsAt.Format(time.RFC3339))
	if offer.Note != "" {
		body += "\n" + offer.Note
	}
	msg := &entity.PTPackageChatMessage{
		UserPTPackageID: up.ID,
		SenderUserID:    studentUserID,
		Body:            body,
		MessageType:     entity.ChatMsgTypeSessionOffer,
		SessionOfferID:  &offerID,
	}
	_ = s.chatRepo.Create(ctx, msg)
	if s.growth != nil {
		s.growth.TrackBooking(ctx, up.TrainerProfileID, studentUserID, req.SourceContentType, req.SourceContentID, req.SourceTitle)
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

// ExpireStalePendingOffers cancels session offers left "pending" (never accepted or
// declined) past olderThan — otherwise a forgotten proposal keeps blocking the
// trainer's slot and the package's session-credit indefinitely.
func (s *gymCommerceService) ExpireStalePendingOffers(ctx context.Context, olderThan time.Duration) (int, error) {
	if olderThan <= 0 {
		olderThan = 48 * time.Hour
	}
	before := time.Now().UTC().Add(-olderThan)
	n, err := s.offerRepo.ExpireStalePending(ctx, before)
	return int(n), err
}

// RunExpiryHousekeeping actively flips memberships/PT packages whose end date has
// passed to "expired". Without this it only happened lazily on the next read of
// each list, so a user who doesn't open the app could look "active" indefinitely.
func (s *gymCommerceService) RunExpiryHousekeeping(ctx context.Context) error {
	now := time.Now().UTC()
	if err := s.membRepo.ExpireEnded(ctx, now); err != nil {
		return fmt.Errorf("expire gym memberships: %w", err)
	}
	if err := s.userPtPkgRepo.ExpireEnded(ctx, now); err != nil {
		return fmt.Errorf("expire pt packages: %w", err)
	}
	return nil
}

// CancelStalePendingOrders cancels gym-membership and PT-package orders that were
// created but never paid (still "pending") past olderThan — otherwise an abandoned
// Stripe/VNPay checkout stays "pending" forever instead of freeing up for a retry.
func (s *gymCommerceService) CancelStalePendingOrders(ctx context.Context, olderThan time.Duration) (int, error) {
	if olderThan <= 0 {
		olderThan = 6 * time.Hour
	}
	before := time.Now().UTC().Add(-olderThan)
	total := 0
	n, err := s.membRepo.CancelStalePending(ctx, before)
	if err != nil {
		return total, fmt.Errorf("cancel stale pending memberships: %w", err)
	}
	total += int(n)
	n, err = s.userPtPkgRepo.CancelStalePending(ctx, before)
	if err != nil {
		return total, fmt.Errorf("cancel stale pending pt packages: %w", err)
	}
	total += int(n)
	return total, nil
}

func (s *gymCommerceService) AutoConfirmExpiredSessionProofs(ctx context.Context, olderThan time.Duration, limit int) (int, error) {
	if olderThan <= 0 {
		olderThan = 24 * time.Hour
	}
	before := time.Now().UTC().Add(-olderThan)
	rows, err := s.offerRepo.ListAwaitingConfirmationOlderThan(ctx, before, limit)
	if err != nil {
		return 0, err
	}
	confirmed := 0
	for i := range rows {
		o := &rows[i]
		if _, err := s.finalizeSessionConfirmation(ctx, o, 0); err != nil {
			continue
		}
		confirmed++
	}
	return confirmed, nil
}

// MarkSessionNoShow: PT marks a scheduled session as student no-show; still consumes 1 session credit.
func (s *gymCommerceService) MarkSessionNoShow(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error) {
	offer, up, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	t, err := s.trainerProfileForUser(ctx, requesterUserID)
	if err != nil || t.ID != up.TrainerProfileID {
		return nil, fmt.Errorf("chỉ huấn luyện viên của gói tập này mới có thể đánh dấu vắng mặt")
	}
	if offer.Status != entity.SessionOfferScheduled {
		return nil, fmt.Errorf("buổi tập phải ở trạng thái đã lên lịch để đánh dấu vắng mặt")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("gói tập không còn hoạt động")
	}
	if up.SessionUsed >= up.SessionTotal {
		return nil, fmt.Errorf("đã hết số buổi trong gói tập")
	}
	now := time.Now().UTC()
	if now.Before(offer.StartsAt) {
		return nil, fmt.Errorf("không thể đánh dấu vắng mặt trước giờ hẹn của buổi tập")
	}
	nextIndex := up.SessionUsed + 1
	note := strings.TrimSpace(offer.Note)
	if note == "" {
		note = "Vắng mặt"
	} else {
		note = "Vắng mặt — " + note
	}
	logRow := &entity.PTSessionLog{
		UserPTPackageID:  up.ID,
		TrainerProfileID: up.TrainerProfileID,
		UserID:           up.UserID,
		SessionIndex:     nextIndex,
		TaughtAt:         offer.StartsAt,
		Note:             note,
		ProofImageURL:    "noshow://marked",
		CreatedByUserID:  requesterUserID,
	}
	if err := s.sessionLogRepo.Create(ctx, logRow); err != nil {
		return nil, err
	}
	up.SessionUsed = nextIndex
	if err := s.userPtPkgRepo.Update(ctx, up); err != nil {
		return nil, err
	}
	offer.Status = entity.SessionOfferNoShow
	offer.Note = note
	offer.CompletedAt = &now
	offer.CompletedByUserID = requesterUserID
	offer.SessionIndex = nextIndex
	offer.ProofImageURL = ""
	offer.ProofSubmittedAt = nil
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	// PT still gets paid for a no-show: they were present and ready, the
	// student didn't turn up. The session credit is consumed either way.
	if err := s.recordPTSessionEarning(ctx, up, nextIndex, offer.ID, note); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

func (s *gymCommerceService) finalizeSessionConfirmation(ctx context.Context, offer *entity.PTSessionOffer, confirmedByUserID uint) (*gcv1.SessionOfferRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, offer.UserPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "không tìm thấy gói tập của người dùng")
	}
	if offer.Status != entity.SessionOfferAwaitingConfirmation {
		return nil, fmt.Errorf("buổi tập không ở trạng thái chờ xác nhận")
	}
	proof := strings.TrimSpace(offer.ProofImageURL)
	if proof == "" {
		return nil, fmt.Errorf("thiếu ảnh minh chứng")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("gói tập không còn hoạt động")
	}
	if up.SessionUsed >= up.SessionTotal {
		return nil, fmt.Errorf("đã hết số buổi trong gói tập")
	}

	now := time.Now().UTC()
	nextIndex := up.SessionUsed + 1
	note := strings.TrimSpace(offer.Note)
	logRow := &entity.PTSessionLog{
		UserPTPackageID:  up.ID,
		TrainerProfileID: up.TrainerProfileID,
		UserID:           up.UserID,
		SessionIndex:     nextIndex,
		TaughtAt:         offer.StartsAt,
		Note:             note,
		ProofImageURL:    proof,
		CreatedByUserID:  offer.CompletedByUserID,
	}
	if logRow.CreatedByUserID == 0 {
		logRow.CreatedByUserID = confirmedByUserID
	}
	if err := s.sessionLogRepo.Create(ctx, logRow); err != nil {
		return nil, err
	}
	up.SessionUsed = nextIndex
	if err := s.userPtPkgRepo.Update(ctx, up); err != nil {
		return nil, err
	}
	offer.Status = entity.SessionOfferCompleted
	offer.CompletedAt = &now
	offer.SessionIndex = nextIndex
	offer.ConfirmedByUserID = confirmedByUserID
	offer.ConfirmedAt = &now
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	if err := s.recordPTSessionEarning(ctx, up, nextIndex, offer.ID, "session completed"); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	// Email both parties when possible.
	studentEmail, studentName := s.userEmail(ctx, offer.StudentUserID)
	s.notifyEmail(ctx, "pt_session_confirmed", map[string]interface{}{
		"UserName": studentName,
		"StartsAt": offer.StartsAt.In(vnLocation()).Format("15:04 02/01/2006"),
	}, studentEmail)
	if t, err := s.trainerRepo.GetByID(ctx, offer.TrainerProfileID); err == nil && t != nil {
		te, tn := s.userEmail(ctx, t.UserID)
		s.notifyEmail(ctx, "pt_session_confirmed", map[string]interface{}{
			"UserName": tn,
			"StartsAt": offer.StartsAt.In(vnLocation()).Format("15:04 02/01/2006"),
		}, te)
	}
	return &res, nil
}

func (s *gymCommerceService) AdminTrainerOpsOverview(ctx context.Context, trainerProfileID uint) (*gcv1.TrainerOpsOverviewRes, error) {
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "không tìm thấy huấn luyện viên")
	}
	active, err := s.userPtPkgRepo.CountActiveClients(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	counts, err := s.offerRepo.CountStatusesByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	loc := vnLocation()
	now := time.Now().In(loc)
	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}
	weekEnd := weekStart.AddDate(0, 0, 7)
	slots, err := s.ListAvailableSlots(ctx, t.ID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}
	status := "active"
	if t.BookingPaused {
		status = "paused"
	} else if !t.IsPublic {
		status = "private"
	}
	total := int64(0)
	for _, n := range counts {
		total += n
	}
	upcoming := counts[entity.SessionOfferScheduled] + counts[entity.SessionOfferPending] + counts[entity.SessionOfferAwaitingConfirmation]
	return &gcv1.TrainerOpsOverviewRes{
		TrainerProfileID:    t.ID,
		DisplayName:         t.DisplayName,
		Status:              status,
		AcceptingNewClients: t.AcceptingNewClients,
		BookingPaused:       t.BookingPaused,
		ActiveClients:       active,
		AvailableSlotsWeek:  int(slots.Total),
		TotalBookings:       total,
		Completed:           counts[entity.SessionOfferCompleted],
		Upcoming:            upcoming,
		Cancelled:           counts[entity.SessionOfferCancelled] + counts[entity.SessionOfferDeclined],
		AwaitingConfirm:     counts[entity.SessionOfferAwaitingConfirmation],
	}, nil
}

func (s *gymCommerceService) AdminListTrainerClients(ctx context.Context, trainerProfileID uint, page, limit int, status string) (*gcv1.ListRes, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "", "all":
		status = ""
	case entity.PTPkgStatusActive, entity.PTPkgStatusExpired, entity.PTPkgStatusCanceled, entity.PTPkgStatusPending:
		// ok
	default:
		return nil, fmt.Errorf("bộ lọc trạng thái không hợp lệ")
	}
	rows, total, err := s.userPtPkgRepo.ListByTrainerProfileID(ctx, trainerProfileID, (page-1)*limit, limit, status)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.TrainerClientRes, 0, len(rows))
	for i := range rows {
		p := &rows[i]
		name, email := "", ""
		if p.User.ID != 0 {
			email = p.User.Email
			name = strings.TrimSpace(p.User.Name)
			if name == "" {
				name = strings.TrimSpace(p.User.FirstName + " " + p.User.LastName)
			}
		}
		title := ""
		if p.PTPackage.ID != 0 {
			title = p.PTPackage.Title
		}
		out = append(out, gcv1.TrainerClientRes{
			UserID:          p.UserID,
			UserName:        name,
			UserEmail:       email,
			UserPTPackageID: p.ID,
			PackageTitle:    title,
			JoinedAt:        p.CreatedAt,
			SessionTotal:    p.SessionTotal,
			SessionUsed:     p.SessionUsed,
			Status:          p.Status,
		})
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) AdminListTrainerBookings(ctx context.Context, trainerProfileID uint, from, to time.Time) (*gcv1.ListRes, error) {
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("cần chọn khoảng thời gian từ/đến")
	}
	rows, err := s.offerRepo.ListByTrainerInRange(ctx, trainerProfileID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	names := s.userDisplayMap(ctx, studentIDsFromOffers(rows))
	out := make([]gcv1.SessionOfferRes, 0, len(rows))
	for i := range rows {
		res := toSessionOfferRes(&rows[i])
		if u, ok := names[rows[i].StudentUserID]; ok {
			res.StudentName = u.name
			res.StudentEmail = u.email
		}
		out = append(out, res)
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

type userDisplay struct {
	name  string
	email string
}

func studentIDsFromOffers(rows []entity.PTSessionOffer) []uint {
	seen := map[uint]struct{}{}
	ids := make([]uint, 0, len(rows))
	for i := range rows {
		id := rows[i].StudentUserID
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

// ============================== Recurring (standing weekly) bookings ==============================

// recurringHorizonDays is how far ahead a standing booking stays materialized
// into real PTSessionOffer rows — comfortably beyond the 14-day lookahead
// generateAvailableSlots ever queries, so other students never see a
// recurring student's weekly slot as "available".
const recurringHorizonDays = 21

func rangesOverlapMinutes(aStart, aEnd, bStart, bEnd int) bool {
	return aStart < bEnd && bStart < aEnd
}

var weekdayLabelsVN = [...]string{"Chủ nhật", "thứ 2", "thứ 3", "thứ 4", "thứ 5", "thứ 6", "thứ 7"}

func weekdayLabelVN(weekday int) string {
	if weekday < 0 || weekday > 6 {
		return ""
	}
	return weekdayLabelsVN[weekday]
}

func toRecurringBookingRes(rb *entity.PTRecurringBooking) gcv1.RecurringBookingRes {
	res := gcv1.RecurringBookingRes{
		ID:               rb.ID,
		UserPTPackageID:  rb.UserPTPackageID,
		TrainerProfileID: rb.TrainerProfileID,
		StudentUserID:    rb.StudentUserID,
		Weekday:          rb.Weekday,
		StartMinute:      rb.StartMinute,
		EndMinute:        rb.EndMinute,
		Status:           rb.Status,
		CreatedAt:        rb.CreatedAt,
		LastGeneratedFor: rb.LastGeneratedFor,
	}
	if rb.UserPTPackage.PTPackage.ID != 0 {
		res.PackageTitle = rb.UserPTPackage.PTPackage.Title
	}
	if rb.UserPTPackage.User.ID != 0 {
		res.StudentEmail = rb.UserPTPackage.User.Email
		res.StudentName = displayNameFromUser(&rb.UserPTPackage.User)
	}
	return res
}

func (s *gymCommerceService) CreateRecurringBooking(ctx context.Context, studentUserID uint, req *gcv1.CreateRecurringBookingReq) (*gcv1.RecurringBookingRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, req.UserPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "không tìm thấy gói tập của người dùng")
	}
	if up.UserID != studentUserID {
		return nil, fmt.Errorf("gói tập không thuộc về người dùng này")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("gói tập không còn hoạt động")
	}
	if err := s.assertSessionsAvailable(ctx, up); err != nil {
		return nil, err
	}
	if req.Weekday < 0 || req.Weekday > 6 {
		return nil, fmt.Errorf("thứ trong tuần không hợp lệ")
	}
	t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "không tìm thấy huấn luyện viên")
	}
	if t.BookingPaused {
		return nil, fmt.Errorf("huấn luyện viên đang tạm dừng nhận lịch")
	}
	dur := s.sessionDuration(t)
	startMinute := req.StartMinute
	endMinute := startMinute + int(dur.Minutes())

	hours, err := s.hoursRepo.ListByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	withinHours := false
	for _, h := range hours {
		if h.Weekday == req.Weekday && startMinute >= h.StartMinute && endMinute <= h.EndMinute {
			withinHours = true
			break
		}
	}
	if !withinHours {
		return nil, fmt.Errorf("khung giờ này nằm ngoài giờ nhận khách của huấn luyện viên")
	}

	existing, err := s.recurringRepo.ListActiveByTrainerAndWeekday(ctx, t.ID, req.Weekday)
	if err != nil {
		return nil, err
	}
	for _, e := range existing {
		if rangesOverlapMinutes(startMinute, endMinute, e.StartMinute, e.EndMinute) {
			return nil, fmt.Errorf("khung giờ cố định này đã có học viên khác đăng ký")
		}
	}

	rb := &entity.PTRecurringBooking{
		UserPTPackageID:  up.ID,
		TrainerProfileID: t.ID,
		StudentUserID:    studentUserID,
		Weekday:          req.Weekday,
		StartMinute:      startMinute,
		EndMinute:        endMinute,
		Status:           entity.RecurringBookingStatusActive,
	}
	if err := s.recurringRepo.Create(ctx, rb); err != nil {
		return nil, err
	}
	rb.UserPTPackage = *up

	queued, _ := s.materializeRecurringBooking(ctx, rb, recurringHorizonDays)
	res := toRecurringBookingRes(rb)
	res.OccurrencesQueued = queued

	// Auto-activated by design (no approval gate — keeps the low-friction
	// self-service booking model), but the trainer is emailed immediately so
	// they can cancel within the first couple of days if it's unwanted.
	trainerEmail, trainerName := s.userEmail(ctx, t.UserID)
	studentName := displayNameFromUser(&up.User)
	if studentName == "" {
		studentName = up.User.Email
	}
	packageTitle := ""
	if up.PTPackage.ID != 0 {
		packageTitle = up.PTPackage.Title
	}
	s.notifyEmail(ctx, "pt_recurring_booking_created", map[string]interface{}{
		"TrainerName":  trainerName,
		"StudentName":  studentName,
		"Weekday":      weekdayLabelVN(rb.Weekday),
		"TimeRange":    fmt.Sprintf("%s–%s", minutesToHHMM(startMinute), minutesToHHMM(endMinute)),
		"PackageTitle": packageTitle,
	}, trainerEmail)

	return &res, nil
}

func minutesToHHMM(m int) string {
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

func (s *gymCommerceService) ListMyRecurringBookingsAsStudent(ctx context.Context, studentUserID uint) (*gcv1.ListRes, error) {
	rows, err := s.recurringRepo.ListByStudentUserID(ctx, studentUserID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.RecurringBookingRes, 0, len(rows))
	for i := range rows {
		out = append(out, toRecurringBookingRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) ListMyRecurringBookingsAsTrainer(ctx context.Context, trainerUserID uint) (*gcv1.ListRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	rows, err := s.recurringRepo.ListByTrainerProfileID(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.RecurringBookingRes, 0, len(rows))
	for i := range rows {
		out = append(out, toRecurringBookingRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) CancelRecurringBooking(ctx context.Context, requesterUserID, id uint) error {
	rb, err := s.recurringRepo.GetByID(ctx, id)
	if err != nil {
		return notFoundOr(err, "không tìm thấy lịch cố định")
	}
	isStudent := rb.StudentUserID == requesterUserID
	isTrainer := false
	if t, terr := s.trainerProfileForUser(ctx, requesterUserID); terr == nil && t.ID == rb.TrainerProfileID {
		isTrainer = true
	}
	if !isStudent && !isTrainer {
		return fmt.Errorf("bạn không có quyền hủy lịch cố định này")
	}
	rb.Status = entity.RecurringBookingStatusCanceled
	if err := s.recurringRepo.Update(ctx, rb); err != nil {
		return err
	}
	// Best-effort cleanup: cancel future, not-yet-completed materialized
	// occurrences — never touch past/completed sessions.
	now := time.Now().UTC()
	offers, err := s.offerRepo.ListByTrainerInRange(ctx, rb.TrainerProfileID, now, now.AddDate(1, 0, 0))
	if err != nil {
		return nil
	}
	for i := range offers {
		o := &offers[i]
		if o.RecurringBookingID == nil || *o.RecurringBookingID != rb.ID {
			continue
		}
		if o.Status != entity.SessionOfferScheduled && o.Status != entity.SessionOfferPending {
			continue
		}
		o.Status = entity.SessionOfferCancelled
		_ = s.offerRepo.Update(ctx, o)
	}
	return nil
}

// materializeRecurringBooking generates any missing dated occurrences between
// the last-generated point and horizonDays from today, skipping (but still
// advancing past) any date that turns out to conflict — a permanently
// blocked week just leaves a gap rather than stalling all future weeks.
func (s *gymCommerceService) materializeRecurringBooking(ctx context.Context, rb *entity.PTRecurringBooking, horizonDays int) (int, error) {
	loc := vnLocation()
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	horizon := today.AddDate(0, 0, horizonDays)

	start := today
	if rb.LastGeneratedFor != nil {
		afterLast := rb.LastGeneratedFor.In(loc).AddDate(0, 0, 1)
		if afterLast.After(start) {
			start = afterLast
		}
	}
	for start.Weekday() != time.Weekday(rb.Weekday) {
		start = start.AddDate(0, 0, 1)
	}

	created := 0
	var furthest *time.Time
	for d := start; !d.After(horizon); d = d.AddDate(0, 0, 7) {
		ok, err := s.materializeOccurrence(ctx, rb, d)
		if err != nil {
			return created, err
		}
		if ok {
			created++
		}
		dCopy := d
		furthest = &dCopy
	}
	if furthest != nil {
		rb.LastGeneratedFor = furthest
		_ = s.recurringRepo.Update(ctx, rb)
	}
	return created, nil
}

// materializeOccurrence tries to create one dated PTSessionOffer for a
// standing booking, guarded by the same advisory-lock transaction as BookSlot
// so it can never collide with a concurrent one-off booking.
func (s *gymCommerceService) materializeOccurrence(ctx context.Context, rb *entity.PTRecurringBooking, dateVN time.Time) (bool, error) {
	loc := vnLocation()
	dayStart := time.Date(dateVN.Year(), dateVN.Month(), dateVN.Day(), 0, 0, 0, 0, loc)
	startsAt := dayStart.Add(time.Duration(rb.StartMinute) * time.Minute).UTC()
	endsAt := dayStart.Add(time.Duration(rb.EndMinute) * time.Minute).UTC()
	if !startsAt.After(time.Now().UTC()) {
		return false, nil
	}
	up, err := s.userPtPkgRepo.GetByID(ctx, rb.UserPTPackageID)
	if err != nil {
		return false, err
	}
	if up.Status != entity.PTPkgStatusActive {
		return false, nil
	}
	if err := s.assertSessionsAvailable(ctx, up); err != nil {
		return false, nil
	}
	dur := endsAt.Sub(startsAt)
	created := false
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(rb.TrainerProfileID)).Error; err != nil {
			return err
		}
		free, err := s.tryReserveSlot(ctx, tx, rb.TrainerProfileID, startsAt, endsAt, dur)
		if err != nil {
			return err
		}
		if !free {
			return nil
		}
		now := time.Now().UTC()
		recurringID := rb.ID
		offer := &entity.PTSessionOffer{
			UserPTPackageID:    up.ID,
			TrainerProfileID:   rb.TrainerProfileID,
			StudentUserID:      rb.StudentUserID,
			StartsAt:           startsAt,
			EndsAt:             &endsAt,
			Note:               "Lịch cố định hàng tuần",
			ProposedByUserID:   rb.StudentUserID,
			Status:             entity.SessionOfferScheduled,
			AcceptedByUserID:   rb.StudentUserID,
			AcceptedAt:         &now,
			BookedViaSlot:      true,
			RecurringBookingID: &recurringID,
		}
		if err := tx.Create(offer).Error; err != nil {
			return err
		}
		created = true
		return nil
	})
	return created, err
}

// MaterializeRecurringBookings is invoked periodically (see main.go) to keep
// every active standing booking generated recurringHorizonDays ahead, and to
// auto-pause any whose package has expired or run out of session credits.
func (s *gymCommerceService) MaterializeRecurringBookings(ctx context.Context, horizonDays int) (int, error) {
	if horizonDays <= 0 {
		horizonDays = recurringHorizonDays
	}
	rows, err := s.recurringRepo.ListActive(ctx)
	if err != nil {
		return 0, err
	}
	total := 0
	for i := range rows {
		rb := &rows[i]
		up, err := s.userPtPkgRepo.GetByID(ctx, rb.UserPTPackageID)
		if err != nil {
			continue
		}
		if up.Status != entity.PTPkgStatusActive {
			rb.Status = entity.RecurringBookingStatusPaused
			_ = s.recurringRepo.Update(ctx, rb)
			continue
		}
		n, err := s.materializeRecurringBooking(ctx, rb, horizonDays)
		if err != nil {
			continue
		}
		total += n
		if err := s.assertSessionsAvailable(ctx, up); err != nil {
			rb.Status = entity.RecurringBookingStatusPaused
			_ = s.recurringRepo.Update(ctx, rb)
		}
	}
	return total, nil
}

func (s *gymCommerceService) userDisplayMap(ctx context.Context, ids []uint) map[uint]userDisplay {
	out := map[uint]userDisplay{}
	if len(ids) == 0 || s.userRepo == nil {
		return out
	}
	users, err := s.userRepo.ListByIDs(ctx, ids)
	if err != nil {
		return out
	}
	for i := range users {
		u := &users[i]
		name := strings.TrimSpace(u.Name)
		if name == "" {
			name = strings.TrimSpace(u.FirstName + " " + u.LastName)
		}
		out[u.ID] = userDisplay{name: name, email: u.Email}
	}
	return out
}
