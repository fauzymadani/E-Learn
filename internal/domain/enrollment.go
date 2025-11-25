package domain

import "time"

// EnrollmentStatus represents the status of an enrollment
type EnrollmentStatus string

const (
	EnrollmentStatusActive    EnrollmentStatus = "active"
	EnrollmentStatusCompleted EnrollmentStatus = "completed"
	EnrollmentStatusDropped   EnrollmentStatus = "dropped"
)

type Enrollment struct {
	ID          uint             `json:"id" gorm:"primaryKey"`
	UserID      uint             `json:"user_id" gorm:"not null"`
	CourseID    uint             `json:"course_id" gorm:"not null"`
	EnrolledAt  time.Time        `json:"enrolled_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	Status      EnrollmentStatus `json:"status" gorm:"type:enrollment_status;not null;default:'active'"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`

	// Relations
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Course Course `json:"course,omitempty" gorm:"foreignKey:CourseID"`
}

func (Enrollment) TableName() string {
	return "enrollments"
}
