package errors

import (
	"context"
	"errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net/http"
)

// CustomErrorPresenter maps errors to GraphQL errors with status codes
func CustomErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) {
		return gqlErr
	}

	// Default to internal server error
	return &gqlerror.Error{
		Message: "Internal Server Error",
		Extensions: map[string]interface{}{
			"code":   "INTERNAL_SERVER_ERROR",
			"status": http.StatusInternalServerError,
		},
	}
}

// CustomRecoverFunc handles panics and returns a GraphQL error with a 500 status code
func CustomRecoverFunc(ctx context.Context, err interface{}) error {
	return &gqlerror.Error{
		Message: "Internal Server Error",
		Extensions: map[string]interface{}{
			"code":   "INTERNAL_SERVER_ERROR",
			"status": http.StatusInternalServerError,
		},
	}
}
