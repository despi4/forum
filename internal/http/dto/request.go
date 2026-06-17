package dto

type LoginRequest struct {
	EmailOrUsername string `json:"usernameOrEmail"`
	Password        string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string
	NewPassword string
}
