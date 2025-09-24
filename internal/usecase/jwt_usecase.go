package usecase

import (
	_config "bitbucket.org/edts/go-task-management/config"
	_model "bitbucket.org/edts/go-task-management/internal/model"
	_customErr "bitbucket.org/edts/go-task-management/pkg/errors"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type JwtServiceInterface interface {
	GenerateToken(email string, tokenType string) (string, error)
	ParseToken(tokenStr string) (*JWTClaims, error)
}

type JWTClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("secret_key")

// GenerateToken creates JWT tokens (access & refresh)
func GenerateToken(user _model.User, tokenType string) (string, time.Time, error) {
	var expirationTime time.Duration

	if tokenType == "access" {
		expirationTime = _config.AppConfigInstance.JWT.AccessTokenTTL
	} else if tokenType == "refresh" {
		expirationTime = _config.AppConfigInstance.JWT.RefreshTokenTTL
	} else {
		return "", time.Time{}, nil
	}

	expirationDate := time.Now().Add(expirationTime).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  user.Email,
		"exp":    expirationDate,
		"type":   tokenType,
		"userId": user.ID,
	})

	finalToken, _ := token.SignedString(jwtSecret)

	return finalToken, time.Unix(expirationDate, 0), nil
}

// VerifyToken verifies a JWT and extracts the email
func VerifyToken(tokenString string) (map[string]string, error) {

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, _customErr.NewGraphQLError(http.StatusInternalServerError, "unexpected signing method")

		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email, emailOk := claims["email"].(string)
		userID, userIdOk := claims["userId"].(string)
		if !emailOk || !userIdOk {
			return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "missing or invalid email/userId claim")
		}
		// Return both email and userID in a map
		return map[string]string{
			"email":  email,
			"userId": userID,
		}, nil
	}

	return nil, errors.New("invalid token")
}

func ParseToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid JWT token")
}
