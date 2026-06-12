package dto

type LoginRequest struct {
	EmailOrUsername string
	Password        string
}

type ChangePasswordRequest struct {
	OldPassword string
	NewPassword string
}
