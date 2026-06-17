package handler

import (
	"html/template"
	"net/http"
)

type Handler struct {
	tmpl *template.Template
}

func NewHandler(tmpl *template.Template) *Handler {
	return &Handler{tmpl: tmpl}
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	Render(w, "home", nil, h.tmpl)
}
