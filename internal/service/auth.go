package service

import (
	"errors"
	"log"
	"time"

	"elearning/internal/domain"
	"elearning/internal/repository"
	"elearning/pkg/hash"
	"elearning/pkg/token"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRole        = errors.New("invalid role")
)

// RegisterRequest represents registration request
type RegisterRequest struct {
	Name     string          `json:"name" binding:"required"`
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=6"`
	Role     domain.UserRole `json:"role" binding:"required"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	AccessToken string      `json:"access_token"`
	User        UserProfile `json:"user"`
}

// UserProfile represents user profile
type UserProfile struct {
	ID    uint            `json:"id"`
	Name  string          `json:"name"`
	Email string          `json:"email"`
	Role  domain.UserRole `json:"role"`
}

// AuthService handles authentication business logic
type AuthService interface {
	Register(req RegisterRequest) (*AuthResponse, error)
	Login(req LoginRequest) (*AuthResponse, error)
	GetProfile(userID uint) (*UserProfile, error)
	Logout(userID uint, token string) error
}

type authService struct {
	userRepo   repository.UserRepository
	tokenMaker token.TokenMaker
	blacklist  token.TokenBlacklist
	jwtExpiry  time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	tokenMaker token.TokenMaker,
	blacklist token.TokenBlacklist,
	jwtExpiry time.Duration,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
		blacklist:  blacklist,
		jwtExpiry:  jwtExpiry,
	}
}

// Register registers a new user
func (s *authService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Validate role
	if !req.Role.IsValid() {
		return nil, ErrInvalidRole
	}

	// Hash password
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate token
	accessToken, err := s.tokenMaker.CreateToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.jwtExpiry,
	)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken: accessToken,
		User: UserProfile{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

// Login authenticates a user
func (s *authService) Login(req LoginRequest) (*AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if err := hash.CheckPassword(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	accessToken, err := s.tokenMaker.CreateToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.jwtExpiry,
	)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken: accessToken,
		User: UserProfile{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

// GetProfile gets user profile
func (s *authService) GetProfile(userID uint) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return &UserProfile{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

// Logout handles user logout
func (s *authService) Logout(userID uint, tokenString string) error {
	// Verify the token and get its expiration time
	claims, err := s.tokenMaker.VerifyToken(tokenString)
	if err != nil {
		// Token is already invalid, nothing to blacklist
		return nil
	}

	// Log the logout event
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	log.Printf("User logged out: ID=%d, Email=%s", user.ID, user.Email)

	// Add token to blacklist with its expiration time
	if s.blacklist != nil {
		expiresAt := claims.ExpiresAt.Time
		if err := s.blacklist.Add(tokenString, expiresAt); err != nil {
			log.Printf("Failed to blacklist token: %v", err)
			return err
		}
		log.Printf("Token blacklisted until %v", expiresAt)
	}

	return nil
}
