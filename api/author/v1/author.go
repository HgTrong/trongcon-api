package v1

// AuthorRes describes the PT author of a public catalog item.
type AuthorRes struct {
	TrainerID   uint   `json:"trainer_id"`
	UserID      uint   `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Title       string `json:"title,omitempty"`
}
