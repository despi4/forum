package http

import (
	"html/template"
	"net/http"
	"strconv"
)

type ErrorData struct {
	Title      string
	Error      string
	StatusCode string
}

func Render(w http.ResponseWriter, name string, data map[string]any, errData *ErrorData, tmpl *template.Template) {
	if errData != nil {
		tmpl.Execute(w, errData)
		return
	}

	if data != nil {
		tmpl.ExecuteTemplate(w, name, data)
		return
	}

	err := tmpl.ExecuteTemplate(w, name, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		code := strconv.Itoa(http.StatusInternalServerError)

		data := ErrorData{
			Title:      "Error " + code,
			Error:      "Internal Server Error",
			StatusCode: code,
		}

		tmpl.ExecuteTemplate(w, "error", data)
	}

	w.WriteHeader(200)
}
