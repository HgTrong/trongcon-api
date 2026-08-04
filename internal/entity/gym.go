package entity

// GymBranch is a physical club location shown on marketing + PT assignment.
type GymBranch struct {
	BaseEntity
	Name        string   `json:"name" gorm:"type:varchar(200);not null;index"`
	Slug        string   `json:"slug" gorm:"type:varchar(220);uniqueIndex"`
	Address     string   `json:"address" gorm:"type:varchar(500);not null"`
	District    string   `json:"district" gorm:"type:varchar(120)"`
	City        string   `json:"city" gorm:"type:varchar(120);index"`
	Phone       string   `json:"phone" gorm:"type:varchar(64)"`
	Email       string   `json:"email" gorm:"type:varchar(255)"`
	Hours       string   `json:"hours" gorm:"type:varchar(255)"` // e.g. Mon–Fri 5:30–21:30
	Description string   `json:"description" gorm:"type:text"`
	ImageURL    string   `json:"image_url" gorm:"type:varchar(512)"` // cover / thumbnail
	GalleryURLs string   `json:"-" gorm:"column:gallery_urls;type:text"` // JSON []string
	Latitude    *float64 `json:"latitude" gorm:"type:decimal(10,7)"`  // for Google Maps
	Longitude   *float64 `json:"longitude" gorm:"type:decimal(10,7)"` // for Google Maps
	IsActive    bool     `json:"is_active" gorm:"not null;default:true;index"`
	SortOrder   int      `json:"sort_order" gorm:"not null;default:0;index"`
}

// TrainerProfile is the public PT card for a user with role "pt".
type TrainerProfile struct {
	BaseEntity
	UserID              uint       `json:"user_id" gorm:"not null;uniqueIndex"`
	User                User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	BranchID            *uint      `json:"branch_id" gorm:"index"`
	Branch              *GymBranch `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	DisplayName         string     `json:"display_name" gorm:"type:varchar(200);not null"`
	Title               string     `json:"title" gorm:"type:varchar(200)"` // e.g. Strength coach
	Bio                 string     `json:"bio" gorm:"type:text"`
	Specialties         string     `json:"specialties" gorm:"type:varchar(500)"` // comma-separated
	Certifications      string     `json:"certifications" gorm:"type:varchar(500)"`
	YearsExperience     int        `json:"years_experience" gorm:"not null;default:0"`
	IsPublic            bool       `json:"is_public" gorm:"not null;default:true;index"`
	Views               int64      `json:"views" gorm:"not null;default:0"`
	// Booking / schedule controls (slot-based PT booking).
	SessionDurationMin  int        `json:"session_duration_min" gorm:"not null;default:60"`
	AcceptingNewClients bool       `json:"accepting_new_clients" gorm:"not null;default:true;index"`
	BookingPaused       bool       `json:"booking_paused" gorm:"not null;default:false;index"`
	MaxActiveClients    int        `json:"max_active_clients" gorm:"not null;default:10"`
}
