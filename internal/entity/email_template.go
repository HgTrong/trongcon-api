package entity

// EmailTemplate is a reusable SES/SMTP mail body keyed by a stable slug.
// Body + subject use Go text/template syntax, e.g. {{.UserName}}.
type EmailTemplate struct {
	BaseEntity
	Name        string `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	Key         string `json:"key" gorm:"type:varchar(255);not null;uniqueIndex"`
	Subject     string `json:"subject" gorm:"type:varchar(500);not null"`
	Body        string `json:"body" gorm:"type:text;not null"`
	Description string `json:"description" gorm:"type:text"`
	IsActive    bool   `json:"is_active" gorm:"not null;default:true;index"`
}

func (EmailTemplate) TableName() string { return "email_templates" }
