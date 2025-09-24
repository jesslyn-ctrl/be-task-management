package repository

import (
	_db "bitbucket.org/edts/go-task-management/internal/db"
	_model "bitbucket.org/edts/go-task-management/internal/model"
	_projection "bitbucket.org/edts/go-task-management/internal/model/projection"
	"context"
	"github.com/jackc/pgx/v5"
)

type TeamRepositoryInterface interface {
	CreateTeam(ctx context.Context, team *_model.Team) (*_model.Team, error)
	UpdateTeam(ctx context.Context, team *_model.Team) (*_model.Team, error)
	GetTeamByID(ctx context.Context, teamID string) (*_model.Team, error)
	GetTeamsByIDs(ctx context.Context, teamIDs []string) ([]*_model.Team, error)
	GetTeamsByUserID(ctx context.Context, userID string) ([]*_projection.TeamSummary, error)
}

type TeamRepository struct {
	db *_db.Database
}

func NewTeamRepository(db *_db.Database) TeamRepositoryInterface {
	return &TeamRepository{
		db: db,
	}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *_model.Team) (*_model.Team, error) {
	query := `
		INSERT INTO app.teams ("name", description, created_at, modified_at, created_by, modified_by)
		VALUES (@name, @description, current_timestamp, current_timestamp, @created_by, @modified_by)
		RETURNING id, created_at
	`

	// Query arguments
	args := pgx.NamedArgs{
		"name":        team.Name,
		"description": team.Description,
	}

	err := r.db.Pool.QueryRow(ctx, query, args).Scan(&team.ID, &team.CreatedAt)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, team *_model.Team) (*_model.Team, error) {
	query := `
		UPDATE app.teams
		SET name = @name, 
		    description = @description,
		    modified_at = current_timestamp,
		    modified_by = @modified_by
		WHERE id = @id
		RETURNING id, created_at, modified_at
	`

	// Query arguments
	args := pgx.NamedArgs{
		"id":          team.ID,
		"name":        team.Name,
		"description": team.Description,
		"modified_by": team.ModifiedBy,
	}

	err := r.db.Pool.QueryRow(ctx, query, args).Scan(&team.ID, &team.CreatedAt, &team.ModifiedAt)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, teamID string) (*_model.Team, error) {
	query := `
		SELECT
		    t.id,
		    t.name,
		    t.created_at,
		    t.modified_at,
		    t.created_by,
		    t.modified_by
		FROM app.teams t 
		WHERE t.id = @team_id
`
	args := pgx.NamedArgs{
		"team_id": teamID,
	}

	var team _model.Team
	scan := func(row pgx.Row) error {
		return row.Scan(
			&team.ID,
			&team.Name,
			&team.CreatedAt,
			&team.ModifiedAt,
			&team.CreatedBy,
			&team.ModifiedBy,
		)
	}

	err := scan(r.db.Pool.QueryRow(ctx, query, args))
	if err != nil {
		return nil, err
	}

	return &team, nil
}

func (r *TeamRepository) GetTeamsByIDs(ctx context.Context, teamIDs []string) ([]*_model.Team, error) {
	query := `
		SELECT
		    t.id,
		    t.name,
		    t.created_at,
		    t.modified_at,
		    t.created_by,
		    t.modified_by
		FROM app.teams t 
		WHERE t.id = ANY($1)
	`

	rows, err := r.db.Pool.Query(ctx, query, teamIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*_model.Team
	for rows.Next() {
		var team _model.Team
		if err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.CreatedAt,
			&team.ModifiedAt,
			&team.CreatedBy,
			&team.ModifiedBy,
		); err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return teams, nil
}

func (r *TeamRepository) GetTeamsByUserID(ctx context.Context, userID string) ([]*_projection.TeamSummary, error) {
	query := `
		WITH count_team_member AS (
			SELECT ut.team_id, count(1) member_count
			FROM app.user_teams ut
			GROUP BY 1
		)
		SELECT
			t.id,
			t."name",
			t.description,
			(SELECT member_count FROM count_team_member WHERE team_id = t.id) AS member_count,
			t.created_at,
			t.modified_at,
			t.created_by,
			t.modified_by
		FROM app.teams t 
		JOIN app.user_teams ut ON t.id = ut.team_id
		WHERE ut.user_id = @userId
	`

	// Query arguments
	args := pgx.NamedArgs{
		"userId": userID,
	}

	var teams []*_projection.TeamSummary
	rows, err := r.db.Pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var row _projection.TeamSummary
		var team _model.Team
		if err = rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&row.MemberCount,
			&team.CreatedAt,
			&team.ModifiedAt,
			&team.CreatedBy,
			&team.ModifiedBy,
		); err != nil {
			return nil, err
		}
		row.Team = &team
		teams = append(teams, &row)
	}
	return teams, nil
}
