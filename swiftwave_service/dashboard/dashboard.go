package dashboard

import (
	"embed"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"strings"
)

var (
	//go:embed all:www
	dist embed.FS
	//go:embed www/index.html
	indexHTML     embed.FS
	distDirFS     = echo.MustSubFS(dist, "www")
	distIndexHtml = echo.MustSubFS(indexHTML, "www")
)

func RegisterHandlers(e *echo.Echo, isBootStrapping bool) {
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: func(c echo.Context) bool {
			// if request path doesn't start with /dashboard, skip
			return !strings.HasPrefix(c.Request().URL.Path, "/dashboard")
		},
		Index:      "index.html",
		Browse:     false,
		HTML5:      true,
		Filesystem: http.FS(distDirFS),
	}))
	e.FileFS("/dashboard", "index.html", distIndexHtml)
	e.StaticFS("/dashboard", distDirFS)
	if isBootStrapping {
		// Re-direct / to /dashboard/setup
		e.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, "/dashboard/setup")
		})
	} else {
		// Re-direct / to /dashboard
		e.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusFound, "/dashboard")
		})
	}
}
