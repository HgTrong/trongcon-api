package entity

const (
	RoleUser  = "user"
	RoleSuper = "super"
	RolePT    = "pt"
)

// Role bảng roles.
type Role struct {
	BaseEntity
	Name string `json:"name" gorm:"type:varchar(64);not null;uniqueIndex"`
	Key  string `json:"key,omitempty" gorm:"type:varchar(64);uniqueIndex"` // legacy DB; mặc định = name
}
