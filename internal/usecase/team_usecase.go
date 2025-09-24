package usecase

import (
	_model "bitbucket.org/edts/go-task-management/internal/model"
	_genModel "bitbucket.org/edts/go-task-management/internal/model/_generated"
	_projection "bitbucket.org/edts/go-task-management/internal/model/projection"
	_repo "bitbucket.org/edts/go-task-management/internal/repository"
	_customErr "bitbucket.org/edts/go-task-management/pkg/errors"
	"context"
	"net/http"
)

type TeamUsecaseInterface interface {
	CreateTeam(ctx context.Context, input _genModel.CreateTeamInput) (*_model.Team, error)
	UpdateTeam(ctx context.Context, input _genModel.UpdateTeamInput) (*_model.Team, error)
	GetTeamsByUser(ctx context.Context) ([]*_genModel.TeamSummary, error)
}

type TeamUsecase struct {
	teamRepo     _repo.TeamRepositoryInterface
	userRepo     _repo.UserRepositoryInterface
	userTeamRepo _repo.UserTeamRepositoryInterface
}

func NewTeamUsecase(teamRepo _repo.TeamRepositoryInterface, userRepo _repo.UserRepositoryInterface, userTeamRepo _repo.UserTeamRepositoryInterface) TeamUsecaseInterface {
	return &TeamUsecase{
		teamRepo:     teamRepo,
		userRepo:     userRepo,
		userTeamRepo: userTeamRepo,
	}
}

func (uc *TeamUsecase) CreateTeam(ctx context.Context, input _genModel.CreateTeamInput) (*_model.Team, error) {
	team := &_model.Team{
		Name:        input.Name,
		Description: input.Description,
	}

	// Save to repo
	createdTeam, err := uc.teamRepo.CreateTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	// Insert Team Assignee
	team, err = uc.userTeamRepo.InsertUserTeams(ctx, &input.Assignee, createdTeam)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, err.Error())
	}
	// Retrieve the user context data
	return createdTeam, nil
}

func (uc *TeamUsecase) UpdateTeam(ctx context.Context, input _genModel.UpdateTeamInput) (*_model.Team, error) {
	team := &_model.Team{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
	}

	// Save to repo
	updatedTeam, err := uc.teamRepo.UpdateTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	team, err = uc.userTeamRepo.InsertUserTeams(ctx, &input.Assignee, updatedTeam)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, err.Error())
	}

	return updatedTeam, nil
}

func (uc *TeamUsecase) GetTeamsByUser(ctx context.Context) ([]*_genModel.TeamSummary, error) {
	// Retrieve the user context data
	userCtx, ok := ctx.Value("user").(*_projection.UserContext)
	if !ok {
		return nil, _customErr.NewGraphQLError(http.StatusUnauthorized, "unauthorized: user context is missing or invalid")
	}
	// Fetch teams by user ID
	teams, err := uc.teamRepo.GetTeamsByUserID(ctx, userCtx.UserID)
	if err != nil {
		return nil, err
	}

	var results []*_genModel.TeamSummary
	for _, team := range teams {
		results = append(results, &_genModel.TeamSummary{
			Team:        team.Team,
			MemberCount: team.MemberCount,
		})
	}

	return results, nil
}
