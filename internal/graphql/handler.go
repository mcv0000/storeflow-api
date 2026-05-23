package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

// NewHandler creates a real gqlgen GraphQL HTTP handler
func NewHandler(resolver *Resolver) http.Handler {
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{
		Resolvers: resolver,
	}))

	return srv
}

// NewPlaygroundHandler creates a GraphQL Playground handler
func NewPlaygroundHandler() http.Handler {
	return playground.Handler("GraphQL Playground", "/graphql")
}
