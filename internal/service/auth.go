package service

import (
	"context"
	"errors"
	"key-haven-back/internal/domain/user"
	"key-haven-back/internal/repository"
	"key-haven-back/internal/service/dto"
	"key-haven-back/pkg/secret"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, request *dto.CreateUserRequest) (*user.User, error)
	Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
	userService UserService
}

func NewAuthService(userService UserService) AuthService {
	return &authService{
		userService: userService,
	}
}

func (s *authService) Register(ctx context.Context, request *dto.CreateUserRequest) (*user.User, error) {
	return s.userService.CreateUser(ctx, request)
}

func (s *authService) Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error) {
	u, err := s.userService.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {

			return nil, repository.ErrInvalidCredentials
		}
		return nil, err
	}

	valid, err := secret.VerifyPassword(u.Password, request.Password)
	if err != nil || !valid {
		return nil, repository.ErrInvalidCredentials
	}

	token, err := secret.GenerateToken(u.ID, u.Email, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	u.Password = ""

	// Return login response
	return &dto.LoginResponse{
		Token: token,
		User:  *u,
	}, nil
}
