package repository

import (
	_db "bitbucket.org/edts/go-task-management/internal/db"
	_model "bitbucket.org/edts/go-task-management/internal/model"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type UserSessionRepositoryInterface interface {
	CreateUserSession(ctx context.Context, userSession *_model.UserSession) (*_model.UserSession, error)
	GetUserSessionByUserId(ctx context.Context, userId string) (*_model.UserSession, error)
	DeleteSessionByUserId(ctx context.Context, userId string) error
	UpdateSession(ctx context.Context, userSession *_model.UserSession) (*_model.UserSession, error)
}

type UserSessionRepository struct {
	db *_db.Database
}

func NewUserSessionRepository(db *_db.Database) UserSessionRepositoryInterface {
	return &UserSessionRepository{
		db: db,
	}
}

func (r *UserSessionRepository) CreateUserSession(ctx context.Context, userSession *_model.UserSession) (*_model.UserSession, error) {
	insertQuery := `INSERT INTO app.user_sessions ("expired_access_date", "expired_refresh_date", "user_id")
		VALUES (@expiredAccessDate, @expiredRefreshDate, @userId)
		RETURNING "expired_access_date", "expired_refresh_date", "user_id"`

	insertArgs := pgx.NamedArgs{
		"expiredAccessDate":  userSession.ExpiredAccessDate,
		"expiredRefreshDate": userSession.ExpiredRefreshDate,
		"userId":             userSession.UserID,
	}

	err := r.db.Pool.QueryRow(ctx, insertQuery, insertArgs).Scan(&userSession.ExpiredAccessDate, &userSession.ExpiredRefreshDate, &userSession.UserID)

	if err != nil {
		fmt.Printf("", err)
		return nil, err
	}

	return userSession, nil

}

func (r *UserSessionRepository) GetUserSessionByUserId(ctx context.Context, userId string) (*_model.UserSession, error) {
	query := `SELECT id, expired_access_date, expired_refresh_date, user_id FROM app.user_sessions WHERE user_id = @userId`
	args := pgx.NamedArgs{
		"userId": userId,
	}

	var session _model.UserSession
	scan := func(row pgx.Row) error {
		return row.Scan(
			&session.ID,
			&session.ExpiredAccessDate,
			&session.ExpiredRefreshDate,
			&session.UserID,
		)
	}

	err := scan(r.db.Pool.QueryRow(ctx, query, args))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *UserSessionRepository) DeleteSessionByUserId(ctx context.Context, userId string) error {
	query := `
		DELETE FROM app.user_sessions WHERE user_id = $1;
	`

	_, err := r.db.Pool.Exec(ctx, query, userId)
	if err != nil {
		return err
	}
	return nil

}

func (r *UserSessionRepository) UpdateSession(ctx context.Context, userSession *_model.UserSession) (*_model.UserSession, error) {
	query := `
		UPDATE app.user_sessions
		SET expired_access_date = @expiredAccessDate, 
		    modified_at = current_timestamp,
		    modified_by = 'system'
		WHERE user_id = @userId
		RETURNING user_id, expired_access_date, expired_refresh_date, modified_at
	`

	// Query arguments
	args := pgx.NamedArgs{
		"userId":            userSession.UserID,
		"expiredAccessDate": userSession.ExpiredAccessDate,
	}

	err := r.db.Pool.QueryRow(ctx, query, args).Scan(&userSession.UserID, &userSession.ExpiredAccessDate, &userSession.ExpiredRefreshDate, &userSession.ModifiedBy)
	if err != nil {
		return nil, err
	}
	return userSession, nil
}
