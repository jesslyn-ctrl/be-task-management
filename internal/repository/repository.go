package repository

import _db "bitbucket.org/edts/go-task-management/internal/db"

type Repository struct {
	TaskRepo        TaskRepositoryInterface
	UserRepo        UserRepositoryInterface
	TeamRepo        TeamRepositoryInterface
	UserTeamRepo    UserTeamRepositoryInterface
	UserSessionRepo UserSessionRepositoryInterface
}

// NewRepository Repo dependency injection here
func NewRepository(dbConn *_db.Database) *Repository {
	return &Repository{
		TaskRepo:        NewTaskRepository(dbConn),
		UserRepo:        NewUserRepository(dbConn),
		TeamRepo:        NewTeamRepository(dbConn),
		UserTeamRepo:    NewUserTeamRepository(dbConn),
		UserSessionRepo: NewUserSessionRepository(dbConn),
	}
}
