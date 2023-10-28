package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
)

func (server *Server) Initialize() {
	graphqlHandler := handler.NewDefaultServer(
		NewExecutableSchema(
			Config{Resolvers: &Resolver{
				ServiceConfig:  *server.ServiceConfig,
				ServiceManager: *server.ServiceManager,
			}},
		),
	)
	playgroundHandler := playground.Handler("GraphQL", "/graphql")

	server.EchoServer.POST("/graphql", func(c echo.Context) error {
		graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	server.EchoServer.GET("/playground", func(c echo.Context) error {
		playgroundHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
