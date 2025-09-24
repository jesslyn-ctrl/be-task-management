package usecase

import (
	_model "bitbucket.org/edts/go-task-management/internal/model"
	_genModel "bitbucket.org/edts/go-task-management/internal/model/_generated"
	_repo "bitbucket.org/edts/go-task-management/internal/repository"
	_customErr "bitbucket.org/edts/go-task-management/pkg/errors"
	"context"
	"net/http"
)

type UserUsecaseInterface interface {
	GetAllUser(ctx context.Context) ([]*_model.User, error)
	GetAssigneeByTeam(ctx context.Context, id string) ([]*_genModel.AssignedUsers, error)
	AssignUserToTeam(ctx context.Context, input _genModel.AssignUserToTeamInput) (*_model.Team, error)
}

type UserUsecase struct {
	userRepo     _repo.UserRepositoryInterface
	teamRepo     _repo.TeamRepositoryInterface
	userTeamRepo _repo.UserTeamRepositoryInterface
}

func NewUserUsecase(userRepo _repo.UserRepositoryInterface, teamRepo _repo.TeamRepositoryInterface, userTeamRepo _repo.UserTeamRepositoryInterface) UserUsecaseInterface {
	return &UserUsecase{
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		userTeamRepo: userTeamRepo,
	}
}

func (uc *UserUsecase) GetAllUser(ctx context.Context) ([]*_model.User, error) {
	//TODO: Call get all user repo method
	return []*_model.User{}, nil
}

func (uc *UserUsecase) GetAssigneeByTeam(ctx context.Context, teamID string) ([]*_genModel.AssignedUsers, error) {
	// Get the users assigned to the team by team ID
	users, err := uc.userTeamRepo.GetAssigneeByTeamId(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Initialize the Team object with the list of users
	usersTeam := _model.Team{Users: users.Users}

	// Initialize a slice to hold the assigned users
	var results []*_genModel.AssignedUsers

	// Iterate through the list of users and map them to AssignedUsers
	for _, user := range usersTeam.Users {
		assignedUser := &_genModel.AssignedUsers{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		}
		results = append(results, assignedUser)
	}

	// Return the populated assigned users
	return results, nil
}

func (uc *UserUsecase) AssignUserToTeam(ctx context.Context, input _genModel.AssignUserToTeamInput) (*_model.Team, error) {
	// Get the team by ID
	team, err := uc.teamRepo.GetTeamByID(ctx, input.TeamID)
	if err != nil {
		return nil, err
	}

	// Get the users by IDs
	users := make([]string, len(input.UserID))
	for i, userID := range input.UserID {
		user, err := uc.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		users[i] = user.ID
	}

	// Check user_teams is exist or not by teamId
	existsTeam, err := uc.userTeamRepo.ExistUserTeamsByTeamId(ctx, team.ID)
	if err != nil {
		return nil, err
	}

	//Delete user_teams by team id before inserting new one
	if existsTeam {
		err = uc.userTeamRepo.DeleteUserTeamsByTeamId(ctx, team.ID)
		if err != nil {
			return nil, _customErr.NewGraphQLError(http.StatusBadRequest, err.Error())
		}
	}

	// Save the user-team relationships
	if len(users) != 0 {
		team, err = uc.userTeamRepo.InsertUserTeams(ctx, &users, team)
		if err != nil {

			return nil, err
		}
	}

	return team, nil
}
