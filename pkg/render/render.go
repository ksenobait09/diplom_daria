package render

import (
	"io"

	"github.com/flosch/pongo2"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type renderer struct {
	TemplateSet *pongo2.TemplateSet
}

// Render : Custom renderer
func (r *renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	ctx := data.(pongo2.Context)

	t, err := r.TemplateSet.FromFile(name)
	if err != nil {
		return errors.Wrap(err, "failed to get template from file")
	}

	return t.ExecuteWriter(ctx, w)
}

// public/views/*.html
func New(root string) echo.Renderer {
	loader, err := pongo2.NewLocalFileSystemLoader(root)
	if err != nil {
		panic(err)
	}

	r := renderer{TemplateSet: pongo2.NewSet("templates", loader)}

	return &r // use any renderer
}
