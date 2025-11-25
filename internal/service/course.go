package service

import (
	"elearning/internal/domain"
	"elearning/internal/repository"
)

type CourseService interface {
	Create(c *domain.Course) error
	GetByID(id int64) (*domain.Course, error)
	GetList(filter map[string]interface{}) ([]domain.Course, error)
	Update(c *domain.Course) error
	Delete(id int64) error
	Publish(id int64, state bool) error
}

type courseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) CourseService {
	return &courseService{repo}
}

func (s *courseService) Create(c *domain.Course) error {
	return s.repo.Create(c)
}

func (s *courseService) GetByID(id int64) (*domain.Course, error) {
	return s.repo.FindByID(id)
}

func (s *courseService) GetList(filter map[string]interface{}) ([]domain.Course, error) {
	return s.repo.FindAll(filter)
}

func (s *courseService) Update(c *domain.Course) error {
	return s.repo.Update(c)
}

func (s *courseService) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *courseService) Publish(id int64, state bool) error {
	course, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	course.IsPublished = state
	return s.repo.Update(course)
}
