package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/auth"
	"github.com/google/uuid"

	"01.tomorrow-school.ai/git/amadiuly/forum/internal/http/dto"
)

const UserIDKey string = "user_id"
const CookieIDKey string = "session_id"

func Logger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func AuthMiddleware(authSvc domain.AuthService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieIDKey)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				writeJSONError(w, http.StatusUnauthorized, "Unathorized: no session cookie")
				return
			}

			writeJSONError(w, http.StatusBadRequest, "can not read cookie")
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "Unathorized: invalid session ID")
			return
		}

		session, err := authSvc.ValidateSession(r.Context(), sessionID)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "Unathorized: invalid or expired session")
		}

		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
		ctx = context.WithValue(ctx, CookieIDKey, session.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
	})
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(dto.ErrResponse{
		Error:     message,
		ErrorCode: statusCode,
	})
}
