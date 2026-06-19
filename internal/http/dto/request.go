package dto

import (
	"time"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/category"
)

type SingInRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail"`
	Password        string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	AuthorID  string `json:"authorID"`
	CreatedAt  time.Time `json:"createdAt"`
	CategoryID category.Category `json:"categoryID"`
}
