package service

import (
	"errors"
	"fmt"

	"elearning/internal/domain"
	"elearning/internal/repository"

	"gorm.io/gorm"
)

// LessonService handles lesson business logic.
type LessonService struct {
	repo repository.LessonRepository
}

// NewLessonService returns a new LessonService.
func NewLessonService(repo repository.LessonRepository) *LessonService {
	return &LessonService{repo: repo}
}

// CreateLesson creates a lesson and auto-assigns order number.
func (s *LessonService) CreateLesson(l *domain.Lesson) error {
	courseID := int64(l.CourseID)

	lastOrder, err := s.repo.GetLastOrder(courseID)
	if err != nil {
		return fmt.Errorf("get last order for course %d: %w", courseID, err)
	}

	l.OrderNumber = lastOrder + 1

	if err := s.repo.Create(l); err != nil {
		return fmt.Errorf("create lesson for course %d: %w", courseID, err)
	}
	return nil
}

// GetLesson returns a lesson by id.
func (s *LessonService) GetLesson(id int64) (*domain.Lesson, error) {
	return s.repo.GetByID(id)
}

// GetLessonsByCourse returns lessons belonging to a course.
func (s *LessonService) GetLessonsByCourse(courseID int64) ([]domain.Lesson, error) {
	return s.repo.GetByCourse(courseID)
}

// UpdateLesson updates a lesson.
func (s *LessonService) UpdateLesson(l *domain.Lesson) error {
	if err := s.repo.Update(l); err != nil {
		return fmt.Errorf("update lesson %d: %w", l.ID, err)
	}
	return nil
}

// DeleteLesson deletes a lesson by id.
func (s *LessonService) DeleteLesson(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete lesson %d: %w", id, err)
	}
	return nil
}

// Reorder updates order_number for multiple lessons atomically.
func (s *LessonService) Reorder(courseID int64, orders map[int64]int) error {
	if len(orders) == 0 {
		return errors.New("no orders provided")
	}

	// Validate order numbers (must be >= 0 & unique)
	seen := make(map[int]bool)
	for _, order := range orders {
		if order < 0 {
			return fmt.Errorf("invalid order number %d: must be >= 0", order)
		}
		if seen[order] {
			return fmt.Errorf("duplicate order number %d", order)
		}
		seen[order] = true
	}

	for lessonID := range orders {
		lesson, err := s.repo.GetByID(lessonID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("lesson %d not found", lessonID)
		}
		if err != nil {
			return fmt.Errorf("get lesson %d: %w", lessonID, err)
		}
		if int64(lesson.CourseID) != courseID {
			return fmt.Errorf("lesson %d does not belong to course %d", lessonID, courseID)
		}
	}

	// Delegate to repository transaction
	return s.repo.Reorder(courseID, orders)
}
