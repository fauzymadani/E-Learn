package service

import (
	"elearning/internal/domain"
	"elearning/internal/repository"
	"log"
	"os"
	"strings"
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
	repo       repository.CourseRepository
	lessonRepo repository.LessonRepository
}

func NewCourseService(repo repository.CourseRepository, lessonRepo repository.LessonRepository) CourseService {
	return &courseService{
		repo:       repo,
		lessonRepo: lessonRepo,
	}
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
	// Get course to access thumbnail
	course, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Get all lessons for this course
	lessons, err := s.lessonRepo.GetByCourse(id)
	if err != nil {
		log.Printf("failed to get lessons for course %d: %v", id, err)
	}

	// Delete all lesson files
	for _, lesson := range lessons {
		// Delete video file
		if lesson.VideoURL != "" {
			videoPath := strings.TrimPrefix(lesson.VideoURL, "/")
			if err := os.Remove(videoPath); err != nil && !os.IsNotExist(err) {
				log.Printf("failed to delete video file %s: %v", videoPath, err)
			}
		}

		// Delete lesson file (PDF, etc)
		if lesson.FileURL != "" {
			filePath := strings.TrimPrefix(lesson.FileURL, "/")
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				log.Printf("failed to delete lesson file %s: %v", filePath, err)
			}
		}
	}

	// Delete course thumbnail
	if course.Thumbnail != "" {
		thumbnailPath := strings.TrimPrefix(course.Thumbnail, "/")
		if err := os.Remove(thumbnailPath); err != nil && !os.IsNotExist(err) {
			log.Printf("failed to delete thumbnail %s: %v", thumbnailPath, err)
		}
	}

	// Delete course from database (lessons will be cascade deleted if foreign key is set)
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
