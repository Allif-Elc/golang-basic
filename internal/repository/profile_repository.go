package repository

import (
	"context"
	"golang-basic/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(ctx context.Context, profile *model.CreateProfileRequest) (*model.Profile, error) {
	query := `INSERT INTO profiles (user_id, age, gender, bio, phone_number, website) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	var createdProfile model.Profile
	err := r.db.QueryRow(ctx, query, profile.UserID, profile.Age, profile.Gender, profile.Bio, profile.PhoneNumber, profile.Website).Scan(&createdProfile.ProfileId, &createdProfile.CreateAt)
	if err != nil {
		return nil, err
	}
	createdProfile.UserID = profile.UserID
	createdProfile.Age = profile.Age
	createdProfile.Gender = profile.Gender
	createdProfile.Bio = profile.Bio
	createdProfile.PhoneNumber = profile.PhoneNumber
	createdProfile.Website = profile.Website
	return &createdProfile, nil

}
func (r *ProfileRepository) FindByID(ctx context.Context, profileID int64) (model.Profile, error) {
	query := `SELECT id_profile, user_id, age, gender, bio, phone_number, website, created_at, updated_at FROM profiles WHERE id_profile = $1`
	row, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return model.Profile{}, err
	}
	defer row.Close()

	var profile model.Profile
	if row.Next() {
		err = row.Scan(&profile.ProfileId, &profile.UserID, &profile.Age, &profile.Gender, &profile.Bio, &profile.PhoneNumber, &profile.Website, &profile.CreateAt, &profile.UpdateAt)
		if err != nil {
			return model.Profile{}, err
		}
	}
	return profile, nil
}

func (r *ProfileRepository) Update(ctx context.Context, profile model.UpdateProfileRequest) error {
	queryExist := `SELECT id_profile FROM profiles WHERE id_profile = $1`
	row, err := r.db.Query(ctx, queryExist, profile.ProfileId)
	if err != nil {
		return err
	}
	defer row.Close()

	var existingProfile model.Profile
	if row.Next() {
		err = row.Scan(&existingProfile.ProfileId)
		if err != nil {
			return err
		}
	}

	query := `UPDATE profiles SET age = $1, gender = $2, bio = $3, phone_number = $4, website = $5, updated_at = NOW() WHERE id_profile = $6`
	_, err = r.db.Exec(ctx, query, profile.Age, profile.Gender, profile.Bio, profile.PhoneNumber, profile.Website, profile.ProfileId)
	return err
}

func (r *ProfileRepository) FindAll(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error) {
	switch req.Size {
	case 5, 10, 25, 50, 100:
	default:
		req.Size = 10
	}

	if req.Page < 1 {
		req.Page = 1
	}

	query := `SELECT id_profile, user_id, age, gender, bio, phone_number, website, created_at, updated_at FROM profiles WHERE id_profile > $1 ORDER BY id_profile LIMIT $2`
	rows, err := r.db.Query(ctx, query, lastCursor, req.Size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []model.Profile
	var maxID int64 = lastCursor
	for rows.Next() {
		var profile model.Profile
		err = rows.Scan(&profile.ProfileId, &profile.UserID, &profile.Age, &profile.Gender, &profile.Bio, &profile.PhoneNumber, &profile.Website, &profile.CreateAt, &profile.UpdateAt)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
		if int64(profile.ProfileId) > maxID {
			maxID = int64(profile.ProfileId)
		}
	}

	result := &model.PageResult[model.Profile]{
		Data:       profiles,
		Page:       req.Page,
		Size:       req.Size,
		StartRow:   (req.Page-1)*req.Size + 1,
		EndRow:     (req.Page-1)*req.Size + len(profiles),
		NextCursor: maxID,
	}

	return result, nil
}

func (r *ProfileRepository) FindAllWithUsers(ctx context.Context, req model.PageRequest, lastCursor int64) (*model.PageResult[model.Profile], error) {
	switch req.Size {
	case 5, 10, 25, 50, 100:
	default:
		req.Size = 10
	}

	if req.Page < 1 {
		req.Page = 1
	}

	query := `SELECT p.id_profile, p.user_id, p.age, p.gender, p.bio, p.phone_number, p.website, p.created_at, p.updated_at,
			  u.id_user, u.name, u.email, u.is_active, u.created_at, u.updated_at
			  FROM profiles p
			  JOIN users u ON p.user_id = u.id_user
			  WHERE p.id_profile > $1
			  ORDER BY p.id_profile
			  LIMIT $2`
	rows, err := r.db.Query(ctx, query, lastCursor, req.Size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []model.Profile
	var maxID int64 = lastCursor
	for rows.Next() {
		var profile model.Profile
		var user model.User
		err = rows.Scan(&profile.ProfileId, &profile.UserID, &profile.Age, &profile.Gender, &profile.Bio, &profile.PhoneNumber, &profile.Website, &profile.CreateAt, &profile.UpdateAt,
			&user.UserId, &user.Name, &user.Email, &user.IsActive, &user.CreateAt, &user.UpdateAt)
		if err != nil {
			return nil, err
		}
		profile.User = &user
		profiles = append(profiles, profile)
		if int64(profile.ProfileId) > maxID {
			maxID = int64(profile.ProfileId)
		}
	}

	result := &model.PageResult[model.Profile]{
		Data:       profiles,
		Page:       req.Page,
		Size:       req.Size,
		StartRow:   (req.Page-1)*req.Size + 1,
		EndRow:     (req.Page-1)*req.Size + len(profiles),
		NextCursor: maxID,
	}

	return result, nil
}

func (r *ProfileRepository) Delete(ctx context.Context, ProfileId int64) error {
	query := `DELETE FROM profiles WHERE id_profile = $1`
	_, err := r.db.Exec(ctx, query, ProfileId)
	return err
}
