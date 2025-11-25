package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"elearning/internal/middleware"
	"elearning/internal/repository"
	"elearning/internal/service"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Register Request"
// @Success 201 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(ctx *gin.Context) {
	var req service.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			ctx.JSON(http.StatusConflict, ErrorResponse{Error: "email already exists"})
			return
		}
		if errors.Is(err, service.ErrInvalidRole) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid role"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to register user"})
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login Request"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	var req service.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to login"})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetProfile handles get user profile
// @Summary Get user profile
// @Description Get current authenticated user profile
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} service.UserProfile
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) GetProfile(ctx *gin.Context) {
	claims, err := middleware.GetCurrentUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	profile, err := h.authService.GetProfile(claims.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get profile"})
		return
	}

	ctx.JSON(http.StatusOK, profile)
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout the current authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} LogoutResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	claims, err := middleware.GetCurrentUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// Get the actual token string
	tokenString, err := middleware.GetCurrentToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "failed to get token"})
		return
	}

	// Blacklist the token
	err = h.authService.Logout(claims.UserID, tokenString)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, LogoutResponse{
		Message: "successfully logged out",
	})
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// LogoutResponse represents logout response
type LogoutResponse struct {
	Message string `json:"message"`
}
