package utils

import (
	"html/template"
	"net/http"
	"path/filepath"
	"log"
)

var templates map[string]*template.Template

func LoadTemplates() {
	templates = make(map[string]*template.Template)

	templateList := []string {
		"home", "add", "view", "dashboard", "login", "register", "currencies", "recurring", "add_recurring", "base", "defaults", "privacy", "receipts", "report", "terms",
	}

	for _, tmpl := range templateList {
		t, err := template.ParseFiles(
			"templates/base.html",
			filepath.Join("templates", tmpl+".html"),
			"templates/footer.html",
		)
		if err != nil {
			log.Printf("Error loading template %s: %v", tmpl, err)
			continue
		}
		templates[tmpl] = t
	}
}
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templateData := struct {
		PageName string
		Data     interface{}
	}{
		PageName: tmpl,
		Data:     data,
	}

	t := templates[tmpl]
	if t == nil {
		log.Printf("Template not found: %s", tmpl)
		log.Printf("Available template: %v", getTemplateNames())
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	err := t.ExecuteTemplate(w, "base", templateData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
	// Helper function
func getTemplateNames() []string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

func GetTemplateNames() []string {
	return getTemplateNames()
}
