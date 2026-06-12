package dto

type ErrResponse struct {
	Error     string `json:"error"`
	ErrorCode int    `json:"error_code"`
}
