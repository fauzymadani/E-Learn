package repository

import (
	"elearning/internal/domain"

	"gorm.io/gorm"
)

type CourseRepository interface {
	Create(course *domain.Course) error
	FindByID(id int64) (*domain.Course, error)
	FindAll(filter map[string]interface{}) ([]domain.Course, error)
	Update(course *domain.Course) error
	Delete(id int64) error
}

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db}
}

func (r *courseRepository) Create(course *domain.Course) error {
	return r.db.Create(course).Error
}

func (r *courseRepository) FindByID(id int64) (*domain.Course, error) {
	var course domain.Course
	err := r.db.Where("id = ?", id).First(&course).Error
	return &course, err
}

func (r *courseRepository) FindAll(filter map[string]interface{}) ([]domain.Course, error) {
	var courses []domain.Course
	q := r.db.Model(&domain.Course{})

	if title, ok := filter["title"]; ok {
		q = q.Where("title ILIKE ?", "%"+title.(string)+"%")
	}

	if categoryID, ok := filter["category_id"]; ok {
		q = q.Where("category_id = ?", categoryID)
	}

	if published, ok := filter["is_published"]; ok {
		q = q.Where("is_published = ?", published)
	}

	err := q.Order("created_at DESC").Find(&courses).Error
	return courses, err
}

func (r *courseRepository) Update(course *domain.Course) error {
	return r.db.Save(course).Error
}

func (r *courseRepository) Delete(id int64) error {
	return r.db.Delete(&domain.Course{}, id).Error
}
