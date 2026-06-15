package authsvc

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/auth"
	user "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	usersvc "01.tomorrow-school.ai/git/amadiuly/forum/internal/service/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	_ domain.AuthService = (*AuthService)(nil)
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
	if userInput.Username != nil && userInput.Email != nil {
		userInput.Username = usersvc.Ptr(strings.TrimSpace(*userInput.Username))
		userInput.Email = usersvc.Ptr(strings.TrimSpace(strings.ToLower(*userInput.Email)))

		err := s.validInput(userInput.Username, userInput.Email, userInput.Password)
		if err != nil {
			return domain.Session{}, err
		}
	} else {
		return domain.Session{}, user.ErrInvalidCredentials
	}

	passwordHash, err := s.generateHash(userInput.Password)
	if err != nil {
		return domain.Session{}, err
	}

	registerInput := user.User{
		Username:     *userInput.Username,
		Email:        *userInput.Email,
		PasswordHash: user.PasswordHash(passwordHash),
	}

	createdUser, err := s.userRepo.Create(ctx, &registerInput)
	if err != nil {
		return domain.Session{}, err
	}

	now := time.Now()

	session := domain.Session{
		UserID:     createdUser.ID,
		LastSeenAt: now,
		ExpiresAt:  now.Add(sessionDuration),
	}

	if err := s.sessionRepo.Create(ctx, &session); err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (s *AuthService) Login(ctx context.Context, userInput *domain.UserInput) (domain.Session, error) {
	if userInput.Username == nil && userInput.Email != nil {
		userInput.Email = usersvc.Ptr(strings.TrimSpace(strings.ToLower(*userInput.Email)))

		err := s.validInput(userInput.Username, userInput.Email, userInput.Password)
		if err != nil {
			return domain.Session{}, err
		}
	} else if userInput.Username != nil && userInput.Email == nil {
		userInput.Username = usersvc.Ptr(strings.TrimSpace(*userInput.Username))

		err := s.validInput(userInput.Username, userInput.Email, userInput.Password)
		if err != nil {
			return domain.Session{}, err
		}
	} else {
		return domain.Session{}, user.ErrInvalidCredentials
	}

	foundUser, err := s.findUserByEmailOrUsername(ctx, userInput)
	if err != nil {
		return domain.Session{}, user.ErrInvalidCredentials
	}

	if !s.comparePassword(foundUser.PasswordHash, userInput.Password) {
		return domain.Session{}, user.ErrInvalidCredentials
	}

	now := time.Now()

	session := domain.Session{
		UserID:     foundUser.ID,
		LastSeenAt: now,
		ExpiresAt:  now.Add(sessionDuration),
	}

	if err := s.sessionRepo.Create(ctx, &session); err != nil {
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

	now := time.Now()

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

	if err := s.validInput(nil, nil, newPassword); err != nil {
		return err
	}

	if !s.comparePassword(foundUser.PasswordHash, oldPassword) {
		return user.ErrInvalidCredentials
	}

	passwordHash, err := s.generateHash(newPassword)
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

func (s *AuthService) validInput(username, email *string, password string) error {
	if email != nil {
		_, err := mail.ParseAddress(*email)
		if err != nil {
			return user.ErrInvalidArgument
		}

		if len(*email) > 50 || len(*email) < 6 {
			return fmt.Errorf("email must contain at least 6 and not more than 50 characters: %w", user.ErrInvalidArgument)
		}
	}

	if username != nil {
		if len(*username) > 35 || len(*username) < 3 {
			return fmt.Errorf("username must contain at least 3 and not more than 35 characters: %w", user.ErrInvalidArgument)
		}
	}

	if len(password) < 6 || len(password) > 50 {
		return fmt.Errorf("passwor must contain at lleast 6 and not more than 50 characters: %w", user.ErrInvalidArgument)
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

func (s *AuthService) generateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (s *AuthService) comparePassword(hash user.PasswordHash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
