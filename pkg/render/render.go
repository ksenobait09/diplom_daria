package render

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/stnc/pongo2echo"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// public/views/*.html
func New() echo.Renderer {
	return pongo2echo.Renderer{Debug: true} // use any renderer
}
