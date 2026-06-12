package usersvc

import (
	"context"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/service"
	"github.com/google/uuid"
)

// проверка на сходимость с interface
var _ domain.UserService = (*UserService)(nil)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetMe(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := service.ValidateData(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, id uuid.UUID, updatedUser domain.UserUpdate) error {
	if err := service.ValidateData(&updatedUser); err != nil {
		return err
	}

	return s.repo.Update(ctx, updatedUser, id)
}

func (s *UserService) DeleteUser(ctx context.Context, actor, target uuid.UUID) error {
	if actor != target {
		if err := s.requireAdmin(ctx, actor); err != nil {
			return err
		}
	}

	return s.repo.Delete(ctx, target)
}

func (s *UserService) ListUsers(ctx context.Context, params domain.UserFilter) ([]domain.User, error) {
	if err := service.ValidateData(&params); err != nil {
		return nil, err
	}

	return s.repo.List(ctx, params)
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) ChangeRole(ctx context.Context, actor, target uuid.UUID, role domain.Role) error {
	if err := s.requireAdmin(ctx, actor); err != nil {
		return err
	}

	userWithNewRole := domain.UserUpdate{Role: &role}

	return s.repo.Update(ctx, userWithNewRole, target)
}

func (s *UserService) requireAdmin(ctx context.Context, actor uuid.UUID) error {
	actorUser, err := s.repo.GetByID(ctx, actor)
	if err != nil {
		return err
	}

	if actorUser.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	return nil
}
