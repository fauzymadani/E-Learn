package service

import (
	"elearning/internal/domain"
	"elearning/internal/repository"
	"errors"
)

type ProgressService struct {
	progressRepo   repository.ProgressRepository
	enrollmentRepo repository.EnrollmentRepository
	lessonRepo     repository.LessonRepository
}

func NewProgressService(
	progressRepo repository.ProgressRepository,
	enrollmentRepo repository.EnrollmentRepository,
	lessonRepo repository.LessonRepository,
) *ProgressService {
	return &ProgressService{
		progressRepo:   progressRepo,
		enrollmentRepo: enrollmentRepo,
		lessonRepo:     lessonRepo,
	}
}

// MarkLessonCompleted marks a lesson as completed and checks for course completion
func (s *ProgressService) MarkLessonCompleted(userID, lessonID uint) error {
	// Get lesson to find course ID
	lesson, err := s.lessonRepo.GetByID(int64(lessonID))
	if err != nil {
		return errors.New("lesson not found")
	}

	// Check if user is enrolled
	enrolled, err := s.enrollmentRepo.IsEnrolled(userID, lesson.CourseID)
	if err != nil || !enrolled {
		return errors.New("user not enrolled in this course")
	}

	// Mark lesson as completed
	if err := s.progressRepo.MarkAsCompleted(userID, lessonID); err != nil {
		return err
	}

	// Check if course is now complete
	progress, err := s.progressRepo.GetCourseProgress(userID, lesson.CourseID)
	if err != nil {
		return err
	}

	// If course is completed, update enrollment status
	if progress.IsCompleted {
		enrollment, err := s.enrollmentRepo.FindByUserAndCourse(userID, lesson.CourseID)
		if err != nil {
			return err
		}

		enrollment.Status = domain.EnrollmentStatusCompleted
		// Note: CompletedAt field doesn't exist in database, using status only

		if err := s.enrollmentRepo.Update(enrollment); err != nil {
			return err
		}
	}

	return nil
}

// UnmarkLessonCompleted unmarks a lesson as completed
func (s *ProgressService) UnmarkLessonCompleted(userID, lessonID uint) error {
	return s.progressRepo.UnmarkCompleted(userID, lessonID)
}

// GetLessonProgress gets progress for a specific lesson
func (s *ProgressService) GetLessonProgress(userID, lessonID uint) (*domain.Progress, error) {
	return s.progressRepo.GetLessonProgress(userID, lessonID)
}

// GetCourseProgress gets overall progress for a course
func (s *ProgressService) GetCourseProgress(userID, courseID uint) (*domain.CourseProgress, error) {
	// Check if user is enrolled
	enrolled, err := s.enrollmentRepo.IsEnrolled(userID, courseID)
	if err != nil || !enrolled {
		return nil, errors.New("user not enrolled in this course")
	}

	return s.progressRepo.GetCourseProgress(userID, courseID)
}

// GetUserProgressByCourse gets all progress for a course
func (s *ProgressService) GetUserProgressByCourse(userID, courseID uint) ([]domain.Progress, error) {
	return s.progressRepo.GetUserProgressByCourse(userID, courseID)
}
