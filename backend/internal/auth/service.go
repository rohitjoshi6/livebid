package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/rohitjoshi6/livebid/backend/internal/users"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrInvalidRegistration = errors.New("invalid registration")

type UserRepository interface {
	Create(ctx context.Context, email, username, passwordHash string) (users.User, error)
	FindByEmail(ctx context.Context, email string) (users.User, error)
}

type Service struct {
	users  UserRepository
	tokens *TokenManager
	now    func() time.Time
}

type AuthResult struct {
	Token string           `json:"access_token"`
	User  users.PublicUser `json:"user"`
}

func NewService(userRepo UserRepository, tokens *TokenManager) *Service {
	return &Service{
		users:  userRepo,
		tokens: tokens,
		now:    time.Now,
	}
}

func (s *Service) Register(ctx context.Context, email, username, password string) (AuthResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	if !validEmail(email) || len(username) < 3 || len(password) < 8 {
		return AuthResult{}, ErrInvalidRegistration
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return AuthResult{}, err
	}
	user, err := s.users.Create(ctx, email, username, passwordHash)
	if err != nil {
		return AuthResult{}, err
	}
	return s.authResult(user)
}

func (s *Service) Login(ctx context.Context, email, password string) (AuthResult, error) {
	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}
	if !CheckPassword(password, user.PasswordHash) {
		return AuthResult{}, ErrInvalidCredentials
	}
	return s.authResult(user)
}

func (s *Service) authResult(user users.User) (AuthResult, error) {
	token, err := s.tokens.Create(user.ID, user.Email, s.now())
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		Token: token,
		User:  user.Public(),
	}, nil
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
