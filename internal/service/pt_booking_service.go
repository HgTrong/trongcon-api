package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	gcv1 "trongcon-api/api/gym_commerce/v1"
	"trongcon-api/internal/entity"
)

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
			return fmt.Errorf("you already have another session booked at this time")
		}
	}
	return nil
}

func (s *gymCommerceService) assertCanPurchasePTPackage(ctx context.Context, userID uint, trainerProfileID uint) error {
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return notFoundOr(err, "trainer not found")
	}
	// BookingPaused means the trainer stopped taking any new session bookings
	// (e.g. on leave) — it applies to existing clients too, unlike AcceptingNewClients.
	if t.BookingPaused {
		return fmt.Errorf("trainer has paused booking activity")
	}
	existing, err := s.userPtPkgRepo.HasActivePackage(ctx, trainerProfileID, userID)
	if err != nil {
		return err
	}
	if existing {
		return nil
	}
	if !t.AcceptingNewClients {
		return fmt.Errorf("trainer is not accepting new clients")
	}
	if t.MaxActiveClients > 0 {
		active, err := s.userPtPkgRepo.CountActiveClients(ctx, t.ID)
		if err != nil {
			return err
		}
		if int(active) >= t.MaxActiveClients {
			return fmt.Errorf("trainer has reached the maximum number of active clients")
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
			return nil, fmt.Errorf("session_duration_min must be between 15 and 240")
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
			return nil, fmt.Errorf("invalid weekday")
		}
		if h.StartMinute < 0 || h.EndMinute > 24*60 || h.EndMinute <= h.StartMinute {
			return nil, fmt.Errorf("invalid time range for weekday %d", h.Weekday)
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
				return nil, fmt.Errorf("overlapping time ranges on weekday %d", day)
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
		return nil, fmt.Errorf("from/to range required")
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
		return nil, fmt.Errorf("ends_at must be after starts_at")
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
		return nil, fmt.Errorf("from/to range required")
	}
	if to.Sub(from) > 14*24*time.Hour {
		return nil, fmt.Errorf("range cannot exceed 14 days")
	}
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "trainer not found")
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

func (s *gymCommerceService) BookSlot(ctx context.Context, studentUserID uint, req *gcv1.BookSlotReq) (*gcv1.SessionOfferRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, req.UserPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	if up.UserID != studentUserID {
		return nil, fmt.Errorf("package does not belong to user")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("package is not active")
	}
	if err := s.assertSessionsAvailable(ctx, up); err != nil {
		return nil, err
	}
	t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "trainer not found")
	}
	if t.BookingPaused {
		return nil, fmt.Errorf("trainer has paused booking activity")
	}
	startsAt := req.StartsAt.UTC().Truncate(time.Minute)
	if !startsAt.After(time.Now().UTC()) {
		return nil, fmt.Errorf("slot must be in the future")
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
		return nil, fmt.Errorf("slot is no longer available")
	}
	if err := s.assertStudentNotBusy(ctx, studentUserID, 0, startsAt, endsAt); err != nil {
		return nil, err
	}

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
	if err := s.offerRepo.Create(ctx, offer); err != nil {
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
		return nil, fmt.Errorf("only the package trainer can mark no-show")
	}
	if offer.Status != entity.SessionOfferScheduled {
		return nil, fmt.Errorf("offer must be scheduled to mark no-show")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("package is not active")
	}
	if up.SessionUsed >= up.SessionTotal {
		return nil, fmt.Errorf("no sessions left")
	}
	now := time.Now().UTC()
	if now.Before(offer.StartsAt) {
		return nil, fmt.Errorf("cannot mark no-show before the session's scheduled start time")
	}
	nextIndex := up.SessionUsed + 1
	note := strings.TrimSpace(offer.Note)
	if note == "" {
		note = "No-show"
	} else {
		note = "No-show — " + note
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
		return nil, notFoundOr(err, "user pt package not found")
	}
	if offer.Status != entity.SessionOfferAwaitingConfirmation {
		return nil, fmt.Errorf("offer is not awaiting confirmation")
	}
	proof := strings.TrimSpace(offer.ProofImageURL)
	if proof == "" {
		return nil, fmt.Errorf("missing proof image")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("package is not active")
	}
	if up.SessionUsed >= up.SessionTotal {
		return nil, fmt.Errorf("no sessions left")
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
		return nil, notFoundOr(err, "trainer not found")
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
		return nil, fmt.Errorf("invalid status filter")
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
		return nil, fmt.Errorf("from/to range required")
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
