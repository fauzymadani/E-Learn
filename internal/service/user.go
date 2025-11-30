package service

import (
	"elearning/internal/domain"
	"elearning/internal/repository"
	"elearning/pkg/hash"
	"errors"
)

var (
	ErrInvalidOldPassword = errors.New("old password is incorrect")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrEmailAlreadyExists = errors.New("email already registered")
)

type UserService interface {
	GetProfile(userID uint) (*domain.User, error)
	UpdateProfile(userID uint, name string, avatar *string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error

	// Admin methods (simplified - no IsActive)
	GetAllUsers() ([]domain.User, error)
	CreateUser(name, email, password, role string) (*domain.User, error)
	UpdateUser(userID uint, name, email, role string) error // ← Tanpa isActive
	DeleteUser(userID uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetProfile(userID uint) (*domain.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *userService) UpdateProfile(userID uint, name string, avatar *string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if name != "" {
		user.Name = name
	}
	if avatar != nil {
		user.Avatar = avatar
	}

	return s.userRepo.Update(user)
}

func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if err := hash.CheckPassword(user.Password, oldPassword); err != nil {
		return ErrInvalidOldPassword
	}

	if len(newPassword) < 6 {
		return ErrPasswordTooShort
	}

	hashedPassword, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, hashedPassword)
}

func (s *userService) GetAllUsers() ([]domain.User, error) {
	return s.userRepo.FindAll()
}

func (s *userService) CreateUser(name, email, password, role string) (*domain.User, error) {
	// Check if email already exists
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	if len(password) < 6 {
		return nil, ErrPasswordTooShort
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     domain.UserRole(role),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *userService) UpdateUser(userID uint, name, email, role string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Check email conflict
	if email != user.Email {
		existing, _ := s.userRepo.FindByEmail(email)
		if existing != nil && existing.ID != userID {
			return ErrEmailAlreadyExists
		}
	}

	user.Name = name
	user.Email = email
	user.Role = domain.UserRole(role)

	return s.userRepo.Update(user)
}

func (s *userService) DeleteUser(userID uint) error {
	return s.userRepo.Delete(userID)
}
