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
)

type UserService interface {
	GetProfile(userID uint) (*domain.User, error)
	UpdateProfile(userID uint, name string, avatar *string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
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
