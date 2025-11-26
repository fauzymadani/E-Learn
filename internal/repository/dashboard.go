package repository

import (
	"context"
	"elearning/internal/domain"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetStudentEnrolledCourses(ctx context.Context, userID uint) ([]domain.EnrolledCourseCard, error)
	GetStudentStats(ctx context.Context, userID uint) (*domain.StudentStats, error)

	GetTeacherCourses(ctx context.Context, teacherID int64) ([]domain.TeacherCourse, error)
	GetTeacherStats(ctx context.Context, teacherID int64) (*domain.TeacherStats, error)
	GetRecentEnrollments(ctx context.Context, teacherID int64, limit int) ([]domain.RecentEnrollment, error)

	GetAdminStats(ctx context.Context) (*domain.AdminStatistics, error)
	GetRecentActivities(ctx context.Context, limit int) ([]domain.RecentActivity, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

// ===== STUDENT DASHBOARD =====

func (r *dashboardRepository) GetStudentEnrolledCourses(ctx context.Context, userID uint) ([]domain.EnrolledCourseCard, error) {
	var enrollments []domain.Enrollment

	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("User").
		Where("user_id = ? AND status = ?", userID, domain.EnrollmentStatusActive).
		Order("enrolled_at DESC").
		Find(&enrollments).Error

	if err != nil {
		return nil, err
	}

	if len(enrollments) == 0 {
		return []domain.EnrolledCourseCard{}, nil
	}

	cards := make([]domain.EnrolledCourseCard, 0, len(enrollments))

	for _, enrollment := range enrollments {
		// Get teacher info
		var teacher domain.User
		r.db.WithContext(ctx).First(&teacher, enrollment.Course.TeacherID)

		// Count total lessons
		var totalLessons int64
		r.db.WithContext(ctx).
			Model(&domain.Lesson{}).
			Where("course_id = ?", enrollment.Course.ID).
			Count(&totalLessons)

		// Count completed lessons
		var completedCount int64
		if totalLessons > 0 {
			r.db.WithContext(ctx).
				Table("progress").
				Joins("JOIN lessons ON progress.lesson_id = lessons.id").
				Where("lessons.course_id = ? AND progress.user_id = ? AND progress.is_completed = ?",
					enrollment.Course.ID, userID, true).
				Count(&completedCount)
		}

		var progressPercent float64
		if totalLessons > 0 {
			progressPercent = float64(completedCount) / float64(totalLessons) * 100
		}

		cards = append(cards, domain.EnrolledCourseCard{
			ID:               uint(enrollment.Course.ID),
			Title:            enrollment.Course.Title,
			Thumbnail:        enrollment.Course.Thumbnail,
			TeacherName:      teacher.Name,
			TotalLessons:     int(totalLessons),
			CompletedLessons: int(completedCount),
			ProgressPercent:  progressPercent,
			EnrolledAt:       enrollment.EnrolledAt,
			Status:           string(enrollment.Status),
		})
	}

	return cards, nil
}

func (r *dashboardRepository) GetStudentStats(ctx context.Context, userID uint) (*domain.StudentStats, error) {
	stats := &domain.StudentStats{}

	// Total enrolled (active)
	var totalEnrolled int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Enrollment{}).
		Where("user_id = ? AND status = ?", userID, domain.EnrollmentStatusActive).
		Count(&totalEnrolled).Error; err != nil {
		return nil, err
	}
	stats.TotalEnrolled = int(totalEnrolled)

	// Total completed courses
	var coursesCompleted int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Enrollment{}).
		Where("user_id = ? AND status = ?", userID, domain.EnrollmentStatusCompleted).
		Count(&coursesCompleted).Error; err != nil {
		return nil, err
	}
	stats.CoursesCompleted = int(coursesCompleted)

	// In progress
	stats.InProgress = stats.TotalEnrolled

	// Total lessons completed
	var totalLessonsCompleted int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Progress{}).
		Where("user_id = ? AND is_completed = ?", userID, true).
		Count(&totalLessonsCompleted).Error; err != nil {
		return nil, err
	}
	stats.TotalLessonsCompleted = int(totalLessonsCompleted)

	return stats, nil
}

// ===== TEACHER DASHBOARD =====

func (r *dashboardRepository) GetTeacherCourses(ctx context.Context, teacherID int64) ([]domain.TeacherCourse, error) {
	var courses []domain.Course

	err := r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").
		Find(&courses).Error

	if err != nil {
		return nil, err
	}

	if len(courses) == 0 {
		return []domain.TeacherCourse{}, nil
	}

	teacherCourses := make([]domain.TeacherCourse, 0, len(courses))

	for _, course := range courses {
		// Count lessons
		var totalLessons int64
		r.db.WithContext(ctx).
			Model(&domain.Lesson{}).
			Where("course_id = ?", course.ID).
			Count(&totalLessons)

		// Count enrollments by status
		var totalStudents, activeStudents, completedStudents int64

		r.db.WithContext(ctx).
			Model(&domain.Enrollment{}).
			Where("course_id = ?", course.ID).
			Count(&totalStudents)

		r.db.WithContext(ctx).
			Model(&domain.Enrollment{}).
			Where("course_id = ? AND status = ?", course.ID, domain.EnrollmentStatusActive).
			Count(&activeStudents)

		r.db.WithContext(ctx).
			Model(&domain.Enrollment{}).
			Where("course_id = ? AND status = ?", course.ID, domain.EnrollmentStatusCompleted).
			Count(&completedStudents)

		teacherCourses = append(teacherCourses, domain.TeacherCourse{
			ID:                course.ID,
			Title:             course.Title,
			Thumbnail:         course.Thumbnail,
			IsPublished:       course.IsPublished,
			TotalLessons:      int(totalLessons),
			TotalStudents:     int(totalStudents),
			ActiveStudents:    int(activeStudents),
			CompletedStudents: int(completedStudents),
			CreatedAt:         course.CreatedAt,
		})
	}

	return teacherCourses, nil
}

func (r *dashboardRepository) GetTeacherStats(ctx context.Context, teacherID int64) (*domain.TeacherStats, error) {
	stats := &domain.TeacherStats{}

	// Total courses
	var totalCourses int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Course{}).
		Where("teacher_id = ?", teacherID).
		Count(&totalCourses).Error; err != nil {
		return nil, err
	}
	stats.TotalCourses = int(totalCourses)

	// Published courses
	var publishedCourses int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Course{}).
		Where("teacher_id = ? AND is_published = ?", teacherID, true).
		Count(&publishedCourses).Error; err != nil {
		return nil, err
	}
	stats.PublishedCourses = int(publishedCourses)

	// Get course IDs
	var courseIDs []int64
	r.db.WithContext(ctx).
		Model(&domain.Course{}).
		Where("teacher_id = ?", teacherID).
		Pluck("id", &courseIDs)

	if len(courseIDs) > 0 {
		// Total unique students
		var totalStudents int64
		r.db.WithContext(ctx).
			Model(&domain.Enrollment{}).
			Where("course_id IN ?", courseIDs).
			Distinct("user_id").
			Count(&totalStudents)
		stats.TotalStudents = int(totalStudents)

		// Total lessons
		var totalLessons int64
		r.db.WithContext(ctx).
			Model(&domain.Lesson{}).
			Where("course_id IN ?", courseIDs).
			Count(&totalLessons)
		stats.TotalLessons = int(totalLessons)
	}

	return stats, nil
}

func (r *dashboardRepository) GetRecentEnrollments(ctx context.Context, teacherID int64, limit int) ([]domain.RecentEnrollment, error) {
	// Get teacher's course IDs
	var courseIDs []int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Course{}).
		Where("teacher_id = ?", teacherID).
		Pluck("id", &courseIDs).Error; err != nil {
		return nil, err
	}

	if len(courseIDs) == 0 {
		return []domain.RecentEnrollment{}, nil
	}

	var enrollments []domain.Enrollment

	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Course").
		Where("course_id IN ?", courseIDs).
		Order("enrolled_at DESC").
		Limit(limit).
		Find(&enrollments).Error

	if err != nil {
		return nil, err
	}

	recentEnrollments := make([]domain.RecentEnrollment, 0, len(enrollments))
	for _, e := range enrollments {
		recentEnrollments = append(recentEnrollments, domain.RecentEnrollment{
			ID:           e.ID,
			StudentName:  e.User.Name,
			StudentEmail: e.User.Email,
			CourseName:   e.Course.Title,
			EnrolledAt:   e.EnrolledAt,
			Status:       string(e.Status),
		})
	}

	return recentEnrollments, nil
}

// ===== ADMIN DASHBOARD =====

func (r *dashboardRepository) GetAdminStats(ctx context.Context) (*domain.AdminStatistics, error) {
	stats := &domain.AdminStatistics{
		UsersByRole:         make(map[string]int),
		EnrollmentsByStatus: make(map[string]int),
	}

	// Users by role
	var userRoles []struct {
		Role  string
		Count int64
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Select("role, COUNT(*) as count").
		Group("role").
		Scan(&userRoles).Error; err != nil {
		return nil, err
	}

	for _, ur := range userRoles {
		stats.UsersByRole[ur.Role] = int(ur.Count)
	}

	// Courses by status
	var published, unpublished, total int64

	r.db.WithContext(ctx).Model(&domain.Course{}).Where("is_published = ?", true).Count(&published)
	r.db.WithContext(ctx).Model(&domain.Course{}).Where("is_published = ?", false).Count(&unpublished)
	r.db.WithContext(ctx).Model(&domain.Course{}).Count(&total)

	stats.CoursesByStatus = domain.CourseStatus{
		Published:   int(published),
		Unpublished: int(unpublished),
		Total:       int(total),
	}

	// Enrollments by status
	var enrollmentStatuses []struct {
		Status string
		Count  int64
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.Enrollment{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&enrollmentStatuses).Error; err != nil {
		return nil, err
	}

	for _, es := range enrollmentStatuses {
		stats.EnrollmentsByStatus[es.Status] = int(es.Count)
	}

	// Growth stats (this month)
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())

	var newUsers, newCourses, newEnrollments int64

	r.db.WithContext(ctx).Model(&domain.User{}).Where("created_at >= ?", startOfMonth).Count(&newUsers)
	r.db.WithContext(ctx).Model(&domain.Course{}).Where("created_at >= ?", startOfMonth).Count(&newCourses)
	r.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("enrolled_at >= ?", startOfMonth).Count(&newEnrollments)

	stats.GrowthStats = domain.GrowthStats{
		NewUsersThisMonth:       int(newUsers),
		NewCoursesThisMonth:     int(newCourses),
		NewEnrollmentsThisMonth: int(newEnrollments),
	}

	return stats, nil
}

func (r *dashboardRepository) GetRecentActivities(ctx context.Context, limit int) ([]domain.RecentActivity, error) {
	activities := make([]domain.RecentActivity, 0, limit)

	// Get recent users
	var users []domain.User
	r.db.WithContext(ctx).
		Select("id, name, created_at").
		Order("created_at DESC").
		Limit(limit / 3).
		Find(&users)

	for _, u := range users {
		activities = append(activities, domain.RecentActivity{
			ID:          u.ID,
			Type:        "user_registered",
			Description: "New user registered",
			UserName:    u.Name,
			CreatedAt:   u.CreatedAt,
		})
	}

	// Get recent courses
	var courses []domain.Course
	r.db.WithContext(ctx).
		Select("id, title, created_at, teacher_id").
		Order("created_at DESC").
		Limit(limit / 3).
		Find(&courses)

	for _, c := range courses {
		var teacher domain.User
		r.db.WithContext(ctx).First(&teacher, c.TeacherID)

		activities = append(activities, domain.RecentActivity{
			ID:          uint(c.ID),
			Type:        "course_created",
			Description: "New course: " + c.Title,
			UserName:    teacher.Name,
			CreatedAt:   c.CreatedAt,
		})
	}

	// Get recent enrollments
	var enrollments []domain.Enrollment
	r.db.WithContext(ctx).
		Preload("User").
		Preload("Course").
		Order("enrolled_at DESC").
		Limit(limit / 3).
		Find(&enrollments)

	for _, e := range enrollments {
		activities = append(activities, domain.RecentActivity{
			ID:          e.ID,
			Type:        "enrollment",
			Description: "Enrolled in: " + e.Course.Title,
			UserName:    e.User.Name,
			CreatedAt:   e.EnrolledAt,
		})
	}

	// Simple sort by created_at
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].CreatedAt.Before(activities[j].CreatedAt) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}
