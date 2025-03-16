package utils

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templates := template.Must(template.ParseFiles(filepath.Join("templates", tmpl+".html")))
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
