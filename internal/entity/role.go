package entity

// Tên role cố định; có thể mở rộng thêm bản ghi trong bảng roles sau.
const (
	RoleUser  = "user"
	RoleSuper = "super"
)

// Role bảng roles.
type Role struct {
	BaseEntity
	Name string `json:"name" gorm:"type:varchar(64);not null;uniqueIndex"`
}
