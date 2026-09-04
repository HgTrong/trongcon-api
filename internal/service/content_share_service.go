package service

import (
	"context"
	"errors"
	"fmt"

	contentsharev1 "trongcon-api/api/content_share/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrContentShareInvalidType     = errors.New("invalid content type")
	ErrContentShareForbidden       = errors.New("content not owned by trainer")
	ErrContentShareNotTrainer      = errors.New("user is not a trainer")
	ErrContentShareNotActiveClient = errors.New("student does not have an active package with this trainer")
)

type ContentShareService interface {
	Share(ctx context.Context, trainerUserID uint, contentType string, contentID, recipientUserID uint) error
	Unshare(ctx context.Context, trainerUserID uint, contentType string, contentID, recipientUserID uint) error
	ListRecipients(ctx context.Context, trainerUserID uint, contentType string, contentID uint) (*contentsharev1.ListRecipientsRes, error)
	ListShareableStudents(ctx context.Context, trainerUserID uint) (*contentsharev1.ListStudentsRes, error)
	ListSharedWithMe(ctx context.Context, userID uint, contentType string) (*contentsharev1.ListSharedRes, error)
	IsSharedWithUser(ctx context.Context, contentType string, contentID, userID uint) (bool, error)
}

type contentShareService struct {
	shareRepo     repository.ContentShareRepository
	workoutRepo   repository.WorkoutRepository
	routineRepo   repository.RoutineRepository
	mealPlanRepo  repository.MealPlanRepository
	trainerRepo   repository.TrainerProfileRepository
	ptPackageRepo repository.UserPTPackageRepository
	userRepo      repository.UserRepository
	chatRepo      repository.PTPackageChatRepository
}

func NewContentShareService(
	shareRepo repository.ContentShareRepository,
	workoutRepo repository.WorkoutRepository,
	routineRepo repository.RoutineRepository,
	mealPlanRepo repository.MealPlanRepository,
	trainerRepo repository.TrainerProfileRepository,
	ptPackageRepo repository.UserPTPackageRepository,
	userRepo repository.UserRepository,
	chatRepo repository.PTPackageChatRepository,
) ContentShareService {
	return &contentShareService{
		shareRepo: shareRepo, workoutRepo: workoutRepo, routineRepo: routineRepo, mealPlanRepo: mealPlanRepo,
		trainerRepo: trainerRepo, ptPackageRepo: ptPackageRepo, userRepo: userRepo, chatRepo: chatRepo,
	}
}

// assertOwnedContent confirms trainerUserID authored contentID and returns its title.
func (s *contentShareService) assertOwnedContent(ctx context.Context, trainerUserID uint, contentType string, contentID uint) (string, error) {
	switch contentType {
	case ContentTypeWorkout:
		w, err := s.workoutRepo.GetByID(ctx, contentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", ErrWorkoutNotFound
			}
			return "", err
		}
		if workoutAuthorID(w) != trainerUserID {
			return "", ErrContentShareForbidden
		}
		return w.Title, nil
	case ContentTypeRoutine:
		rt, err := s.routineRepo.GetByID(ctx, contentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", ErrRoutineNotFound
			}
			return "", err
		}
		if rt.UserID != trainerUserID {
			return "", ErrContentShareForbidden
		}
		return rt.Title, nil
	case ContentTypeMealPlan:
		mp, err := s.mealPlanRepo.GetByID(ctx, contentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", ErrMealPlanNotFound
			}
			return "", err
		}
		if mp.UserID != trainerUserID {
			return "", ErrContentShareForbidden
		}
		return mp.Title, nil
	default:
		return "", ErrContentShareInvalidType
	}
}

func (s *contentShareService) Share(ctx context.Context, trainerUserID uint, contentType string, contentID, recipientUserID uint) error {
	title, err := s.assertOwnedContent(ctx, trainerUserID, contentType, contentID)
	if err != nil {
		return err
	}
	tp, err := s.trainerRepo.GetByUserID(ctx, trainerUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContentShareNotTrainer
		}
		return err
	}
	active, err := s.ptPackageRepo.HasActivePackage(ctx, tp.ID, recipientUserID)
	if err != nil {
		return err
	}
	if !active {
		return ErrContentShareNotActiveClient
	}
	if err := s.shareRepo.Create(ctx, &entity.ContentShare{
		ContentType:     contentType,
		ContentID:       contentID,
		RecipientUserID: recipientUserID,
		SharedByUserID:  trainerUserID,
	}); err != nil {
		return err
	}
	s.postShareChatMessage(ctx, tp.ID, trainerUserID, recipientUserID, contentType, contentID, title)
	return nil
}

// postShareChatMessage drops a clickable card into the PT↔student chat so the
// student notices the new content without having to check /shared-with-me —
// best-effort: a failure here shouldn't undo an otherwise-successful share.
func (s *contentShareService) postShareChatMessage(ctx context.Context, trainerProfileID, trainerUserID, recipientUserID uint, contentType string, contentID uint, title string) {
	pkgs, _, err := s.ptPackageRepo.ListAdmin(ctx, 0, 1, entity.PTPkgStatusActive, trainerProfileID, recipientUserID, nil, nil)
	if err != nil || len(pkgs) == 0 {
		return
	}
	id := contentID
	_ = s.chatRepo.Create(ctx, &entity.PTPackageChatMessage{
		UserPTPackageID:   pkgs[0].ID,
		SenderUserID:      trainerUserID,
		Body:              fmt.Sprintf("Đã gửi cho bạn: %s", title),
		MessageType:       entity.ChatMsgTypeContentShare,
		SharedContentType: contentType,
		SharedContentID:   &id,
	})
}

func (s *contentShareService) Unshare(ctx context.Context, trainerUserID uint, contentType string, contentID, recipientUserID uint) error {
	if _, err := s.assertOwnedContent(ctx, trainerUserID, contentType, contentID); err != nil {
		return err
	}
	return s.shareRepo.Delete(ctx, contentType, contentID, recipientUserID)
}

func (s *contentShareService) ListRecipients(ctx context.Context, trainerUserID uint, contentType string, contentID uint) (*contentsharev1.ListRecipientsRes, error) {
	if _, err := s.assertOwnedContent(ctx, trainerUserID, contentType, contentID); err != nil {
		return nil, err
	}
	rows, err := s.shareRepo.ListRecipients(ctx, contentType, contentID)
	if err != nil {
		return nil, err
	}
	data := make([]contentsharev1.RecipientRes, 0, len(rows))
	for _, row := range rows {
		data = append(data, contentsharev1.RecipientRes{
			RecipientUserID: row.RecipientUserID,
			Name:            displayNameFromUser(&row.Recipient),
			Email:           row.Recipient.Email,
			SharedAt:        row.CreatedAt,
		})
	}
	return &contentsharev1.ListRecipientsRes{Data: data}, nil
}

// ListShareableStudents lists the trainer's currently-active clients — the
// only valid share recipients (per-session package status, not per item).
func (s *contentShareService) ListShareableStudents(ctx context.Context, trainerUserID uint) (*contentsharev1.ListStudentsRes, error) {
	tp, err := s.trainerRepo.GetByUserID(ctx, trainerUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentShareNotTrainer
		}
		return nil, err
	}
	rows, _, err := s.ptPackageRepo.ListByTrainerProfileID(ctx, tp.ID, 0, 200, entity.PTPkgStatusActive)
	if err != nil {
		return nil, err
	}
	seen := make(map[uint]bool, len(rows))
	data := make([]contentsharev1.StudentRes, 0, len(rows))
	for _, row := range rows {
		if seen[row.UserID] {
			continue
		}
		seen[row.UserID] = true
		data = append(data, contentsharev1.StudentRes{
			UserID: row.UserID,
			Name:   displayNameFromUser(&row.User),
			Email:  row.User.Email,
		})
	}
	return &contentsharev1.ListStudentsRes{Data: data}, nil
}

func (s *contentShareService) ListSharedWithMe(ctx context.Context, userID uint, contentType string) (*contentsharev1.ListSharedRes, error) {
	rows, err := s.shareRepo.ListSharedWithUser(ctx, userID, contentType)
	if err != nil {
		return nil, err
	}
	data := make([]contentsharev1.SharedItemRes, 0, len(rows))
	for _, row := range rows {
		title, imageURL, ok := s.contentSummary(ctx, row.ContentType, row.ContentID)
		if !ok {
			continue
		}
		data = append(data, contentsharev1.SharedItemRes{
			ShareID:      row.ID,
			ContentType:  row.ContentType,
			ContentID:    row.ContentID,
			Title:        title,
			ImageURL:     imageURL,
			SharedByName: displayNameFromUser(&row.SharedBy),
			SharedAt:     row.CreatedAt,
		})
	}
	return &contentsharev1.ListSharedRes{Data: data}, nil
}

func (s *contentShareService) contentSummary(ctx context.Context, contentType string, contentID uint) (title, imageURL string, ok bool) {
	switch contentType {
	case ContentTypeWorkout:
		w, err := s.workoutRepo.GetByID(ctx, contentID)
		if err != nil {
			return "", "", false
		}
		return w.Title, w.ImageURL, true
	case ContentTypeRoutine:
		rt, err := s.routineRepo.GetByID(ctx, contentID)
		if err != nil {
			return "", "", false
		}
		return rt.Title, rt.ImageURL, true
	case ContentTypeMealPlan:
		mp, err := s.mealPlanRepo.GetByID(ctx, contentID)
		if err != nil {
			return "", "", false
		}
		return mp.Title, "", true
	default:
		return "", "", false
	}
}

func (s *contentShareService) IsSharedWithUser(ctx context.Context, contentType string, contentID, userID uint) (bool, error) {
	return s.shareRepo.IsSharedWithUser(ctx, contentType, contentID, userID)
}
