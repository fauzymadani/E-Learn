package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"elearning/internal/domain"
	"elearning/internal/repository"
	"elearning/pkg/grpcclient"
)

var (
	ErrCannotEnrollInOwnCourse = errors.New("teacher cannot enroll in their own course")
	ErrCourseNotPublished      = errors.New("course is not published")
)

// EnrollmentService handles enrollment business logic
type EnrollmentService struct {
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
	userRepo       repository.UserRepository
	notifClient    *grpcclient.NotificationClient
}

// NewEnrollmentService creates a new enrollment service
func NewEnrollmentService(
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
	userRepo repository.UserRepository,
	notifClient *grpcclient.NotificationClient,
) *EnrollmentService {
	return &EnrollmentService{
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		userRepo:       userRepo,
		notifClient:    notifClient,
	}
}

// EnrollRequest represents an enrollment request
type EnrollRequest struct {
	CourseID uint `json:"course_id" binding:"required"`
}

// Enroll enrolls a student in a course
func (s *EnrollmentService) Enroll(userID uint, courseID uint) (*domain.Enrollment, error) {
	// Get course
	course, err := s.courseRepo.FindByID(int64(courseID))
	if err != nil {
		return nil, err
	}

	// Check if course is published
	if !course.IsPublished {
		return nil, ErrCourseNotPublished
	}

	// Get user to check role
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Students and admins can enroll
	// Teachers can enroll in other teacher's courses
	if user.Role == domain.RoleTeacher && course.TeacherID == int64(userID) {
		return nil, ErrCannotEnrollInOwnCourse
	}

	// Create enrollment
	enrollment := &domain.Enrollment{
		UserID:   userID,
		CourseID: courseID,
		Status:   domain.EnrollmentStatusActive,
	}

	if err := s.enrollmentRepo.Create(enrollment); err != nil {
		return nil, err
	}

	log.Printf("User %d enrolled in course %d", userID, courseID)

	// Send notification to teacher
	if s.notifClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := s.notifClient.SendNotification(ctx,
			course.TeacherID,
			"enrollment",
			"New Student Enrolled",
			fmt.Sprintf("A student has enrolled in your course: %s", course.Title),
		); err != nil {
			log.Printf("failed to send enrollment notification: %v", err)
		}
	}

	// Reload with relationships
	return s.enrollmentRepo.FindByID(enrollment.ID)
}

// Unenroll removes a student from a course
func (s *EnrollmentService) Unenroll(userID uint, courseID uint) error {
	enrollment, err := s.enrollmentRepo.FindByUserAndCourse(userID, courseID)
	if err != nil {
		return err
	}

	// Update status to dropped
	enrollment.Status = domain.EnrollmentStatusDropped
	if err := s.enrollmentRepo.Update(enrollment); err != nil {
		return err
	}

	log.Printf("User %d unenrolled from course %d", userID, courseID)
	return nil
}

// GetMyEnrollments returns all enrollments for a user
func (s *EnrollmentService) GetMyEnrollments(userID uint, status string) ([]domain.Enrollment, error) {
	enrollments, err := s.enrollmentRepo.FindByUser(userID)
	if err != nil {
		return nil, err
	}

	// Filter by status if provided
	if status != "" {
		filtered := make([]domain.Enrollment, 0)
		for _, e := range enrollments {
			if string(e.Status) == status {
				filtered = append(filtered, e)
			}
		}
		return filtered, nil
	}

	return enrollments, nil
}

// GetCourseEnrollments returns all enrollments for a course (teacher view)
func (s *EnrollmentService) GetCourseEnrollments(courseID uint, teacherID uint) ([]domain.Enrollment, error) {
	// Verify teacher owns the course
	course, err := s.courseRepo.FindByID(int64(courseID))
	if err != nil {
		return nil, err
	}

	if course.TeacherID != int64(teacherID) {
		return nil, errors.New("not authorized to view enrollments for this course")
	}

	return s.enrollmentRepo.FindByCourse(courseID)
}

// GetEnrollmentStatus gets enrollment status for a specific course
func (s *EnrollmentService) GetEnrollmentStatus(userID uint, courseID uint) (*domain.Enrollment, error) {
	return s.enrollmentRepo.FindByUserAndCourse(userID, courseID)
}

// IsEnrolled checks if a user is enrolled in a course
func (s *EnrollmentService) IsEnrolled(userID uint, courseID uint) (bool, error) {
	return s.enrollmentRepo.IsEnrolled(userID, courseID)
}

// UpdateProgress updates enrollment progress
func (s *EnrollmentService) UpdateProgress(enrollmentID uint, progress float64) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	err := s.enrollmentRepo.UpdateProgress(enrollmentID, progress)
	if err != nil {
		return err
	}

	// If progress is 100%, mark as completed
	if progress >= 100 {
		enrollment, err := s.enrollmentRepo.FindByID(enrollmentID)
		if err != nil {
			return err
		}
		enrollment.Status = domain.EnrollmentStatusCompleted
		return s.enrollmentRepo.Update(enrollment)
	}

	return nil
}
