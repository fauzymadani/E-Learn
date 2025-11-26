package service

import (
	"context"
	"elearning/internal/domain"
	"elearning/internal/repository"
	"elearning/pkg/grpcclient"
	"fmt"
	"time"
)

type DashboardService interface {
	GetStudentDashboard(ctx context.Context, userID uint) (*domain.StudentDashboard, error)
	GetTeacherDashboard(ctx context.Context, teacherID int64) (*domain.TeacherDashboard, error)
	GetAdminDashboard(ctx context.Context) (*domain.AdminDashboard, error)
}

type dashboardService struct {
	dashboardRepo repository.DashboardRepository
	notifClient   *grpcclient.NotificationClient
	userRepo      repository.UserRepository
}

func NewDashboardService(
	dashboardRepo repository.DashboardRepository,
	notifClient *grpcclient.NotificationClient,
	userRepo repository.UserRepository,
) DashboardService {
	return &dashboardService{
		dashboardRepo: dashboardRepo,
		notifClient:   notifClient,
		userRepo:      userRepo,
	}
}

func (s *dashboardService) GetStudentDashboard(ctx context.Context, userID uint) (*domain.StudentDashboard, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.Role != domain.RoleStudent {
		return nil, fmt.Errorf("user is not a student")
	}

	enrolledCourses, err := s.dashboardRepo.GetStudentEnrolledCourses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrolled courses: %w", err)
	}

	stats, err := s.dashboardRepo.GetStudentStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student stats: %w", err)
	}

	// Get notifications via gRPC
	var notifications []domain.Notification
	notifCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	notifResp, err := s.notifClient.GetNotifications(notifCtx, int64(userID), 1, 5, false)
	if err == nil && notifResp != nil {
		for _, n := range notifResp.Notifications {
			createdAt := time.Now()
			if parsedTime, err := time.Parse(time.RFC3339, n.CreatedAt); err == nil {
				createdAt = parsedTime
			}

			notifications = append(notifications, domain.Notification{
				ID:        uint(n.Id),
				UserID:    uint(n.UserId),
				Type:      domain.NotificationType(n.Type),
				Title:     n.Title,
				Message:   n.Message,
				IsRead:    n.IsRead,
				CreatedAt: createdAt,
			})
		}
	}

	learningProgress := make([]domain.CourseProgressSummary, len(enrolledCourses))
	for i, course := range enrolledCourses {
		learningProgress[i] = domain.CourseProgressSummary{
			CourseID:         course.ID,
			CourseTitle:      course.Title,
			TotalLessons:     course.TotalLessons,
			CompletedLessons: course.CompletedLessons,
			ProgressPercent:  course.ProgressPercent,
			LastAccessedAt:   nil,
		}
	}

	return &domain.StudentDashboard{
		EnrolledCourses:     enrolledCourses,
		LearningProgress:    learningProgress,
		RecentNotifications: notifications,
		Stats:               *stats,
	}, nil
}

func (s *dashboardService) GetTeacherDashboard(ctx context.Context, teacherID int64) (*domain.TeacherDashboard, error) {
	user, err := s.userRepo.FindByID(uint(teacherID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.Role != domain.RoleTeacher {
		return nil, fmt.Errorf("user is not a teacher")
	}

	courses, err := s.dashboardRepo.GetTeacherCourses(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher courses: %w", err)
	}

	stats, err := s.dashboardRepo.GetTeacherStats(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher stats: %w", err)
	}

	recentEnrollments, err := s.dashboardRepo.GetRecentEnrollments(ctx, teacherID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent enrollments: %w", err)
	}

	return &domain.TeacherDashboard{
		MyCourses:         courses,
		TotalStudents:     stats.TotalStudents,
		RecentEnrollments: recentEnrollments,
		Stats:             *stats,
	}, nil
}

func (s *dashboardService) GetAdminDashboard(ctx context.Context) (*domain.AdminDashboard, error) {
	adminStats, err := s.dashboardRepo.GetAdminStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin stats: %w", err)
	}

	recentActivities, err := s.dashboardRepo.GetRecentActivities(ctx, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}

	var totalUsers, totalEnrollments int

	for _, count := range adminStats.UsersByRole {
		totalUsers += count
	}

	for _, count := range adminStats.EnrollmentsByStatus {
		totalEnrollments += count
	}

	return &domain.AdminDashboard{
		TotalUsers:       totalUsers,
		TotalCourses:     adminStats.CoursesByStatus.Total,
		TotalEnrollments: totalEnrollments,
		Statistics:       *adminStats,
		RecentActivities: recentActivities,
	}, nil
}
