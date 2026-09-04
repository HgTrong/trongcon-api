package entity

// ContentShare records a PT privately sharing one piece of authored content
// (workout/routine/meal_plan) with one specific student. Independent of
// IsPublic — grants access without adding the item to the public catalog.
// Access persists even after the underlying UserPTPackage later expires;
// eligibility (must have an active package) is only checked at share time.
type ContentShare struct {
	BaseEntity
	ContentType     string `json:"content_type" gorm:"type:varchar(20);not null;uniqueIndex:idx_content_share_unique,priority:1"`
	ContentID       uint   `json:"content_id" gorm:"not null;uniqueIndex:idx_content_share_unique,priority:2"`
	RecipientUserID uint   `json:"recipient_user_id" gorm:"not null;uniqueIndex:idx_content_share_unique,priority:3;index"`
	Recipient       User   `json:"recipient,omitempty" gorm:"foreignKey:RecipientUserID"`
	SharedByUserID  uint   `json:"shared_by_user_id" gorm:"not null;index"`
	SharedBy        User   `json:"shared_by,omitempty" gorm:"foreignKey:SharedByUserID"`
}

func (ContentShare) TableName() string { return "content_shares" }
