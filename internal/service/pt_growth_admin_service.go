package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	gcv1 "trongcon-api/api/gym_commerce/v1"
	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

func (s *gymCommerceService) AdminContentFunnel(ctx context.Context, trainerProfileID uint) (*gcv1.ContentFunnelRes, error) {
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "trainer not found")
	}
	if s.statRepo != nil {
		_ = s.statRepo.BackfillFromCatalog(ctx, t.ID, t.UserID)
	}
	rows, err := s.statRepo.ListByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	views, likes, saves, profileFromContent, bookings, err := s.statRepo.SumByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	articles, workouts, routines, mealPlans := 0, 0, 0, 0
	items := make([]gcv1.ContentFunnelItemRes, 0, len(rows))
	for _, r := range rows {
		switch r.ContentType {
		case ContentTypeArticle:
			articles++
		case ContentTypeWorkout:
			workouts++
		case ContentTypeRoutine:
			routines++
		case ContentTypeMealPlan:
			mealPlans++
		}
		items = append(items, gcv1.ContentFunnelItemRes{
			ContentType: r.ContentType, ContentID: r.ContentID, Title: r.Title,
			Views: r.Views, Likes: r.Likes, Saves: r.Saves,
			ProfileVisits: r.ProfileVisits, BookingsGenerated: r.BookingsGenerated,
		})
	}
	profileVisits := t.Views
	if profileFromContent > profileVisits {
		profileVisits = profileFromContent
	}
	return &gcv1.ContentFunnelRes{
		TrainerProfileID:    t.ID,
		DisplayName:         t.DisplayName,
		Articles:            articles,
		Workouts:            workouts,
		Routines:            routines,
		MealPlans:           mealPlans,
		TotalViews:          views,
		TotalLikes:          likes,
		TotalSaves:          saves,
		ProfileVisits:       profileVisits,
		BookingsFromContent: bookings,
		Items:               items,
	}, nil
}

func (s *gymCommerceService) AdminTrainerQuality(ctx context.Context, trainerProfileID uint) (*gcv1.TrainerQualityRes, error) {
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "trainer not found")
	}
	avg, reviews, err := s.reviewRepo.AvgRating(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	counts, err := s.offerRepo.CountStatusesByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	completed := counts[entity.SessionOfferCompleted]
	noShows := counts[entity.SessionOfferNoShow]
	cancelled := counts[entity.SessionOfferCancelled] + counts[entity.SessionOfferDeclined]
	total := int64(0)
	for _, n := range counts {
		total += n
	}
	cancelRate := 0.0
	if total > 0 {
		cancelRate = float64(cancelled) / float64(total) * 100
	}
	held := completed + noShows
	noShowRate := 0.0
	if held > 0 {
		noShowRate = float64(noShows) / float64(held) * 100
	}
	active, err := s.userPtPkgRepo.CountActiveClients(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	ever, err := s.userPtPkgRepo.CountClientsEver(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	retention := 0.0
	if ever > 0 {
		retention = float64(active) / float64(ever) * 100
	}
	return &gcv1.TrainerQualityRes{
		TrainerProfileID:  t.ID,
		DisplayName:       t.DisplayName,
		Rating:            avg,
		Reviews:           reviews,
		CompletedSessions: completed,
		CancellationRate:  cancelRate,
		NoShowCount:       noShows,
		NoShowRate:        noShowRate,
		ClientRetention:   retention,
		ActiveClients:     active,
		TotalClientsEver:  ever,
	}, nil
}

func (s *gymCommerceService) AdminTrainerCalendar(ctx context.Context, trainerProfileID uint, from, to time.Time) (*gcv1.TrainerCalendarRes, error) {
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("from/to range required")
	}
	if to.Sub(from) > 35*24*time.Hour {
		return nil, fmt.Errorf("range cannot exceed 35 days")
	}
	t, err := s.trainerRepo.GetByID(ctx, trainerProfileID)
	if err != nil {
		return nil, notFoundOr(err, "trainer not found")
	}
	offers, err := s.offerRepo.ListByTrainerInRange(ctx, t.ID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	blocked, err := s.blockedRepo.ListInRange(ctx, t.ID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	hours, err := s.hoursRepo.ListByTrainer(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	byDayHours := map[int][]entity.PTWorkingHours{}
	for _, h := range hours {
		byDayHours[h.Weekday] = append(byDayHours[h.Weekday], h)
	}
	dur := s.sessionDuration(t)
	loc := vnLocation()
	dayStart := time.Date(from.In(loc).Year(), from.In(loc).Month(), from.In(loc).Day(), 0, 0, 0, 0, loc)
	dayEnd := time.Date(to.In(loc).Year(), to.In(loc).Month(), to.In(loc).Day(), 0, 0, 0, 0, loc)

	days := make([]gcv1.CalendarDayRes, 0)
	for d := dayStart; !d.After(dayEnd); d = d.AddDate(0, 0, 1) {
		windows := byDayHours[int(d.Weekday())]
		open := false
		for _, w := range windows {
			if w.IsActive {
				open = true
				break
			}
		}
		slots := make([]gcv1.CalendarDaySlotRes, 0)
		type window struct{ start, end int }
		var mins []window
		if open {
			for _, w := range windows {
				if w.IsActive {
					mins = append(mins, window{w.StartMinute, w.EndMinute})
				}
			}
		} else {
			mins = []window{{8 * 60, 20 * 60}}
		}
		for _, win := range mins {
			for m := win.start; m+int(dur.Minutes()) <= win.end; m += int(dur.Minutes()) {
				slotStart := d.Add(time.Duration(m) * time.Minute)
				slotEnd := slotStart.Add(dur)
				sUTC, eUTC := slotStart.UTC(), slotEnd.UTC()
				status := "available"
				if !open {
					status = "empty"
				}
				var offerRes *gcv1.SessionOfferRes
				label := ""
				for i := range offers {
					o := &offers[i]
					oEnd := offerEnd(o, dur)
					if rangesOverlap(sUTC, eUTC, o.StartsAt.UTC(), oEnd) {
						status = "booked"
						or := toSessionOfferRes(o)
						offerRes = &or
						label = o.Status
						break
					}
				}
				if status != "booked" {
					for i := range blocked {
						if rangesOverlap(sUTC, eUTC, blocked[i].StartsAt.UTC(), blocked[i].EndsAt.UTC()) {
							status = "blocked"
							label = blocked[i].Reason
							if label == "" {
								label = "closed"
							}
							break
						}
					}
				}
				slots = append(slots, gcv1.CalendarDaySlotRes{
					Hour: m / 60, Minute: m % 60, Status: status, Offer: offerRes, Label: label,
				})
			}
		}
		days = append(days, gcv1.CalendarDayRes{
			Date:  d.Format("2006-01-02"),
			Slots: slots,
		})
	}
	// Enrich booked slots with student display names.
	studentIDs := make([]uint, 0)
	seen := map[uint]struct{}{}
	for di := range days {
		for si := range days[di].Slots {
			o := days[di].Slots[si].Offer
			if o == nil || o.StudentUserID == 0 {
				continue
			}
			if _, ok := seen[o.StudentUserID]; ok {
				continue
			}
			seen[o.StudentUserID] = struct{}{}
			studentIDs = append(studentIDs, o.StudentUserID)
		}
	}
	names := s.userDisplayMap(ctx, studentIDs)
	for di := range days {
		for si := range days[di].Slots {
			o := days[di].Slots[si].Offer
			if o == nil {
				continue
			}
			if u, ok := names[o.StudentUserID]; ok {
				o.StudentName = u.name
				o.StudentEmail = u.email
				if days[di].Slots[si].Label == "" || days[di].Slots[si].Label == o.Status {
					days[di].Slots[si].Label = u.name
				}
			}
		}
	}
	return &gcv1.TrainerCalendarRes{
		TrainerProfileID: t.ID,
		From:             dayStart.Format("2006-01-02"),
		To:               dayEnd.Format("2006-01-02"),
		Days:             days,
	}, nil
}

func (s *gymCommerceService) CreatePTReview(ctx context.Context, studentUserID uint, req *gcv1.CreatePTReviewReq) (*gcv1.PTReviewRes, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, fmt.Errorf("rating must be 1-5")
	}
	offer, up, err := s.getPackageOffer(ctx, studentUserID, req.UserPTPackageID, req.SessionOfferID)
	if err != nil {
		return nil, err
	}
	if up.UserID != studentUserID {
		return nil, fmt.Errorf("only the student can review")
	}
	if offer.Status != entity.SessionOfferCompleted {
		return nil, fmt.Errorf("only completed sessions can be reviewed")
	}
	if existing, err := s.reviewRepo.GetBySessionOfferID(ctx, offer.ID); err == nil && existing != nil {
		return nil, fmt.Errorf("session already reviewed")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	rev := &entity.PTReview{
		TrainerProfileID: offer.TrainerProfileID,
		StudentUserID:    studentUserID,
		UserPTPackageID:  up.ID,
		SessionOfferID:   offer.ID,
		Rating:           req.Rating,
		Comment:          strings.TrimSpace(req.Comment),
	}
	if err := s.reviewRepo.Create(ctx, rev); err != nil {
		return nil, err
	}
	return &gcv1.PTReviewRes{
		ID: rev.ID, TrainerProfileID: rev.TrainerProfileID, StudentUserID: rev.StudentUserID,
		UserPTPackageID: rev.UserPTPackageID, SessionOfferID: rev.SessionOfferID,
		Rating: rev.Rating, Comment: rev.Comment, CreatedAt: rev.CreatedAt,
	}, nil
}

func (s *gymCommerceService) ListPTReviews(ctx context.Context, trainerProfileID uint, page, limit int) (*gcv1.ListRes, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	rows, total, err := s.reviewRepo.ListByTrainer(ctx, trainerProfileID, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	seen := map[uint]struct{}{}
	for _, r := range rows {
		if r.StudentUserID == 0 {
			continue
		}
		if _, ok := seen[r.StudentUserID]; ok {
			continue
		}
		seen[r.StudentUserID] = struct{}{}
		ids = append(ids, r.StudentUserID)
	}
	names := s.userDisplayMap(ctx, ids)
	out := make([]gcv1.PTReviewRes, 0, len(rows))
	for _, r := range rows {
		item := gcv1.PTReviewRes{
			ID: r.ID, TrainerProfileID: r.TrainerProfileID, StudentUserID: r.StudentUserID,
			UserPTPackageID: r.UserPTPackageID, SessionOfferID: r.SessionOfferID,
			Rating: r.Rating, Comment: r.Comment, CreatedAt: r.CreatedAt,
		}
		if u, ok := names[r.StudentUserID]; ok {
			item.StudentName = u.name
			item.StudentEmail = u.email
		}
		out = append(out, item)
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) TouchPTFunnel(ctx context.Context, viewerUserID uint, req *gcv1.TouchPTFunnelReq) error {
	if s.growth == nil {
		return nil
	}
	event := strings.TrimSpace(strings.ToLower(req.Event))
	switch event {
	case "content_view":
		t, err := s.trainerRepo.GetByID(ctx, req.TrainerProfileID)
		if err != nil {
			return notFoundOr(err, "trainer not found")
		}
		s.growth.TrackContentView(ctx, req.ContentType, req.ContentID, req.Title, t.UserID, viewerUserID)
	case "profile_visit":
		s.growth.TrackProfileVisit(ctx, req.TrainerProfileID, viewerUserID, req.ContentType, req.ContentID, req.Title)
	case "booking":
		s.growth.TrackBooking(ctx, req.TrainerProfileID, viewerUserID, req.ContentType, req.ContentID, req.Title)
	case "like":
		t, err := s.trainerRepo.GetByID(ctx, req.TrainerProfileID)
		if err != nil {
			return notFoundOr(err, "trainer not found")
		}
		s.growth.TrackContentLike(ctx, req.ContentType, req.ContentID, req.Title, t.UserID, viewerUserID)
	default:
		return fmt.Errorf("invalid event")
	}
	return nil
}
