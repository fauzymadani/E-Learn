package domain

import "time"

type Progress struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"not null"`
	LessonID    uint       `json:"lesson_id" gorm:"not null"`
	IsCompleted bool       `json:"is_completed" gorm:"not null;default:false"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Relations
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Lesson Lesson `json:"lesson,omitempty" gorm:"foreignKey:LessonID"`
}

func (Progress) TableName() string {
	return "progress"
}

// CourseProgress represents overall progress for a course
type CourseProgress struct {
	CourseID           uint    `json:"course_id"`
	TotalLessons       int     `json:"total_lessons"`
	CompletedLessons   int     `json:"completed_lessons"`
	ProgressPercentage float64 `json:"progress_percentage"`
	IsCompleted        bool    `json:"is_completed"`
}

func TimeNow() time.Time {
	return time.Now()
}
