package repository

import (
	"elearning/internal/domain"
	"errors"

	"gorm.io/gorm"
)

type LessonRepository interface {
	Create(lesson *domain.Lesson) error
	GetByID(id int64) (*domain.Lesson, error)
	GetByCourse(courseID int64) ([]domain.Lesson, error)
	Update(lesson *domain.Lesson) error
	Delete(id int64) error
	GetLastOrder(courseID int64) (int, error)
	Reorder(courseID int64, orders map[int64]int) error
}

type lessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) LessonRepository {
	return &lessonRepository{db}
}

func (r *lessonRepository) Create(lesson *domain.Lesson) error {
	return r.db.Create(lesson).Error
}

func (r *lessonRepository) GetByID(id int64) (*domain.Lesson, error) {
	var lesson domain.Lesson
	err := r.db.Where("id = ?", id).First(&lesson).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &lesson, err
}

func (r *lessonRepository) GetByCourse(courseID int64) ([]domain.Lesson, error) {
	var lessons []domain.Lesson
	err := r.db.
		Where("course_id = ?", courseID).
		Order("order_number ASC").
		Find(&lessons).Error

	return lessons, err
}

func (r *lessonRepository) Update(lesson *domain.Lesson) error {
	return r.db.Model(lesson).Updates(lesson).Error
}

func (r *lessonRepository) Delete(id int64) error {
	res := r.db.Delete(&domain.Lesson{}, id)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return res.Error
}

func (r *lessonRepository) GetLastOrder(courseID int64) (int, error) {
	var lastOrder int
	err := r.db.
		Model(&domain.Lesson{}).
		Where("course_id = ?", courseID).
		Select("COALESCE(MAX(order_number), 0)").
		Scan(&lastOrder).Error

	return lastOrder, err
}

func (r *lessonRepository) Reorder(courseID int64, orders map[int64]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		for lessonID, newOrder := range orders {
			// update only order_number
			res := tx.Model(&domain.Lesson{}).
				Where("id = ? AND course_id = ?", lessonID, courseID).
				Update("order_number", newOrder)

			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}

		return nil
	})
}
