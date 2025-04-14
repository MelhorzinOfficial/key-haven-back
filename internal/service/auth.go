package service

import (
	"context"
	"errors"
	"key-haven-back/internal/domain/user"
	"key-haven-back/internal/repository"
	"key-haven-back/internal/service/dto"
	"key-haven-back/pkg/cache"
	"key-haven-back/pkg/secret"
	"sync"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, request *dto.CreateUserRequest) (*user.User, error)
	Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error)
	GetMe(ctx context.Context, userID string) (*dto.UserResponse, error)
}

type authService struct {
	userService UserService
	userCache   *cache.UserCache
	once        sync.Once
}

func NewAuthService(userService UserService) AuthService {
	return &authService{
		userService: userService,
	}
}

func (s *authService) initCache() {
	s.once.Do(func() {
		s.userCache = cache.NewUserCache(5*time.Minute, 10*time.Minute)
	})
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

	s.initCache()
	s.userCache.Delete(u.ID)

	u.Password = ""

	// Return login response
	return &dto.LoginResponse{
		Token: token,
		User:  *u,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID string) (*dto.UserResponse, error) {
	s.initCache()

	if cachedUser, found := s.userCache.Get(userID); found {
		return cachedUser, nil
	}

	u, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	response := &dto.UserResponse{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Avatar: "",
	}

	s.userCache.Set(userID, response)

	return response, nil
}
