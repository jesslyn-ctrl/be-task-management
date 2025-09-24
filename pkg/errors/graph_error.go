package errors

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net/http"
)

// Map HTTP status to GraphQL error codes
var statusToCode = map[int]string{
	http.StatusBadRequest:          "BAD_REQUEST",
	http.StatusUnauthorized:        "UNAUTHORIZED",
	http.StatusForbidden:           "FORBIDDEN",
	http.StatusNotFound:            "NOT_FOUND",
	http.StatusInternalServerError: "INTERNAL_SERVER_ERROR",
}

// NewGraphQLError creates a new GraphQL error with a mapped code
func NewGraphQLError(status int, message string) *gqlerror.Error {
	code, exists := statusToCode[status]
	if !exists {
		code = "UNKNOWN_ERROR" // Default fallback
	}

	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code":   code,
			"status": status,
		},
	}
}
