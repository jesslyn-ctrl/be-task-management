package usecase

import (
	"context"
	"net/http"
	"time"

	_model "bitbucket.org/edts/go-task-management/internal/model"
	_genModel "bitbucket.org/edts/go-task-management/internal/model/_generated"
	_repo "bitbucket.org/edts/go-task-management/internal/repository"
	_customErr "bitbucket.org/edts/go-task-management/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecaseInterface interface {
	RegisterUser(ctx context.Context, input _genModel.CreateUserInput) (*_model.User, error)
	LoginUser(ctx context.Context, input _genModel.LoginUserInput) (*_genModel.AuthResponse, error)
	RefreshToken(ctx context.Context, input _genModel.RefreshTokenInput) (*_genModel.AuthResponse, error)
	LogoutUser(ctx context.Context, input _genModel.RefreshTokenInput) (bool, error)
}

type AuthUsecase struct {
	userRepo        _repo.UserRepositoryInterface
	userSessionRepo _repo.UserSessionRepositoryInterface
}

func NewAuthUsecase(
	userRepo _repo.UserRepositoryInterface,
	userSessionRepo _repo.UserSessionRepositoryInterface) AuthUsecaseInterface {
	return &AuthUsecase{
		userRepo:        userRepo,
		userSessionRepo: userSessionRepo,
	}
}

func (uc *AuthUsecase) RegisterUser(ctx context.Context, input _genModel.CreateUserInput) (*_model.User, error) {
	// Check if user already exists by email
	existingUser, err := uc.userRepo.GetUserByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, gqlerror.Errorf("email already registered")
	}

	// Hash the password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, gqlerror.Errorf("failed to hash password")
	}

	// Create new user entity
	newUser := &_model.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPass),
	}

	// Save user to repo
	createdUser, err := uc.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}
	// Set the password to empty (do not expose it)
	createdUser.Password = ""

	return createdUser, nil
}

func (uc *AuthUsecase) LoginUser(ctx context.Context, input _genModel.LoginUserInput) (*_genModel.AuthResponse, error) {
	//check user exist
	userExist, err := uc.userRepo.GetUserByEmail(ctx, *input.Email)
	if userExist == nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "User does not exist")
	}

	//compare the password
	err = bcrypt.CompareHashAndPassword([]byte(userExist.Password), []byte(input.Password))
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Invalid email or password")
	}

	// Generate JWT tokens
	accessToken, expiredAccessTokenDate, err := GenerateToken(*userExist, "access")
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Failed to generate access token")
	}

	// Generate refresh token
	refreshToken, expiredRefreshTokenDate, err := GenerateToken(*userExist, "refresh")
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Failed to generate refresh token")
	}

	//create new session
	userSession := &_model.UserSession{
		UserID:             userExist.ID,
		ExpiredAccessDate:  expiredAccessTokenDate,
		ExpiredRefreshDate: expiredRefreshTokenDate,
	}
	_, err = uc.userSessionRepo.CreateUserSession(ctx, userSession)
	if err != nil {
		return nil, err
	}

	// Set the password to empty (do not expose it)
	userExist.Password = ""

	// Return AuthPayload
	return &_genModel.AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         userExist,
	}, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, input _genModel.RefreshTokenInput) (*_genModel.AuthResponse, error) {
	// Verify and decode the refresh token
	user, err := VerifyToken(input.RefreshToken)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid or expired refresh token")
	}

	// Get user by email
	userExist, err := uc.userRepo.GetUserByEmail(ctx, user["email"])
	if userExist == nil {
		return nil, gqlerror.Errorf("User does not exist")
	}

	// Generate new access
	newAccessToken, accessTokenExpiredDate, err := GenerateToken(*userExist, "access")
	if err != nil {
		return nil, gqlerror.Errorf("Failed to generate new access token")
	}

	//update user session
	sessions, err := uc.userSessionRepo.GetUserSessionByUserId(ctx, userExist.ID)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Failed to get user session")
	}

	if sessions != nil {
		//update session by userId
		sessions := &_model.UserSession{
			UserID:            userExist.ID,
			ExpiredAccessDate: accessTokenExpiredDate,
		}

		_, err = uc.userSessionRepo.CreateUserSession(ctx, sessions)
		if err != nil {
			return nil, err
		}
	}

	// Set the password to empty (do not expose it)
	userExist.Password = ""

	// Return new tokens
	return &_genModel.AuthResponse{
		Token:        newAccessToken,
		RefreshToken: input.RefreshToken,
		User:         userExist,
	}, nil
}

func (uc *AuthUsecase) LogoutUser(ctx context.Context, input _genModel.RefreshTokenInput) (bool, error) {
	// Decode the JWT
	claims, err := ParseToken(input.RefreshToken)
	if err != nil {
		return false, _customErr.NewGraphQLError(http.StatusBadRequest, "invalid token")
	}

	userID := claims.UserID
	if userID == "" {
		return false, _customErr.NewGraphQLError(http.StatusBadRequest, "user ID not found in token")
	}

	// Get the session using user ID
	session, err := uc.userSessionRepo.GetUserSessionByUserId(ctx, userID)
	if err != nil {
		return false, _customErr.NewGraphQLError(http.StatusBadRequest, "failed to get user session")
	}

	if session == nil {
		// Session does not exist — treat as already logged out
		return true, nil
	}

	if !session.ExpiredRefreshDate.After(time.Now()) {
		return false, _customErr.NewGraphQLError(http.StatusUnauthorized, "Refresh Token is Expired")
	}

	// Delete the session by session ID
	err = uc.userSessionRepo.DeleteSessionByUserId(ctx, userID)
	if err != nil {
		return false, _customErr.NewGraphQLError(http.StatusBadRequest, "failed to delete user session")
	}

	// Step 4: Return success
	return true, nil
}
