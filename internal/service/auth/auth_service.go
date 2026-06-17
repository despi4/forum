package authsvc

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/auth"
	user "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	_             domain.AuthService = (*AuthService)(nil)
	usernameRegex                    = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,35}$`)
	emailRegex                       = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,50}$`)
)

const sessionDuration = 24 * time.Hour

type AuthService struct {
	sessionRepo domain.SessionRepository
	userRepo    user.UserRepository
}

func NewAuthService(sessionRepo domain.SessionRepository, userRepo user.UserRepository) *AuthService {
	return &AuthService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, userInput *domain.UserInput) (domain.Session, error) {
	if userInput == nil {
		return domain.Session{}, user.ErrInvalidArgument
	}

	if userInput.Username != nil && userInput.Email != nil {
		userInput.Username = service.Ptr(strings.TrimSpace(strings.ToLower(*userInput.Username)))
		userInput.Email = service.Ptr(strings.TrimSpace(strings.ToLower(*userInput.Email)))

		err := validInput(userInput.Username, userInput.Email, userInput.Password)
		if err != nil {
			return domain.Session{}, err
		}
	} else {
		return domain.Session{}, user.ErrInvalidArgument
	}

	passwordHash, err := generateHash(userInput.Password)
	if err != nil {
		return domain.Session{}, err
	}

	registerInput := user.User{
		Username:     *userInput.Username,
		Email:        *userInput.Email,
		PasswordHash: user.PasswordHash(passwordHash),
	}

	session, err := s.sessionRepo.CreateUserWithSession(ctx, sessionDuration, &registerInput, s.userRepo)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (s *AuthService) Login(ctx context.Context, userInput *domain.UserInput) (domain.Session, error) {
	if userInput == nil {
		return domain.Session{}, user.ErrInvalidCredentials
	}

	if userInput.Username == nil && userInput.Email != nil {
		userInput.Email = service.Ptr(strings.TrimSpace(strings.ToLower(*userInput.Email)))

		err := validInput(userInput.Username, userInput.Email, userInput.Password)
		if err != nil {
			log.Print(1)
			return domain.Session{}, user.ErrInvalidCredentials
		}
	} else if userInput.Username != nil && userInput.Email == nil {
		userInput.Username = service.Ptr(strings.TrimSpace(strings.ToLower(*userInput.Username)))

		err := validInput(userInput.Username, userInput.Email, userInput.Password)
		if err != nil {
			return domain.Session{}, user.ErrInvalidCredentials
		}
	} else {
		log.Print(3)
		return domain.Session{}, user.ErrInvalidCredentials
	}

	foundUser, err := s.findUserByEmailOrUsername(ctx, userInput)
	if err != nil {
		log.Print(4)
		return domain.Session{}, user.ErrInvalidCredentials
	}

	if foundUser == nil {
		return domain.Session{}, domain.ErrSessionNotFound
	}

	if !comparePassword(foundUser.PasswordHash, userInput.Password) {
		log.Print(5)
		return domain.Session{}, user.ErrInvalidCredentials
	}

	now := time.Now().UTC()

	session := domain.Session{
		UserID:     foundUser.ID,
		LastSeenAt: now,
		ExpiresAt:  now.Add(sessionDuration),
	}

	if err := s.sessionRepo.Create(ctx, &session, nil); err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.DeleteByID(ctx, sessionID)
}

func (s *AuthService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (domain.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return domain.Session{}, err
	}

	if session == nil {
		return domain.Session{}, domain.ErrSessionNotFound
	}

	now := time.Now().UTC()

	if !session.ExpiresAt.After(now) {
		_ = s.sessionRepo.DeleteByID(ctx, sessionID)
		return domain.Session{}, domain.ErrSessionExpired
	}

	session.LastSeenAt = now

	if err := s.sessionRepo.UpdateLastSeen(ctx, sessionID, now); err != nil {
		return domain.Session{}, err
	}

	return *session, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	foundUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := validInput(nil, nil, newPassword); err != nil {
		return err
	}

	if !comparePassword(foundUser.PasswordHash, oldPassword) {
		return user.ErrInvalidCredentials
	}

	passwordHash, err := generateHash(newPassword)
	if err != nil {
		return err
	}

	newPasswordHash := user.PasswordHash(passwordHash)

	updatePassword := user.UserUpdate{
		PasswordHash: &newPasswordHash,
	}

	if err := s.userRepo.Update(ctx, updatePassword, foundUser.ID); err != nil {
		return err
	}

	if err := s.sessionRepo.DeleteAllUserSessions(ctx, userID); err != nil {
		return err
	}

	return nil
}

// ============== Private helper methods ==============

func validInput(username, email *string, password string) error {
	if email != nil {
		if !emailRegex.Match([]byte(*email)) {
			return user.ErrInvalidArgument
		}

		if len(*email) > 50 || len(*email) < 6 {
			return fmt.Errorf("email must contain at least 6 and not more than 50 characters: %w", user.ErrInvalidArgument)
		}
	}

	if username != nil {
		if !usernameRegex.MatchString(*username) {
			return fmt.Errorf("username must be 3-35 chars and contain only A-Z, a-z, 0-9, _ and -: %w", user.ErrInvalidArgument)
		}
	}

	trimmedPassword := strings.TrimSpace(password)

	if trimmedPassword == "" {
		return user.ErrInvalidArgument
	}

	if len(trimmedPassword) < 6 || len(trimmedPassword) > 50 {
		return fmt.Errorf("password must contain at least 6 and not more than 50 characters: %w", user.ErrInvalidArgument)
	}

	return nil
}

func (s *AuthService) findUserByEmailOrUsername(ctx context.Context, userInput *domain.UserInput) (*user.User, error) {
	if userInput.Email != nil {
		return s.userRepo.GetByEmail(ctx, *userInput.Email)
	}

	if userInput.Username != nil {
		return s.userRepo.GetByUsername(ctx, *userInput.Username)
	}

	return nil, user.ErrInvalidCredentials
}

func generateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func comparePassword(hash user.PasswordHash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
