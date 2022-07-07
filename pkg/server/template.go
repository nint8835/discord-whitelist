package server

import (
	"embed"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

//go:embed templates
var templateFS embed.FS

type EmbeddedTemplater struct {
	templates *template.Template
}

func (t *EmbeddedTemplater) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var _ echo.Renderer = (*EmbeddedTemplater)(nil)

func NewEmbeddedTemplater() *EmbeddedTemplater {
	return &EmbeddedTemplater{
		templates: template.Must(template.ParseFS(templateFS, "templates/*.gohtml")),
	}
}
