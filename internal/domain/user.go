package domain

import (
	"time"
)

// UserRole represents user role types
type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleTeacher UserRole = "teacher"
	RoleAdmin   UserRole = "admin"
)

// User represents a user entity
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:100;not null;uniqueIndex" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      UserRole  `gorm:"type:varchar(20);not null;default:'student'" json:"role"`
	Avatar    *string   `gorm:"size:255" json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleStudent, RoleTeacher, RoleAdmin:
		return true
	}
	return false
}
