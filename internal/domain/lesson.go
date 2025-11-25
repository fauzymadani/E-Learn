package domain

import "time"

type Lesson struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CourseID    uint      `json:"course_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	VideoURL    string    `json:"video_url"`
	FileURL     string    `json:"file_url"`
	OrderNumber int       `json:"order_number"`
	Duration    int       `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
