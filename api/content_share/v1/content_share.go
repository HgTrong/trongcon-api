package v1

import "time"

type ShareReq struct {
	ContentType     string `json:"content_type" binding:"required"`
	ContentID       uint   `json:"content_id" binding:"required"`
	RecipientUserID uint   `json:"recipient_user_id" binding:"required"`
}

type RecipientRes struct {
	RecipientUserID uint      `json:"recipient_user_id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	SharedAt        time.Time `json:"shared_at"`
}

type ListRecipientsRes struct {
	Data []RecipientRes `json:"data"`
}

type StudentRes struct {
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type ListStudentsRes struct {
	Data []StudentRes `json:"data"`
}

type SharedItemRes struct {
	ShareID      uint      `json:"share_id"`
	ContentType  string    `json:"content_type"`
	ContentID    uint      `json:"content_id"`
	Title        string    `json:"title"`
	ImageURL     string    `json:"image_url,omitempty"`
	SharedByName string    `json:"shared_by_name"`
	SharedAt     time.Time `json:"shared_at"`
}

type ListSharedRes struct {
	Data []SharedItemRes `json:"data"`
}

type StatusRes struct {
	Status string `json:"status"`
}
