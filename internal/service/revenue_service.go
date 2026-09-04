package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	revenuev1 "trongcon-api/api/revenue/v1"
	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type RevenueService interface {
	// from/to scope the new Selected*/RangeFrom/RangeTo fields; pass nil for
	// both to have them default to "today" (fully backward compatible).
	Overview(ctx context.Context, leaderboardLimit int, from, to *time.Time) (*revenuev1.OverviewRes, error)
	ListPayments(ctx context.Context, page, limit int, source string, from, to *time.Time) (*revenuev1.PaymentsRes, error)
	PTLeaderboard(ctx context.Context, limit int, sortBy string, from, to *time.Time) (*revenuev1.PTLeaderboardRes, error)
	TodayActivity(ctx context.Context) (*revenuev1.TodayActivityRes, error)
}

type revenueService struct {
	db *gorm.DB
}

func NewRevenueService(db *gorm.DB) RevenueService {
	return &revenueService{db: db}
}

func dayBoundsUTC(t time.Time) (time.Time, time.Time) {
	u := t.UTC()
	start := time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
	return start, start.Add(24 * time.Hour)
}

func pctChange(curr, prev float64) float64 {
	if prev == 0 {
		if curr == 0 {
			return 0
		}
		return 100
	}
	return math.Round(((curr-prev)/prev)*1000) / 10
}

func (s *revenueService) sumPremium(ctx context.Context, from, to *time.Time) (gross float64, orders int64, err error) {
	q := s.db.WithContext(ctx).Model(&entity.PaymentHistory{}).
		Where("status = ?", entity.PHStatusSucceeded)
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at < ?", *to)
	}
	type agg struct {
		Gross  float64 `gorm:"column:gross"`
		Orders int64   `gorm:"column:orders"`
	}
	var a agg
	err = q.Select("COALESCE(SUM(amount),0) as gross, COUNT(*) as orders").Scan(&a).Error
	return a.Gross, a.Orders, err
}

func (s *revenueService) sumGym(ctx context.Context, from, to *time.Time) (gross float64, orders int64, err error) {
	q := s.db.WithContext(ctx).Model(&entity.UserGymMembership{}).
		Where("status = ?", entity.GymMemStatusActive)
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at < ?", *to)
	}
	type agg struct {
		Gross  float64 `gorm:"column:gross"`
		Orders int64   `gorm:"column:orders"`
	}
	var a agg
	err = q.Select("COALESCE(SUM(price),0) as gross, COUNT(*) as orders").Scan(&a).Error
	return a.Gross, a.Orders, err
}

func (s *revenueService) sumPT(ctx context.Context, from, to *time.Time) (gross, ptShare, platform float64, orders int64, err error) {
	q := s.db.WithContext(ctx).Model(&entity.PTEarning{})
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at < ?", *to)
	}
	type agg struct {
		Gross    float64 `gorm:"column:gross"`
		PTShare  float64 `gorm:"column:pt_share"`
		Platform float64 `gorm:"column:platform"`
		Orders   int64   `gorm:"column:orders"`
	}
	var a agg
	err = q.Select(`
		COALESCE(SUM(gross_amount),0) as gross,
		COALESCE(SUM(pt_amount),0) as pt_share,
		COALESCE(SUM(gym_amount),0) as platform,
		COUNT(*) as orders
	`).Scan(&a).Error
	return a.Gross, a.PTShare, a.Platform, a.Orders, err
}

func (s *revenueService) snap(ctx context.Context, from, to *time.Time) (revenuev1.MoneySnap, error) {
	prem, premN, err := s.sumPremium(ctx, from, to)
	if err != nil {
		return revenuev1.MoneySnap{}, err
	}
	gym, gymN, err := s.sumGym(ctx, from, to)
	if err != nil {
		return revenuev1.MoneySnap{}, err
	}
	ptGross, ptShare, ptPlatform, ptN, err := s.sumPT(ctx, from, to)
	if err != nil {
		return revenuev1.MoneySnap{}, err
	}
	return revenuev1.MoneySnap{
		Gross:     prem + gym + ptGross,
		Platform:  prem + gym + ptPlatform,
		PTShare:   ptShare,
		Orders:    premN + gymN + ptN,
		Premium:   prem,
		GymPass:   gym,
		PTPackage: ptGross,
	}, nil
}

// resolveRange defaults an optional [from, to) window to "today" (UTC), and
// returns the immediately preceding window of equal length for comparison —
// generalizes the old fixed "today vs yesterday" to any caller-chosen range.
func resolveRange(from, to *time.Time) (rangeFrom, rangeTo, prevFrom, prevTo time.Time) {
	now := time.Now().UTC()
	if from == nil || to == nil {
		rangeFrom, rangeTo = dayBoundsUTC(now)
	} else {
		rangeFrom, rangeTo = from.UTC(), to.UTC()
	}
	length := rangeTo.Sub(rangeFrom)
	if length <= 0 {
		length = 24 * time.Hour
	}
	prevTo = rangeFrom
	prevFrom = rangeFrom.Add(-length)
	return
}

func (s *revenueService) Overview(ctx context.Context, leaderboardLimit int, from, to *time.Time) (*revenuev1.OverviewRes, error) {
	now := time.Now().UTC()
	todayStart, todayEnd := dayBoundsUTC(now)
	ystStart, ystEnd := dayBoundsUTC(now.AddDate(0, 0, -1))
	d7 := now.AddDate(0, 0, -7)
	d30 := now.AddDate(0, 0, -30)

	today, err := s.snap(ctx, &todayStart, &todayEnd)
	if err != nil {
		return nil, err
	}
	yesterday, err := s.snap(ctx, &ystStart, &ystEnd)
	if err != nil {
		return nil, err
	}
	allTime, err := s.snap(ctx, nil, nil)
	if err != nil {
		return nil, err
	}
	last7, err := s.snap(ctx, &d7, nil)
	if err != nil {
		return nil, err
	}
	last30, err := s.snap(ctx, &d30, nil)
	if err != nil {
		return nil, err
	}

	rangeFrom, rangeTo, prevFrom, prevTo := resolveRange(from, to)
	selected, err := s.snap(ctx, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, err
	}
	previousPeriod, err := s.snap(ctx, &prevFrom, &prevTo)
	if err != nil {
		return nil, err
	}
	selectedBySource, err := s.bySource(ctx, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, err
	}
	selectedBoard, _ := s.ptBoard(ctx, leaderboardLimit, "pt_share", &rangeFrom, &rangeTo)

	// Leaderboard is best-effort — don't blank the whole overview if ranking query fails.
	board, _ := s.ptBoard(ctx, leaderboardLimit, "pt_share", nil, nil)
	series, _ := s.dailySeries(ctx, 30)

	bySource, err := s.bySource(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	return &revenuev1.OverviewRes{
		Currency:  "VND",
		Today:     today,
		Yesterday: yesterday,
		ChangePct: revenuev1.ChangePct{
			Gross:    pctChange(today.Gross, yesterday.Gross),
			Platform: pctChange(today.Platform, yesterday.Platform),
			PTShare:  pctChange(today.PTShare, yesterday.PTShare),
			Orders:   pctChange(float64(today.Orders), float64(yesterday.Orders)),
		},
		AllTime:       allTime,
		Last7Days:     last7,
		Last30Days:    last30,
		OwedToPT:      allTime.PTShare,
		PlatformTake:  allTime.Platform,
		BySource:      bySource,
		PTLeaderboard: board,
		DailySeries:   series,
		GeneratedAt:   now,

		RangeFrom: rangeFrom.Format("2006-01-02"),
		RangeTo:   rangeTo.Add(-time.Second).Format("2006-01-02"),
		Selected:  selected,
		PreviousPeriod: previousPeriod,
		SelectedChangePct: revenuev1.ChangePct{
			Gross:    pctChange(selected.Gross, previousPeriod.Gross),
			Platform: pctChange(selected.Platform, previousPeriod.Platform),
			PTShare:  pctChange(selected.PTShare, previousPeriod.PTShare),
			Orders:   pctChange(float64(selected.Orders), float64(previousPeriod.Orders)),
		},
		SelectedBySource:    selectedBySource,
		SelectedLeaderboard: selectedBoard,
	}, nil
}

// bySource breaks a [from, to) window down by revenue source (premium, gym
// pass, PT package) — nil/nil means all-time.
func (s *revenueService) bySource(ctx context.Context, from, to *time.Time) ([]revenuev1.SourceBreakdown, error) {
	snap, err := s.snap(ctx, from, to)
	if err != nil {
		return nil, err
	}
	out := []revenuev1.SourceBreakdown{
		{Source: "premium", Label: "Premium số", Gross: snap.Premium, Platform: snap.Premium},
		{Source: "gym_membership", Label: "Thẻ hội viên", Gross: snap.GymPass, Platform: snap.GymPass},
		{Source: "pt_package", Label: "Gói PT", Gross: snap.PTPackage, Platform: snap.Platform - snap.Premium - snap.GymPass, PTShare: snap.PTShare},
	}
	if _, n, e := s.sumPremium(ctx, from, to); e == nil {
		out[0].Orders = n
	}
	if _, n, e := s.sumGym(ctx, from, to); e == nil {
		out[1].Orders = n
	}
	if _, _, _, n, e := s.sumPT(ctx, from, to); e == nil {
		out[2].Orders = n
	}
	return out, nil
}

func (s *revenueService) dailySeries(ctx context.Context, days int) ([]revenuev1.DailyPoint, error) {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}
	now := time.Now().UTC()
	start, _ := dayBoundsUTC(now.AddDate(0, 0, -(days-1)))
	endExclusive := start.AddDate(0, 0, days)

	type keyed struct {
		Day      time.Time
		Gross    float64
		Platform float64
		PTShare  float64
		Orders   int64
	}
	buckets := make(map[string]*keyed, days)
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		buckets[key] = &keyed{Day: d}
	}

	add := func(day time.Time, gross, platform, ptShare float64, orders int64) {
		key := day.UTC().Format("2006-01-02")
		b, ok := buckets[key]
		if !ok {
			return
		}
		b.Gross += gross
		b.Platform += platform
		b.PTShare += ptShare
		b.Orders += orders
	}

	type premRow struct {
		Day    time.Time `gorm:"column:day"`
		Gross  float64   `gorm:"column:gross"`
		Orders int64     `gorm:"column:orders"`
	}
	var prem []premRow
	if err := s.db.WithContext(ctx).Raw(`
		SELECT (created_at AT TIME ZONE 'UTC')::date AS day,
		       COALESCE(SUM(amount),0) AS gross,
		       COUNT(*) AS orders
		FROM payment_histories
		WHERE deleted_at IS NULL AND status = ? AND created_at >= ? AND created_at < ?
		GROUP BY 1
	`, entity.PHStatusSucceeded, start, endExclusive).Scan(&prem).Error; err != nil {
		return nil, err
	}
	for _, r := range prem {
		add(r.Day, r.Gross, r.Gross, 0, r.Orders)
	}

	type gymRow struct {
		Day    time.Time `gorm:"column:day"`
		Gross  float64   `gorm:"column:gross"`
		Orders int64     `gorm:"column:orders"`
	}
	var gym []gymRow
	if err := s.db.WithContext(ctx).Raw(`
		SELECT (created_at AT TIME ZONE 'UTC')::date AS day,
		       COALESCE(SUM(price),0) AS gross,
		       COUNT(*) AS orders
		FROM user_gym_memberships
		WHERE deleted_at IS NULL AND status = ? AND created_at >= ? AND created_at < ?
		GROUP BY 1
	`, entity.GymMemStatusActive, start, endExclusive).Scan(&gym).Error; err != nil {
		return nil, err
	}
	for _, r := range gym {
		add(r.Day, r.Gross, r.Gross, 0, r.Orders)
	}

	type ptRow struct {
		Day      time.Time `gorm:"column:day"`
		Gross    float64   `gorm:"column:gross"`
		PTShare  float64   `gorm:"column:pt_share"`
		Platform float64   `gorm:"column:platform"`
		Orders   int64     `gorm:"column:orders"`
	}
	var pt []ptRow
	if err := s.db.WithContext(ctx).Raw(`
		SELECT (created_at AT TIME ZONE 'UTC')::date AS day,
		       COALESCE(SUM(gross_amount),0) AS gross,
		       COALESCE(SUM(pt_amount),0) AS pt_share,
		       COALESCE(SUM(gym_amount),0) AS platform,
		       COUNT(*) AS orders
		FROM pt_earnings
		WHERE deleted_at IS NULL AND created_at >= ? AND created_at < ?
		GROUP BY 1
	`, start, endExclusive).Scan(&pt).Error; err != nil {
		return nil, err
	}
	for _, r := range pt {
		add(r.Day, r.Gross, r.Platform, r.PTShare, r.Orders)
	}

	out := make([]revenuev1.DailyPoint, 0, days)
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		b := buckets[key]
		out = append(out, revenuev1.DailyPoint{
			Date:     key,
			Gross:    b.Gross,
			Platform: b.Platform,
			PTShare:  b.PTShare,
			Orders:   b.Orders,
		})
	}
	return out, nil
}

func (s *revenueService) ptBoard(ctx context.Context, limit int, sortBy string, from, to *time.Time) ([]revenuev1.PTLeaderboardRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	orderCol := "pt_share DESC"
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "gross":
		orderCol = "gross DESC"
	case "orders":
		orderCol = "orders DESC"
	case "platform", "platform_share":
		orderCol = "platform_share DESC"
	case "pt_share", "pt", "":
		orderCol = "pt_share DESC"
	}

	type row struct {
		TrainerProfileID uint    `gorm:"column:trainer_profile_id"`
		DisplayName      string  `gorm:"column:display_name"`
		Title            string  `gorm:"column:title"`
		AvatarURL        string  `gorm:"column:avatar_url"`
		Gross            float64 `gorm:"column:gross"`
		PTShare          float64 `gorm:"column:pt_share"`
		PlatformShare    float64 `gorm:"column:platform_share"`
		Orders           int64   `gorm:"column:orders"`
	}
	var rows []row
	where := "e.deleted_at IS NULL"
	args := []interface{}{}
	if from != nil {
		where += " AND e.created_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		where += " AND e.created_at < ?"
		args = append(args, *to)
	}
	args = append(args, limit)
	// trainer_profiles has no avatar_url — use users.profile_picture instead.
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			e.trainer_profile_id,
			COALESCE(t.display_name, '') AS display_name,
			COALESCE(t.title, '') AS title,
			COALESCE(u.profile_picture, '') AS avatar_url,
			COALESCE(SUM(e.gross_amount), 0) AS gross,
			COALESCE(SUM(e.pt_amount), 0) AS pt_share,
			COALESCE(SUM(e.gym_amount), 0) AS platform_share,
			COUNT(*) AS orders
		FROM pt_earnings e
		LEFT JOIN trainer_profiles t ON t.id = e.trainer_profile_id AND t.deleted_at IS NULL
		LEFT JOIN users u ON u.id = t.user_id AND u.deleted_at IS NULL
		WHERE `+where+`
		GROUP BY e.trainer_profile_id, t.display_name, t.title, u.profile_picture
		ORDER BY `+orderCol+`
		LIMIT ?
	`, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]revenuev1.PTLeaderboardRow, 0, len(rows))
	for _, r := range rows {
		name := strings.TrimSpace(r.DisplayName)
		if name == "" {
			name = fmt.Sprintf("PT #%d", r.TrainerProfileID)
		}
		out = append(out, revenuev1.PTLeaderboardRow{
			TrainerProfileID: r.TrainerProfileID,
			DisplayName:      name,
			Title:            r.Title,
			AvatarURL:        r.AvatarURL,
			Gross:            r.Gross,
			PTShare:          r.PTShare,
			PlatformShare:    r.PlatformShare,
			Orders:           r.Orders,
			Currency:         "VND",
		})
	}
	return out, nil
}

func (s *revenueService) PTLeaderboard(ctx context.Context, limit int, sortBy string, from, to *time.Time) (*revenuev1.PTLeaderboardRes, error) {
	rows, err := s.ptBoard(ctx, limit, sortBy, from, to)
	if err != nil {
		return nil, err
	}
	return &revenuev1.PTLeaderboardRes{
		Total:    int64(len(rows)),
		Currency: "VND",
		SortBy:   sortBy,
		Data:     rows,
	}, nil
}

func (s *revenueService) ListPayments(ctx context.Context, page, limit int, source string, from, to *time.Time) (*revenuev1.PaymentsRes, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	source = strings.ToLower(strings.TrimSpace(source))
	all := make([]revenuev1.PaymentRow, 0, 200)
	applyRange := func(q *gorm.DB) *gorm.DB {
		if from != nil {
			q = q.Where("created_at >= ?", *from)
		}
		if to != nil {
			q = q.Where("created_at < ?", *to)
		}
		return q
	}

	if source == "" || source == "premium" {
		var hist []entity.PaymentHistory
		q := applyRange(s.db.WithContext(ctx).Preload("User").
			Where("status = ?", entity.PHStatusSucceeded)).
			Order("id DESC").Limit(200)
		if err := q.Find(&hist).Error; err != nil {
			return nil, err
		}
		for i := range hist {
			h := &hist[i]
			email, name := "", ""
			if h.User.ID != 0 {
				email = h.User.Email
				name = strings.TrimSpace(h.User.Name)
			}
			all = append(all, revenuev1.PaymentRow{
				ID:        fmt.Sprintf("ph-%d", h.ID),
				Source:    "premium",
				Label:     "Premium số",
				UserID:    h.UserID,
				UserEmail: email,
				UserName:  name,
				Amount:    h.Amount,
				Platform:  h.Amount,
				Currency:  defaultCur(h.Currency),
				Provider:  h.PaymentMethod,
				Status:    h.Status,
				CreatedAt: h.CreatedAt,
			})
		}
	}

	if source == "" || source == "gym_membership" || source == "gym" {
		var memb []entity.UserGymMembership
		q := applyRange(s.db.WithContext(ctx).Preload("User").Preload("GymMembershipPlan").
			Where("status = ?", entity.GymMemStatusActive)).
			Order("id DESC").Limit(200)
		if err := q.Find(&memb).Error; err != nil {
			return nil, err
		}
		for i := range memb {
			m := &memb[i]
			email, name := "", ""
			if m.User.ID != 0 {
				email = m.User.Email
				name = strings.TrimSpace(m.User.Name)
			}
			label := "Thẻ hội viên"
			if m.GymMembershipPlan.ID != 0 {
				label = m.GymMembershipPlan.Name
			}
			all = append(all, revenuev1.PaymentRow{
				ID:        fmt.Sprintf("gm-%d", m.ID),
				Source:    "gym_membership",
				Label:     label,
				UserID:    m.UserID,
				UserEmail: email,
				UserName:  name,
				Amount:    m.Price,
				Platform:  m.Price,
				Currency:  defaultCur(m.Currency),
				Provider:  m.PaymentProvider,
				Status:    m.Status,
				CreatedAt: m.CreatedAt,
			})
		}
	}

	if source == "" || source == "pt_package" || source == "pt" {
		var earns []entity.PTEarning
		q := applyRange(s.db.WithContext(ctx).
			Preload("Trainer").
			Preload("UserPTPackage").
			Preload("UserPTPackage.User").
			Preload("UserPTPackage.PTPackage")).
			Order("id DESC").Limit(200)
		if err := q.Find(&earns).Error; err != nil {
			return nil, err
		}
		for i := range earns {
			e := &earns[i]
			email, name, trainer, label := "", "", "", "Gói PT"
			if e.UserPTPackage.ID != 0 {
				if e.UserPTPackage.User.ID != 0 {
					email = e.UserPTPackage.User.Email
					name = strings.TrimSpace(e.UserPTPackage.User.Name)
				}
				if e.UserPTPackage.PTPackage.ID != 0 {
					label = e.UserPTPackage.PTPackage.Title
				}
			}
			if e.Trainer.ID != 0 {
				trainer = e.Trainer.DisplayName
			}
			all = append(all, revenuev1.PaymentRow{
				ID:          fmt.Sprintf("pt-%d", e.ID),
				Source:      "pt_package",
				Label:       label,
				UserID:      e.UserPTPackage.UserID,
				UserEmail:   email,
				UserName:    name,
				TrainerName: trainer,
				Amount:      e.GrossAmount,
				Platform:    e.GymAmount,
				PTShare:     e.PTAmount,
				Currency:    defaultCur(e.Currency),
				Status:      "succeeded",
				CreatedAt:   e.CreatedAt,
			})
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	total := int64(len(all))
	start := (page - 1) * limit
	if start > len(all) {
		start = len(all)
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return &revenuev1.PaymentsRes{
		Total:    total,
		Page:     page,
		Limit:    limit,
		Currency: "VND",
		Data:     all[start:end],
	}, nil
}

// TodayActivity feeds the admin dashboard's "hoạt động hôm nay" panel: how many
// new members, how much money, and exactly who bought what today — without
// admin having to open every list page to piece it together.
func (s *revenueService) TodayActivity(ctx context.Context) (*revenuev1.TodayActivityRes, error) {
	now := time.Now().UTC()
	todayStart, todayEnd := dayBoundsUTC(now)

	var newUsers int64
	if err := s.db.WithContext(ctx).Model(&entity.User{}).
		Where("created_at >= ? AND created_at < ?", todayStart, todayEnd).
		Count(&newUsers).Error; err != nil {
		return nil, err
	}

	var memberships []entity.UserGymMembership
	if err := s.db.WithContext(ctx).Preload("User").Preload("GymMembershipPlan").
		Where("status = ? AND created_at >= ? AND created_at < ?", entity.GymMemStatusActive, todayStart, todayEnd).
		Order("id DESC").
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	gymItems := make([]revenuev1.TodayGymMembershipItem, 0, len(memberships))
	for i := range memberships {
		m := &memberships[i]
		name, email := "", ""
		if m.User.ID != 0 {
			email = m.User.Email
			name = strings.TrimSpace(m.User.Name)
			if name == "" {
				name = email
			}
		}
		plan := "Thẻ hội viên"
		if m.GymMembershipPlan.ID != 0 {
			plan = m.GymMembershipPlan.Name
		}
		gymItems = append(gymItems, revenuev1.TodayGymMembershipItem{
			ID: m.ID, UserName: name, UserEmail: email, PlanName: plan,
			Price: m.Price, Currency: defaultCur(m.Currency), CreatedAt: m.CreatedAt,
		})
	}

	var ptPackages []entity.UserPTPackage
	if err := s.db.WithContext(ctx).Preload("User").Preload("PTPackage").Preload("PTPackage.Trainer").
		Where("status = ? AND created_at >= ? AND created_at < ?", entity.PTPkgStatusActive, todayStart, todayEnd).
		Order("id DESC").
		Find(&ptPackages).Error; err != nil {
		return nil, err
	}
	ptItems := make([]revenuev1.TodayPTPackageItem, 0, len(ptPackages))
	for i := range ptPackages {
		p := &ptPackages[i]
		name, email := "", ""
		if p.User.ID != 0 {
			email = p.User.Email
			name = strings.TrimSpace(p.User.Name)
			if name == "" {
				name = email
			}
		}
		title, trainer, sessions := "Gói PT", "", p.SessionTotal
		if p.PTPackage.ID != 0 {
			title = p.PTPackage.Title
			if p.PTPackage.Trainer.ID != 0 {
				trainer = p.PTPackage.Trainer.DisplayName
			}
		}
		ptItems = append(ptItems, revenuev1.TodayPTPackageItem{
			ID: p.ID, StudentName: name, StudentEmail: email, TrainerName: trainer,
			PackageTitle: title, SessionCount: sessions, Price: p.Price,
			Currency: defaultCur(p.Currency), CreatedAt: p.CreatedAt,
		})
	}

	var subs []entity.UserSubscription
	if err := s.db.WithContext(ctx).Preload("User").Preload("SubscriptionPlan").
		Where("status = ? AND payment_provider != ? AND created_at >= ? AND created_at < ?",
			entity.SubStatusActive, entity.PaymentProviderMembership, todayStart, todayEnd).
		Order("id DESC").
		Find(&subs).Error; err != nil {
		return nil, err
	}
	premiumItems := make([]revenuev1.TodayPremiumItem, 0, len(subs))
	for i := range subs {
		sub := &subs[i]
		name, email := "", ""
		if sub.User.ID != 0 {
			email = sub.User.Email
			name = strings.TrimSpace(sub.User.Name)
			if name == "" {
				name = email
			}
		}
		plan := "Premium"
		if sub.SubscriptionPlan.ID != 0 {
			plan = sub.SubscriptionPlan.PlanName
		}
		premiumItems = append(premiumItems, revenuev1.TodayPremiumItem{
			ID: sub.ID, UserName: name, UserEmail: email, PlanName: plan,
			Price: sub.FinalPrice, Currency: defaultCur(sub.Currency), CreatedAt: sub.CreatedAt,
		})
	}

	// "Money in today" here means gross value of today's new signups
	// (membership + PT package + premium) — not PT commission recognized
	// today via sumPT/snap(), which can lag purchase by weeks (a session only
	// gets taught, and its PTEarning row created, well after the package is
	// bought). Using snap() here made a same-day PT package sale show as "0đ
	// doanh thu" — confusing on a panel whose whole point is "who signed up
	// today, for how much".
	var gymGross, ptGross, premiumGross float64
	for i := range gymItems {
		gymGross += gymItems[i].Price
	}
	for i := range ptItems {
		ptGross += ptItems[i].Price
	}
	for i := range premiumItems {
		premiumGross += premiumItems[i].Price
	}
	revenue := revenuev1.MoneySnap{
		Gross:     gymGross + ptGross + premiumGross,
		Orders:    int64(len(gymItems) + len(ptItems) + len(premiumItems)),
		GymPass:   gymGross,
		PTPackage: ptGross,
		Premium:   premiumGross,
	}

	return &revenuev1.TodayActivityRes{
		Date:              todayStart.Format("2006-01-02"),
		NewUsers:          newUsers,
		Revenue:           revenue,
		NewGymMemberships: gymItems,
		NewPTPackages:     ptItems,
		NewPremium:        premiumItems,
	}, nil
}

func defaultCur(c string) string {
	c = strings.ToUpper(strings.TrimSpace(c))
	if c == "" || c == "USD" {
		return "VND"
	}
	return c
}
