package repository

import (
	"elearning/internal/domain"
	"time"

	"gorm.io/gorm"
)

type ProgressRepository interface {
	MarkAsCompleted(userID, lessonID uint) error
	UnmarkCompleted(userID, lessonID uint) error
	GetLessonProgress(userID, lessonID uint) (*domain.Progress, error)
	GetCourseProgress(userID, courseID uint) (*domain.CourseProgress, error)
	GetUserProgressByCourse(userID, courseID uint) ([]domain.Progress, error)
	IsLessonCompleted(userID, lessonID uint) (bool, error)
}

type progressRepository struct {
	db *gorm.DB
}

func NewProgressRepository(db *gorm.DB) ProgressRepository {
	return &progressRepository{db: db}
}

// MarkAsCompleted marks a lesson as completed
func (r *progressRepository) MarkAsCompleted(userID, lessonID uint) error {
	now := time.Now()
	progress := domain.Progress{
		UserID:      userID,
		LessonID:    lessonID,
		IsCompleted: true,
		CompletedAt: &now,
	}

	// Upsert: create or update if exists
	return r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Assign(map[string]interface{}{
			"is_completed": true,
			"completed_at": now,
		}).
		FirstOrCreate(&progress).Error
}

// UnmarkCompleted unmarks a lesson as completed
func (r *progressRepository) UnmarkCompleted(userID, lessonID uint) error {
	return r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Delete(&domain.Progress{}).Error
}

// GetLessonProgress gets progress for a specific lesson
func (r *progressRepository) GetLessonProgress(userID, lessonID uint) (*domain.Progress, error) {
	var progress domain.Progress
	err := r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetCourseProgress calculates overall progress for a course
func (r *progressRepository) GetCourseProgress(userID, courseID uint) (*domain.CourseProgress, error) {
	var totalLessons int64
	var completedLessons int64

	// Count total lessons in the course
	if err := r.db.Model(&domain.Lesson{}).
		Where("course_id = ?", courseID).
		Count(&totalLessons).Error; err != nil {
		return nil, err
	}

	// Count completed lessons
	if err := r.db.Model(&domain.Progress{}).
		Joins("JOIN lessons ON lessons.id = progress.lesson_id").
		Where("progress.user_id = ? AND lessons.course_id = ? AND progress.is_completed = ?",
			userID, courseID, true).
		Count(&completedLessons).Error; err != nil {
		return nil, err
	}

	percentage := float64(0)
	if totalLessons > 0 {
		percentage = (float64(completedLessons) / float64(totalLessons)) * 100
	}

	return &domain.CourseProgress{
		CourseID:           courseID,
		TotalLessons:       int(totalLessons),
		CompletedLessons:   int(completedLessons),
		ProgressPercentage: percentage,
		IsCompleted:        totalLessons > 0 && completedLessons == totalLessons,
	}, nil
}

// GetUserProgressByCourse gets all progress records for a user in a course
func (r *progressRepository) GetUserProgressByCourse(userID, courseID uint) ([]domain.Progress, error) {
	var progress []domain.Progress
	err := r.db.Joins("JOIN lessons ON lessons.id = progress.lesson_id").
		Where("progress.user_id = ? AND lessons.course_id = ?", userID, courseID).
		Preload("Lesson").
		Find(&progress).Error
	return progress, err
}

// IsLessonCompleted checks if a lesson is completed by user
func (r *progressRepository) IsLessonCompleted(userID, lessonID uint) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Progress{}).
		Where("user_id = ? AND lesson_id = ? AND is_completed = ?", userID, lessonID, true).
		Count(&count).Error
	return count > 0, err
}
