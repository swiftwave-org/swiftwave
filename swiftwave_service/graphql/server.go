package graphql

import (
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
)

func (server *Server) Initialize() {
	graphqlHandler := handler.New(
		NewExecutableSchema(
			Config{Resolvers: &Resolver{
				Config:         *server.Config,
				ServiceManager: *server.ServiceManager,
				WorkerManager:  *server.WorkerManager,
			}},
		),
	)

	graphqlHandler.AddTransport(transport.POST{})
	//graphqlHandler.AddTransport(&transport.Websocket{
	//	KeepAlivePingInterval: 10 * time.Second,
	//	Upgrader: websocket.Upgrader{
	//		CheckOrigin: func(r *http.Request) bool {
	//			return true
	//		},
	//	},
	//	// Authentication middleware for websocket connections
	//	InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
	//		jwtToken := strings.ReplaceAll(strings.ReplaceAll(initPayload.Authorization(), "Bearer ", ""), "bearer", "")
	//		if jwtToken == "" {
	//			return ctx, nil, errors.New("missing jwt token")
	//		}
	//		//nolint:staticcheck
	//		ctx = context.WithValue(ctx, "jwt_data", jwtToken)
	//		// decode jwt token
	//		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
	//			return []byte(server.Config.SystemConfig.JWTSecretKey), nil
	//		})
	//		if err != nil {
	//			return ctx, nil, errors.New("invalid jwt token")
	//		}
	//		if !token.Valid {
	//			return ctx, nil, errors.New("invalid jwt token")
	//		}
	//		claims := token.Claims.(jwt.MapClaims)
	//		// check if username is present
	//		if _, ok := claims["username"]; !ok {
	//			return ctx, nil, errors.New("malformed jwt token")
	//		}
	//		// Data in context is available in all resolvers
	//		username := claims["username"].(string)
	//		//nolint:staticcheck
	//		ctx = context.WithValue(ctx, "authorized", true)
	//		//nolint:staticcheck
	//		ctx = context.WithValue(ctx, "username", username)
	//		return ctx, nil, nil
	//	},
	//})

	graphqlHandler.Use(extension.Introspection{})

	server.EchoServer.GET("/graphql", func(c echo.Context) error {
		// Inject context
		req := c.Request()
		ctx := context.WithValue(req.Context(), "echoContext", c)
		req = req.WithContext(ctx)
		// Serve
		graphqlHandler.ServeHTTP(c.Response(), req)
		return nil
	})

	server.EchoServer.POST("/graphql", func(c echo.Context) error {
		// Inject context
		req := c.Request()
		ctx := context.WithValue(req.Context(), "echoContext", c)
		req = req.WithContext(ctx)
		// Serve
		graphqlHandler.ServeHTTP(c.Response(), req)
		return nil
	})

	// Create GraphQL Playground
	playgroundHandler := playground.Handler("GraphQL", "/graphql")
	server.EchoServer.GET("/playground", func(c echo.Context) error {
		playgroundHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
