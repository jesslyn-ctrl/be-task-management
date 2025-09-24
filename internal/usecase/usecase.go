package usecase

import (
	_pubsub "bitbucket.org/edts/go-task-management/internal/pubsub"
	_repo "bitbucket.org/edts/go-task-management/internal/repository"
)

type Usecase struct {
	TaskUsecase TaskUsecaseInterface
	AuthUsecase AuthUsecaseInterface
	TeamUsecase TeamUsecaseInterface
	UserUsecase UserUsecaseInterface
}

// NewUsecase Usecase dependency injection here
func NewUsecase(repo *_repo.Repository, pubsub *_pubsub.PubSub) *Usecase {
	return &Usecase{
		TaskUsecase: NewTaskUsecase(repo.TaskRepo, repo.UserRepo, repo.TeamRepo, pubsub.TaskPubSub),
		AuthUsecase: NewAuthUsecase(repo.UserRepo, repo.UserSessionRepo),
		TeamUsecase: NewTeamUsecase(repo.TeamRepo, repo.UserRepo, repo.UserTeamRepo),
		UserUsecase: NewUserUsecase(repo.UserRepo, repo.TeamRepo, repo.UserTeamRepo),
	}
}
