package directives

import (
	_projection "bitbucket.org/edts/go-task-management/internal/model/projection"
	"bitbucket.org/edts/go-task-management/internal/usecase"
	_customErr "bitbucket.org/edts/go-task-management/pkg/errors"
	"context"
	"github.com/99designs/gqlgen/graphql"
	"net/http"
	"strings"
)

// AuthDirective will be used as auth middleware
func AuthDirective(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	// Extract HTTP request from the context
	req, ok := ctx.Value("httpRequest").(*http.Request)
	if !ok {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "unable to extract request from context")
	}

	// Retrieve Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return nil, _customErr.NewGraphQLError(http.StatusUnauthorized, "unauthorized: missing token")
	}

	// Parse the token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, _customErr.NewGraphQLError(http.StatusUnauthorized, "unauthorized: invalid token format")
	}

	// Verify the token
	token := parts[1]
	user, err := usecase.VerifyToken(token)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusUnauthorized, "unauthorized: invalid token")
	}

	// Add user info to context
	ctx = context.WithValue(ctx, "user", &_projection.UserContext{
		Email:  user["email"],
		UserID: user["userId"],
	})

	// Proceed to the resolver
	return next(ctx)
}
