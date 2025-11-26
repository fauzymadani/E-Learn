package domain

import "time"

type NotificationType string

const (
	NotificationTypeEnrollment NotificationType = "enrollment"
	NotificationTypeNewLesson  NotificationType = "new_lesson"
	NotificationTypeCompleted  NotificationType = "completed"
)

type Notification struct {
	ID        uint             `json:"id" gorm:"primaryKey"`
	UserID    uint             `json:"user_id" gorm:"not null;index:idx_notifications_user"`
	Title     string           `json:"title" gorm:"type:varchar(200);not null"`
	Message   string           `json:"message" gorm:"type:text;not null"`
	Type      NotificationType `json:"type" gorm:"type:notification_type;not null"`
	IsRead    bool             `json:"is_read" gorm:"default:false;index:idx_notifications_read"`
	CreatedAt time.Time        `json:"created_at" gorm:"default:CURRENT_TIMESTAMP;index:idx_notifications_created"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Notification) TableName() string {
	return "notifications"
}
