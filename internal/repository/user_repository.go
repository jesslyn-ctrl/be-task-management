package repository

import (
	_db "bitbucket.org/edts/go-task-management/internal/db"
	_model "bitbucket.org/edts/go-task-management/internal/model"
	"context"
	"github.com/jackc/pgx/v5"
)

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *_model.User) (*_model.User, error)
	GetUserByID(ctx context.Context, id string) (*_model.User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*_model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*_model.User, error)
	//GetAllUser(ctx context.Context) ([]*_model.User, error)
}

type UserRepository struct {
	db *_db.Database
}

func NewUserRepository(db *_db.Database) UserRepositoryInterface {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *_model.User) (*_model.User, error) {
	query := `
		INSERT INTO app.users ("name", email, "password", created_at, modified_at, created_by, modified_by)
		VALUES (@name, @email, @password, current_timestamp, current_timestamp, @created_by, @modified_by)
		RETURNING id, created_at	
	`

	// Query arguments
	args := pgx.NamedArgs{
		"name":        user.Name,
		"email":       user.Email,
		"password":    user.Password,
		"created_by":  user.CreatedBy,
		"modified_by": user.ModifiedBy,
	}

	err := r.db.Pool.QueryRow(ctx, query, args).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*_model.User, error) {
	query := `
		SELECT id, "name", email, created_at, modified_at, created_by, modified_by
		FROM app.users
		WHERE id = @id
	`

	// Query arguments
	args := pgx.NamedArgs{
		"id": id,
	}

	var user _model.User
	scan := func(row pgx.Row) error {
		return row.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.ModifiedAt,
			&user.CreatedBy,
			&user.ModifiedBy,
		)
	}

	err := scan(r.db.Pool.QueryRow(ctx, query, args))
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUsersByIDs(ctx context.Context, ids []string) ([]*_model.User, error) {
	if len(ids) == 0 {
		return []*_model.User{}, nil
	}

	query := `
		SELECT id, "name", email, created_at, modified_at, created_by, modified_by
		FROM app.users
		WHERE id = ANY($1)
	`

	rows, err := r.db.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*_model.User
	for rows.Next() {
		var user _model.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.ModifiedAt,
			&user.CreatedBy,
			&user.ModifiedBy,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*_model.User, error) {
	query := `
		SELECT id, "name", email, created_at, modified_at, created_by, modified_by, password
		FROM app.users
		WHERE LOWER(email) = LOWER(@email)
	`

	// Query arguments
	args := pgx.NamedArgs{
		"email": email,
	}

	var user _model.User
	scan := func(row pgx.Row) error {
		return row.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.ModifiedAt,
			&user.CreatedBy,
			&user.ModifiedBy,
			&user.Password,
		)
	}

	err := scan(r.db.Pool.QueryRow(ctx, query, args))
	if err != nil {
		return nil, err
	}

	return &user, nil
}

//TODO: GetAllUser Repository Method
//1. Fetch all user from app.users table
