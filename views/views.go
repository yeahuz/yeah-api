package views

import (
	"embed"
	"text/template"
)

var (
	ViewsFS embed.FS
	err     error
	Base    *template.Template
	Email   *template.Template
)

func LoadViews() {
	funcs := template.FuncMap{
		"defer": func(i *int) int { return *i },
	}

	Base = template.Must(template.New("").Funcs(funcs).ParseFS(ViewsFS))
}
