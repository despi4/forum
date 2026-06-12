package category

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

var (
	ErrCategoryNotFound error = errors.New("category not found")
	ErrCategoryExists error = errors.New("category exists")
)
