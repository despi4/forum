package dto

type SingInRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail"`
	Password        string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string
	NewPassword string
}
