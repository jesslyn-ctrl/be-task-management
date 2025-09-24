package repository

import (
	_db "bitbucket.org/edts/go-task-management/internal/db"
	_model "bitbucket.org/edts/go-task-management/internal/model"
	"context"
	"github.com/jackc/pgx/v5"
)

type UserTeamRepositoryInterface interface {
	InsertUserTeams(ctx context.Context, userIDs *[]string, team *_model.Team) (*_model.Team, error)
	DeleteUserTeamsByTeamId(ctx context.Context, teamId string) error
	GetAssigneeByTeamId(ctx context.Context, teamID string) (*_model.Team, error)
	ExistUserTeamsByTeamId(ctx context.Context, teamID string) (bool, error)
}

type UserTeamRepository struct {
	db *_db.Database
}

func NewUserTeamRepository(db *_db.Database) UserTeamRepositoryInterface {
	return &UserTeamRepository{
		db: db,
	}
}

func (r *UserTeamRepository) InsertUserTeams(ctx context.Context, userIDs *[]string, team *_model.Team) (*_model.Team, error) {
	// Insert new user-team relationships
	for _, userID := range *userIDs {

		// Insert user-team relationship if not already exists
		insertQuery := `
			INSERT INTO app.user_teams (user_id, team_id)
			VALUES (@userId, @teamId)
			ON CONFLICT (user_id, team_id) DO NOTHING
			RETURNING user_id, team_id`
		// Query arguments
		args := pgx.NamedArgs{
			"userId": userID,
			"teamId": team.ID,
		}

		err := r.db.Pool.QueryRow(ctx, insertQuery, args).Scan(&userID, &team.ID)
		if err != nil {
			// Skip the error if it's a conflict (i.e., the user-team pair already exists)
			if err.Error() == "no rows in result set" {
				continue
			}

			return team, err
		}
	}
	return team, nil
}

func (r *UserTeamRepository) GetAssigneeByTeamId(ctx context.Context, teamID string) (*_model.Team, error) {
	query := `
		SELECT
		    u.id,
		    u.name,
		    u.email
		FROM app.user_teams ut
		JOIN app.users u ON u.id = ut.user_id 
		WHERE ut.team_id = @team_id
	`
	args := pgx.NamedArgs{
		"team_id": teamID,
	}

	// Initialize team
	var team _model.Team
	// Initialize slice to hold users
	var users []*_model.User

	// Execute query and scan multiple rows into users slice
	rows, err := r.db.Pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through query result and scan each row into a User object
	for rows.Next() {
		var user _model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, err
		}
		// Append the user to the users slice
		users = append(users, &user)
	}

	// Check for any errors after iterating
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Set the Users field in the Team struct
	team.Users = users

	// Return the populated Team object
	return &team, nil
}

func (r *UserTeamRepository) DeleteUserTeamsByTeamId(ctx context.Context, teamId string) error {
	// Delete all existing user-team for the given team_id
	deleteQuery := `DELETE FROM app.user_teams WHERE team_id = @teamId`
	deleteArgs := pgx.NamedArgs{
		"teamId": teamId,
	}

	_, err := r.db.Pool.Exec(ctx, deleteQuery, deleteArgs)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserTeamRepository) ExistUserTeamsByTeamId(ctx context.Context, teamId string) (bool, error) {
	// Delete all existing user-team for the given team_id
	count := 0
	isExist := false

	existQuery := `SELECT COUNT(1) FROM app.user_teams WHERE team_id = @teamId`
	existArgs := pgx.NamedArgs{
		"teamId": teamId,
	}

	err := r.db.Pool.QueryRow(ctx, existQuery, existArgs).Scan(&count)
	if err != nil {
		return false, err
	}

	if count > 0 {
		isExist = true
	}

	return isExist, nil

}
