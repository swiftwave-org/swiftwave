package dashboard

import (
	"embed"
	"github.com/labstack/echo/v4"
)

var (
	//go:embed all:www
	dist embed.FS
	//go:embed www/index.html
	indexHTML     embed.FS
	distDirFS     = echo.MustSubFS(dist, "www")
	distIndexHtml = echo.MustSubFS(indexHTML, "www")
)

func RegisterHandlers(e *echo.Echo) {
	e.FileFS("/dashboard", "index.html", distIndexHtml)
	e.StaticFS("/dashboard", distDirFS)
	// Re-direct / to /dashboard
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/dashboard")
	})
}
