package domain

import "time"

// StudentDashboard represents student dashboard data
type StudentDashboard struct {
	EnrolledCourses     []EnrolledCourseCard    `json:"enrolled_courses"`
	LearningProgress    []CourseProgressSummary `json:"learning_progress"`
	RecentNotifications []Notification          `json:"recent_notifications"`
	Stats               StudentStats            `json:"stats"`
}

type EnrolledCourseCard struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	Thumbnail        string    `json:"thumbnail"`
	TeacherName      string    `json:"teacher_name"`
	TotalLessons     int       `json:"total_lessons"`
	CompletedLessons int       `json:"completed_lessons"`
	ProgressPercent  float64   `json:"progress_percent"`
	EnrolledAt       time.Time `json:"enrolled_at"`
	Status           string    `json:"status"`
}

// CourseProgressSummary represents learning progress summary (different from domain.CourseProgress)
type CourseProgressSummary struct {
	CourseID         uint       `json:"course_id"`
	CourseTitle      string     `json:"course_title"`
	TotalLessons     int        `json:"total_lessons"`
	CompletedLessons int        `json:"completed_lessons"`
	ProgressPercent  float64    `json:"progress_percent"`
	LastAccessedAt   *time.Time `json:"last_accessed_at"`
}

type StudentStats struct {
	TotalEnrolled         int `json:"total_enrolled"`
	CoursesCompleted      int `json:"courses_completed"`
	InProgress            int `json:"in_progress"`
	TotalLessonsCompleted int `json:"total_lessons_completed"`
}

// TeacherDashboard represents teacher dashboard data
type TeacherDashboard struct {
	MyCourses         []TeacherCourse    `json:"my_courses"`
	TotalStudents     int                `json:"total_students"`
	RecentEnrollments []RecentEnrollment `json:"recent_enrollments"`
	Stats             TeacherStats       `json:"stats"`
}

type TeacherCourse struct {
	ID                int64     `json:"id"`
	Title             string    `json:"title"`
	Thumbnail         string    `json:"thumbnail"`
	IsPublished       bool      `json:"is_published"`
	TotalLessons      int       `json:"total_lessons"`
	TotalStudents     int       `json:"total_students"`
	ActiveStudents    int       `json:"active_students"`
	CompletedStudents int       `json:"completed_students"`
	CreatedAt         time.Time `json:"created_at"`
}

type RecentEnrollment struct {
	ID           uint      `json:"id"`
	StudentName  string    `json:"student_name"`
	StudentEmail string    `json:"student_email"`
	CourseName   string    `json:"course_name"`
	EnrolledAt   time.Time `json:"enrolled_at"`
	Status       string    `json:"status"`
}

type TeacherStats struct {
	TotalCourses     int `json:"total_courses"`
	PublishedCourses int `json:"published_courses"`
	TotalStudents    int `json:"total_students"`
	TotalLessons     int `json:"total_lessons"`
}

// AdminDashboard represents admin dashboard data
type AdminDashboard struct {
	TotalUsers       int              `json:"total_users"`
	TotalCourses     int              `json:"total_courses"`
	TotalEnrollments int              `json:"total_enrollments"`
	Statistics       AdminStatistics  `json:"statistics"`
	RecentActivities []RecentActivity `json:"recent_activities"`
}

type AdminStatistics struct {
	UsersByRole         map[string]int `json:"users_by_role"`
	CoursesByStatus     CourseStatus   `json:"courses_by_status"`
	EnrollmentsByStatus map[string]int `json:"enrollments_by_status"`
	GrowthStats         GrowthStats    `json:"growth_stats"`
}

type CourseStatus struct {
	Published   int `json:"published"`
	Unpublished int `json:"unpublished"`
	Total       int `json:"total"`
}

type GrowthStats struct {
	NewUsersThisMonth       int `json:"new_users_this_month"`
	NewCoursesThisMonth     int `json:"new_courses_this_month"`
	NewEnrollmentsThisMonth int `json:"new_enrollments_this_month"`
}

type RecentActivity struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"` // "user_registered", "course_created", "enrollment", "course_completed"
	Description string    `json:"description"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}
