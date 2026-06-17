package handler

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

func Render(w http.ResponseWriter, name string, errData *ErrorData, tmpl *template.Template) {
	if errData != nil {
		tmpl.Execute(w, errData)
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