package service

import (
	"context"
	"strings"

	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"
)

const (
	ContentTypeArticle  = "article"
	ContentTypeWorkout  = "workout"
	ContentTypeRoutine  = "routine"
	ContentTypeMealPlan = "meal_plan"
)

// PTGrowthTracker records content → profile → booking funnel for trainers.
type PTGrowthTracker interface {
	TrackContentView(ctx context.Context, contentType string, contentID uint, title string, authorUserID, viewerUserID uint)
	TrackContentSave(ctx context.Context, contentType string, contentID uint, title string, authorUserID, viewerUserID uint)
	TrackContentLike(ctx context.Context, contentType string, contentID uint, title string, authorUserID, viewerUserID uint)
	TrackProfileVisit(ctx context.Context, trainerProfileID, viewerUserID uint, contentType string, contentID uint, title string)
	TrackBooking(ctx context.Context, trainerProfileID, studentUserID uint, contentType string, contentID uint, title string)
}

type ptGrowthTracker struct {
	trainerRepo repository.TrainerProfileRepository
	statRepo    repository.PTContentStatRepository
	attrRepo    repository.PTAttributionRepository
}

func NewPTGrowthTracker(
	trainerRepo repository.TrainerProfileRepository,
	statRepo repository.PTContentStatRepository,
	attrRepo repository.PTAttributionRepository,
) PTGrowthTracker {
	return &ptGrowthTracker{trainerRepo: trainerRepo, statRepo: statRepo, attrRepo: attrRepo}
}

func (t *ptGrowthTracker) resolveTrainer(ctx context.Context, authorUserID uint) *entity.TrainerProfile {
	if t == nil || t.trainerRepo == nil || authorUserID == 0 {
		return nil
	}
	tr, err := t.trainerRepo.GetByUserID(ctx, authorUserID)
	if err != nil || tr == nil {
		return nil
	}
	return tr
}

func (t *ptGrowthTracker) TrackContentView(ctx context.Context, contentType string, contentID uint, title string, authorUserID, viewerUserID uint) {
	if t == nil || t.statRepo == nil {
		return
	}
	tr := t.resolveTrainer(ctx, authorUserID)
	if tr == nil {
		return
	}
	ct := strings.TrimSpace(contentType)
	_ = t.statRepo.UpsertTouch(ctx, tr.ID, ct, contentID, strings.TrimSpace(title), 1, 0, 0, 0, 0)
	if viewerUserID > 0 && t.attrRepo != nil {
		_ = t.attrRepo.Upsert(ctx, &entity.PTAttribution{
			UserID:           viewerUserID,
			TrainerProfileID: tr.ID,
			ContentType:      ct,
			ContentID:        contentID,
			Title:            strings.TrimSpace(title),
		})
	}
}

func (t *ptGrowthTracker) TrackContentSave(ctx context.Context, contentType string, contentID uint, title string, authorUserID, viewerUserID uint) {
	if t == nil || t.statRepo == nil {
		return
	}
	tr := t.resolveTrainer(ctx, authorUserID)
	if tr == nil {
		return
	}
	ct := strings.TrimSpace(contentType)
	_ = t.statRepo.UpsertTouch(ctx, tr.ID, ct, contentID, strings.TrimSpace(title), 0, 0, 1, 0, 0)
	if viewerUserID > 0 && t.attrRepo != nil {
		_ = t.attrRepo.Upsert(ctx, &entity.PTAttribution{
			UserID:           viewerUserID,
			TrainerProfileID: tr.ID,
			ContentType:      ct,
			ContentID:        contentID,
			Title:            strings.TrimSpace(title),
		})
	}
}

func (t *ptGrowthTracker) TrackContentLike(ctx context.Context, contentType string, contentID uint, title string, authorUserID, viewerUserID uint) {
	if t == nil || t.statRepo == nil {
		return
	}
	tr := t.resolveTrainer(ctx, authorUserID)
	if tr == nil {
		return
	}
	ct := strings.TrimSpace(contentType)
	_ = t.statRepo.UpsertTouch(ctx, tr.ID, ct, contentID, strings.TrimSpace(title), 0, 1, 0, 0, 0)
}

func (t *ptGrowthTracker) TrackProfileVisit(ctx context.Context, trainerProfileID, viewerUserID uint, contentType string, contentID uint, title string) {
	if t == nil || t.statRepo == nil || trainerProfileID == 0 {
		return
	}
	ct := strings.TrimSpace(contentType)
	if ct == "" || contentID == 0 {
		if viewerUserID > 0 && t.attrRepo != nil {
			if attr, err := t.attrRepo.Get(ctx, viewerUserID, trainerProfileID); err == nil && attr != nil {
				ct, contentID, title = attr.ContentType, attr.ContentID, attr.Title
			}
		}
	}
	if ct == "" || contentID == 0 {
		return
	}
	_ = t.statRepo.UpsertTouch(ctx, trainerProfileID, ct, contentID, strings.TrimSpace(title), 0, 0, 0, 1, 0)
}

func (t *ptGrowthTracker) TrackBooking(ctx context.Context, trainerProfileID, studentUserID uint, contentType string, contentID uint, title string) {
	if t == nil || t.statRepo == nil || trainerProfileID == 0 {
		return
	}
	ct := strings.TrimSpace(contentType)
	if (ct == "" || contentID == 0) && studentUserID > 0 && t.attrRepo != nil {
		if attr, err := t.attrRepo.Get(ctx, studentUserID, trainerProfileID); err == nil && attr != nil {
			ct, contentID, title = attr.ContentType, attr.ContentID, attr.Title
		}
	}
	if ct == "" || contentID == 0 {
		return
	}
	_ = t.statRepo.UpsertTouch(ctx, trainerProfileID, ct, contentID, strings.TrimSpace(title), 0, 0, 0, 0, 1)
}
