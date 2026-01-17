package repository

import (
	"context"
	"golang-basic/api/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.CreateUserRequest) (*model.User, error) {
	query := `INSERT into users (name, email, is_active) VALUES ($1, $2, TRUE) RETURNING id_user, created_at`
	var createdUser model.User
	err := r.db.QueryRow(ctx, query, user.Name, user.Email).Scan(&createdUser.UserId, &createdUser.CreateAt)
	if err != nil {
		return nil, err
	}
	createdUser.Name = user.Name
	createdUser.Email = user.Email
	return &createdUser, nil
}

func (r *UserRepository) FindByID(ctx context.Context, userID int64) (model.User, error) {
	query := `SELECT id_user, name, email, created_at, updated_at FROM users WHERE id_user = $1`
	row, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return model.User{}, err
	}
	defer row.Close()

	var user model.User
	if row.Next() {
		err = row.Scan(&user.UserId, &user.Name, &user.Email, &user.CreateAt, &user.UpdateAt)
		if err != nil {
			return model.User{}, err
		}
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user model.UpdateUserRequest) error {
	querySelect := `SELECT id_user, name, email, created_at, updated_at FROM users WHERE id_user = $1`
	row, err := r.db.Query(ctx, querySelect, user.UserId)
	if err != nil {
		return err
	}
	defer row.Close()

	var existingUser model.User
	if row.Next() {
		err = row.Scan(&existingUser.UserId, &existingUser.Name, &existingUser.Email, &existingUser.CreateAt, &existingUser.UpdateAt)
		if err != nil {
			return err
		}
	}

	query := `UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id_user = $3`
	_, err = r.db.Exec(ctx, query, user.Name, user.Email, user.UserId)
	return err
}

func (r *UserRepository) FindAll(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.User], error) {
	switch req.Size {
	case 5, 10, 25, 50, 100:
	default:
		req.Size = 10
	}

	if req.Page < 1 {
		req.Page = 1
	}

	query := `SELECT id_user, name, email, created_at, updated_at FROM users WHERE id_user > $1 ORDER BY id_user LIMIT $2`
	rows, err := r.db.Query(ctx, query, lastCursor, req.Size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	var maxID int64 = lastCursor
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.UserId, &user.Name, &user.Email, &user.CreateAt, &user.UpdateAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
		if int64(user.UserId) > maxID {
			maxID = int64(user.UserId)
		}
	}

	result := &model.PageResult[model.User]{
		Data:       users,
		Page:       req.Page,
		Size:       req.Size,
		StartRow:   (req.Page-1)*req.Size + 1,
		EndRow:     (req.Page-1)*req.Size + len(users),
		NextCursor: maxID,
	}

	return result, nil
}
