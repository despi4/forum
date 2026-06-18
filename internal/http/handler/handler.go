package handler

import (
	"html/template"
	"net/http"

	render "01.tomorrow-school.ai/git/amadiuly/forum/internal/http"
	"01.tomorrow-school.ai/git/amadiuly/forum/internal/http/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	tmpl *template.Template
}

func NewHandler(tmpl *template.Template) *Handler {
	return &Handler{tmpl: tmpl}
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"IsAuthenticated": false,
	}

	_, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if ok {
		data["IsAuthenticated"] = true
	}

	render.Render(w, "home", data, nil, h.tmpl)
}
