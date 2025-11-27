package repository

import (
	"errors"

	"gorm.io/gorm"

	"elearning/internal/domain"
)

var (
	ErrAlreadyEnrolled = errors.New("student already enrolled in this course")
	ErrNotEnrolled     = errors.New("student not enrolled in this course")
)

// EnrollmentRepository handles database operations for enrollments
type EnrollmentRepository interface {
	Create(enrollment *domain.Enrollment) error
	FindByID(id uint) (*domain.Enrollment, error)
	FindByUserAndCourse(userID, courseID uint) (*domain.Enrollment, error)
	FindByUser(userID uint) ([]domain.Enrollment, error)
	FindByCourse(courseID uint) ([]domain.Enrollment, error)
	Update(enrollment *domain.Enrollment) error
	Delete(id uint) error
	IsEnrolled(userID, courseID uint) (bool, error)
	CountEnrollmentsByCourse(courseID uint) (int64, error)
	UpdateProgress(enrollmentID uint, progress float64) error
}

type enrollmentRepository struct {
	db *gorm.DB
}

// NewEnrollmentRepository creates a new enrollment repository
func NewEnrollmentRepository(db *gorm.DB) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

// Create creates a new enrollment
func (r *enrollmentRepository) Create(enrollment *domain.Enrollment) error {
	// Check if already enrolled
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("user_id = ? AND course_id = ? AND status != ?",
			enrollment.UserID, enrollment.CourseID, domain.EnrollmentStatusDropped).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return ErrAlreadyEnrolled
	}

	return r.db.Create(enrollment).Error
}

// FindByID finds an enrollment by ID
func (r *enrollmentRepository) FindByID(id uint) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := r.db.Preload("User").Preload("Course").First(&enrollment, id).Error
	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}

// FindByUserAndCourse finds an enrollment by user and course
func (r *enrollmentRepository) FindByUserAndCourse(userID, courseID uint) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).
		Preload("Course").
		First(&enrollment).Error
	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}

// FindByUser finds all enrollments for a user
func (r *enrollmentRepository) FindByUser(userID uint) ([]domain.Enrollment, error) {
	var enrollments []domain.Enrollment
	err := r.db.Where("user_id = ?", userID).
		Preload("Course").
		Order("enrolled_at DESC").
		Find(&enrollments).Error
	return enrollments, err
}

// FindByCourse finds all enrollments for a course
func (r *enrollmentRepository) FindByCourse(courseID uint) ([]domain.Enrollment, error) {
	var enrollments []domain.Enrollment
	err := r.db.Where("course_id = ?", courseID).
		Preload("User").
		Order("enrolled_at DESC").
		Find(&enrollments).Error
	return enrollments, err
}

// Update updates an enrollment
func (r *enrollmentRepository) Update(enrollment *domain.Enrollment) error {
	return r.db.Save(enrollment).Error
}

// Delete changes enrollment status to dropped (instead of soft delete)
func (r *enrollmentRepository) Delete(id uint) error {
	return r.db.Model(&domain.Enrollment{}).
		Where("id = ?", id).
		Update("status", domain.EnrollmentStatusDropped).Error
}

// IsEnrolled checks if a user is enrolled in a course
func (r *enrollmentRepository) IsEnrolled(userID, courseID uint) (bool, error) {
	var count int64
	// Consider both active and completed as enrolled; only dropped is not enrolled
	err := r.db.Model(&domain.Enrollment{}).
		Where("user_id = ? AND course_id = ? AND status IN ?",
			userID, courseID, []domain.EnrollmentStatus{domain.EnrollmentStatusActive, domain.EnrollmentStatusCompleted}).
		Count(&count).Error
	return count > 0, err
}

// CountEnrollmentsByCourse counts enrollments for a course
func (r *enrollmentRepository) CountEnrollmentsByCourse(courseID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("course_id = ? AND status = ?", courseID, domain.EnrollmentStatusActive).
		Count(&count).Error
	return count, err
}

// UpdateProgress updates the progress percentage of an enrollment
func (r *enrollmentRepository) UpdateProgress(enrollmentID uint, progress float64) error {
	return r.db.Model(&domain.Enrollment{}).
		Where("id = ?", enrollmentID).
		Update("progress_percent", progress).Error
}
