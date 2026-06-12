package authsvc

import (
	"context"
	"time"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/auth"
	user "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"

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
	if err := service.ValidateData(userInput); err != nil {
		return domain.Session{}, user.ErrInvalidCredentials
	}

	if userInput.Username == nil || userInput.Email == nil {
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
	if err := service.ValidateData(userInput); err != nil {
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
