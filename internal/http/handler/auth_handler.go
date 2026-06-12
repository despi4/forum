package handler

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/auth"
	svc "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/auth"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/user"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/http/dto"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/http/middleware"
	"github.com/google/uuid"
)

// encoder - Он берет вашу структуру или мапу v, трансформирует её в JSON
// decoder - Он читает JSON из потока и пытается «распаковать» (записать) эти данные в структуру v
// decoder.DisallowUnknownFields() - Запрещаем неизвестные поля в JSON
// W reponse writer - interface с помощью этого мы отправляем данные к пользователю
// R request - это то что нам присылает user

/* guide по статусе, header
1) w.Header.Set("...", "...")
2) w.WriteHeader(200)
*/

// Если использовать fetch и template во фронт, то надо
// Fetch(Post, Put, Patch, Delete)  |  Templates(Get)
// главное не надо их смешивать

type AuthHandler struct {
	authSvc svc.AuthService
	tmpl    *template.Template
}

func NewAuthHandler(authSvc svc.AuthService, tmpl *template.Template) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, tmpl: tmpl}
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	Render(w, "register", nil, h.tmpl)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var createReq auth.UserInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&createReq); err != nil {
		h.writeJSONError(w, http.StatusBadRequest, user.ErrInvalidArgument.Error())
		return
	}

	if createReq.Username == nil || createReq.Email == nil {
		h.writeJSONError(w, http.StatusBadRequest, user.ErrInvalidArgument.Error())
		return
	}

	username := strings.TrimSpace(*createReq.Username)
	email := strings.TrimSpace(*createReq.Email)

	createReq.Username = &username
	createReq.Email = &email

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	session, err := h.authSvc.Register(ctx, &createReq)
	if err != nil {
		h.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var cookie http.Cookie = http.Cookie{
		Name:     "session_id",
		Value:    session.ID.String(),
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	h.writeJSON(w, http.StatusCreated, map[string]any{
		"message": "registered succesfully",
		"user_id": session.UserID,
	})
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	Render(w, "login", nil, h.tmpl)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var (
		loginReq   dto.LoginRequest
		loginInput auth.UserInput
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&loginReq); err != nil {
		h.writeJSONError(w, http.StatusBadRequest, user.ErrInvalidArgument.Error())
		return
	}

	loginReq.EmailOrUsername = strings.TrimSpace(loginReq.EmailOrUsername)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if strings.Contains(loginReq.EmailOrUsername, "@") {
		loginInput = auth.UserInput{
			Email:    &loginReq.EmailOrUsername,
			Password: loginReq.Password,
		}
	} else {
		loginInput = auth.UserInput{
			Username: &loginReq.EmailOrUsername,
			Password: loginReq.Password,
		}
	}

	session, err := h.authSvc.Login(ctx, &loginInput)
	if err != nil {
		h.writeJSONError(w, http.StatusUnauthorized, "Invalid email/username or password")
		return
	}

	var cookie http.Cookie = http.Cookie{
		Name:     "session_id",
		Value:    session.ID.String(),
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	h.writeJSON(w, http.StatusOK, map[string]any{
		"message": "logged in successfully",
		"user_id": session.UserID,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value(middleware.CookieIDKey).(uuid.UUID)
	if !ok {
		h.writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.authSvc.Logout(r.Context(), sessionID); err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	h.writeJSON(w, http.StatusOK, map[string]any{
		"message": "logged out successfully",
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var changePasswordReq dto.ChangePasswordRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&changePasswordReq); err != nil {
		h.writeJSONError(w, http.StatusBadRequest, user.ErrInvalidCredentials.Error())
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		h.writeJSONError(w, http.StatusUnauthorized, "Unathorized")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := h.authSvc.ChangePassword(ctx, userID, changePasswordReq.OldPassword, changePasswordReq.NewPassword)
	if err != nil {
		h.writeJSONError(w, http.StatusBadRequest, "change password failed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	h.writeJSON(w, http.StatusOK, map[string]any{
		"message": "password changed successfully, please login again",
	})
}

// ============== Helper Methods ===============

func (h *AuthHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) writeJSONError(w http.ResponseWriter, errorCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)

	errRes := dto.ErrResponse{
		Error:     message,
		ErrorCode: errorCode,
	}

	_ = json.NewEncoder(w).Encode(errRes)
}
