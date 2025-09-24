package repository

import (
	_db "bitbucket.org/edts/go-task-management/internal/db"
	_model "bitbucket.org/edts/go-task-management/internal/model"
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

type TaskRepositoryInterface interface {
	CreateTask(ctx context.Context, task *_model.Task) (*_model.Task, error)
	GetTasksByTeam(ctx context.Context, teamID string, status *string) ([]*_model.Task, error)
	GetTaskByID(ctx context.Context, id string) (*_model.Task, error)
	UpdateTaskById(ctx context.Context, task *_model.Task) (*_model.Task, error)
	DeleteTaskById(ctx context.Context, taskID string) error
	MoveTaskById(ctx context.Context, task *_model.Task) (*_model.Task, error)
	AssignTask(ctx context.Context, task *_model.Task) (*_model.Task, error)
}

type TaskRepository struct {
	db *_db.Database
}

func NewTaskRepository(db *_db.Database) TaskRepositoryInterface {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) CreateTask(ctx context.Context, task *_model.Task) (*_model.Task, error) {
	query := `
		INSERT INTO app.tasks (title, description, status, assigned_to, team_id, due_date, created_at, modified_at, created_by, modified_by)
		VALUES (@title, @description, @status, @assigned_to, @team_id, @due_date, current_timestamp, current_timestamp, @created_by, @modified_by)
		RETURNING id, created_at
	`

	// Query arguments
	args := pgx.NamedArgs{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"assigned_to": task.AssignedTo,
		"team_id":     task.TeamID,
		"due_date":    task.DueDate,
		"created_by":  task.CreatedBy,
		"modified_by": task.ModifiedBy,
	}

	err := r.db.Pool.QueryRow(ctx, query, args).Scan(&task.ID, &task.CreatedAt)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TaskRepository) GetTasksByTeam(ctx context.Context, teamID string, status *string) ([]*_model.Task, error) {
	query := `
		SELECT 
			t.id,
			t.title,
			t.description,
			t.status,
			t.due_date,
			t.assigned_to,
			t.team_id,
			t.created_at,
			t.modified_at
		FROM app.tasks t 
		WHERE t.team_id = @team_id
		AND (@status::varchar IS NULL OR t.status = @status::varchar)
		ORDER BY t.created_at DESC
	`

	// Query arguments
	args := pgx.NamedArgs{
		"team_id": teamID,
		"status":  status,
	}

	var tasks []*_model.Task
	rows, err := r.db.Pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task _model.Task

		if err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.DueDate,
			&task.AssignedTo,
			&task.TeamID,
			&task.CreatedAt,
			&task.ModifiedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *TaskRepository) GetTaskByID(ctx context.Context, id string) (*_model.Task, error) {
	query := `
		SELECT 
		    ts.id,
			ts.title,
			ts.description,
			ts.status,
			ts.due_date,
			ts.assigned_to,
			ts.team_id,
			ts.created_at,
			ts.modified_at,
			u.id AS user_id,
			u."name" AS user_name,
			u.email,
			u.created_at AS user_created_at,
			t.id AS team_id,
			t."name" AS team_name,
			t.description AS team_desc,
			t.created_at AS team_created_at
		FROM app.tasks ts 
		JOIN app.teams t ON t.id = ts.team_id
		LEFT JOIN app.users u ON u.id = ts.assigned_to
		WHERE ts.id = @id
	`

	// Query arguments
	args := pgx.NamedArgs{
		"id": id,
	}

	var task _model.Task
	var assignedUser *_model.User
	var team _model.Team

	// Use pointer types for the fields that can be nil
	var userId *string
	var userName *string
	var userEmail *string
	var userCreatedAt *time.Time

	if err := r.db.Pool.QueryRow(ctx, query, args).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.DueDate,
		&task.AssignedTo,
		&task.TeamID,
		&task.CreatedAt,
		&task.ModifiedAt,
		&userId,
		&userName,
		&userEmail,
		&userCreatedAt,
		&team.ID,
		&team.Name,
		&team.Description,
		&team.CreatedAt,
	); err != nil {
		return nil, err
	}

	// Assign the user data if its not nil
	if userId != nil && userName != nil && userEmail != nil && userCreatedAt != nil {
		assignedUser = &_model.User{
			ID:    *userId,
			Name:  *userName,
			Email: *userEmail,
			Base: _model.Base{
				CreatedAt: *userCreatedAt,
			},
		}
	} else {
		assignedUser = nil
	}
	// Add the assigned user
	task.AssignedUser = assignedUser
	// Add the team
	task.Team = &team

	return &task, nil
}

func (r *TaskRepository) UpdateTaskById(ctx context.Context, task *_model.Task) (*_model.Task, error) {
	sqlStatement := `
		UPDATE app.tasks
		SET title = $1, description = $2, due_date = $3, modified_at = current_timestamp
		WHERE id = $4
		RETURNING id, title, description, status, assigned_to, due_date;
	`

	var updatedTask _model.Task
	err := r.db.Pool.QueryRow(
		ctx,
		sqlStatement,
		task.Title,
		task.Description,
		task.DueDate,
		task.ID,
	).Scan(
		&updatedTask.ID,
		&updatedTask.Title,
		&updatedTask.Description,
		&updatedTask.Status,
		&updatedTask.AssignedTo,
		&updatedTask.DueDate,
	)

	if err != nil {
		return nil, err
	}

	return &updatedTask, nil
}

func (r *TaskRepository) DeleteTaskById(ctx context.Context, taskID string) error {
	query := `
		DELETE FROM app.tasks WHERE id = $1;
	`

	_, err := r.db.Pool.Exec(ctx, query, taskID)
	if err != nil {
		return err
	}
	return nil

}

func (r *TaskRepository) MoveTaskById(ctx context.Context, task *_model.Task) (*_model.Task, error) {
	//TODO: update task status by task id
	sqlStatement := `
		UPDATE app.tasks
		SET title = $1, modified_at = current_timestamp
		WHERE id = $2
		RETURNING id, title, description, status, assigned_to, due_date;
	`

	var moveTask _model.Task
	err := r.db.Pool.QueryRow(
		ctx,
		sqlStatement,
		task.Status,
		task.ID,
	).Scan(
		&moveTask.ID,
		&moveTask.Title,
		&moveTask.Description,
		&moveTask.Status,
		&moveTask.AssignedTo,
		&moveTask.DueDate,
	)

	if err != nil {
		return nil, err
	}

	return &moveTask, nil
}

func (r *TaskRepository) AssignTask(ctx context.Context, task *_model.Task) (*_model.Task, error) {
	sqlStatement := `
		UPDATE app.tasks
		SET assigned_to = $1, modified_at = current_timestamp
		WHERE id = $2
		RETURNING id, title, description, status, assigned_to, due_date;
	`

	var assignedTask _model.Task
	err := r.db.Pool.QueryRow(
		ctx,
		sqlStatement,
		task.AssignedTo,
		task.ID,
	).Scan(
		&assignedTask.ID,
		&assignedTask.Title,
		&assignedTask.Description,
		&assignedTask.Status,
		&assignedTask.AssignedTo,
		&assignedTask.DueDate,
	)

	if err != nil {
		return nil, err
	}

	return &assignedTask, nil
}
