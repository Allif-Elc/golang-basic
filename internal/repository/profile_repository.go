package repository

import (
	"context"
	"golang-basic/api/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Create(ctx context.Context, profile *model.CreateProfileRequest) (*model.Profile, error) {
	query := `
		INSERT INTO profiles (id_user, age, gender, bio, phonenumber, website)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id_user) DO UPDATE SET
			age = EXCLUDED.age,
			gender = EXCLUDED.gender,
			bio = EXCLUDED.bio,
			phonenumber = EXCLUDED.phonenumber,
			website = EXCLUDED.website,
			updated_at = NOW()
		RETURNING id_profile, created_at, updated_at`
	var createdProfile model.Profile
	err := r.db.QueryRow(ctx, query, profile.UserID, profile.Age, profile.Gender, profile.Bio, profile.PhoneNumber, profile.Website).Scan(&createdProfile.ProfileId, &createdProfile.CreateAt, &createdProfile.UpdateAt)
	if err != nil {
		return nil, err
	}
	createdProfile.UserID = profile.UserID
	// Set pgtype fields from pointer values
	if profile.Age != nil {
		createdProfile.Age.Int64 = int64(*profile.Age)
		createdProfile.Age.Valid = true
	}
	if profile.Gender != nil {
		createdProfile.Gender.String = *profile.Gender
		createdProfile.Gender.Valid = true
	}
	if profile.Bio != nil {
		createdProfile.Bio.String = *profile.Bio
		createdProfile.Bio.Valid = true
	}
	if profile.PhoneNumber != nil {
		createdProfile.PhoneNumber.String = *profile.PhoneNumber
		createdProfile.PhoneNumber.Valid = true
	}
	if profile.Website != nil {
		createdProfile.Website.String = *profile.Website
		createdProfile.Website.Valid = true
	}
	return &createdProfile, nil
}

func (r *ProfileRepository) FindByID(ctx context.Context, profileID int64) (model.Profile, error) {
	query := `SELECT p.id_profile, p.id_user, p.age, p.gender, p.bio, p.phonenumber, p.website, p.created_at, p.updated_at,
			  u.id_user, u.name, u.email, u.is_active, u.created_at, u.updated_at
			  FROM profiles p
			  JOIN users u ON p.id_user = u.id_user
			  WHERE p.id_profile = $1`
	var profile model.Profile
	var user model.User
	err := r.db.QueryRow(ctx, query, profileID).Scan(
		&profile.ProfileId, &profile.UserID, &profile.Age, &profile.Gender, &profile.Bio, &profile.PhoneNumber, &profile.Website, &profile.CreateAt, &profile.UpdateAt,
		&user.UserId, &user.Name, &user.Email, &user.IsActive, &user.CreateAt, &user.UpdateAt)
	if err != nil {
		return model.Profile{}, err
	}
	profile.User = &user
	return profile, nil
}

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID int64) (model.Profile, error) {
	query := `SELECT p.id_profile, p.id_user, p.age, p.gender, p.bio, p.phonenumber, p.website, p.created_at, p.updated_at,
			  u.id_user, u.name, u.email, u.is_active, u.created_at, u.updated_at
			  FROM profiles p
			  JOIN users u ON p.id_user = u.id_user
			  WHERE p.id_user = $1`
	var profile model.Profile
	var user model.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.ProfileId, &profile.UserID, &profile.Age, &profile.Gender, &profile.Bio, &profile.PhoneNumber, &profile.Website, &profile.CreateAt, &profile.UpdateAt,
		&user.UserId, &user.Name, &user.Email, &user.IsActive, &user.CreateAt, &user.UpdateAt)
	if err != nil {
		return model.Profile{}, err
	}
	profile.User = &user
	return profile, nil
}

func (r *ProfileRepository) Update(ctx context.Context, profile model.UpdateProfileRequest) error {
	query := `UPDATE profiles SET age = $1, gender = $2, bio = $3, phonenumber = $4, website = $5, updated_at = NOW() WHERE id_profile = $6`
	_, err := r.db.Exec(ctx, query, profile.Age, profile.Gender, profile.Bio, profile.PhoneNumber, profile.Website, profile.ProfileId)
	return err
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

	query := `SELECT p.id_profile, p.id_user, p.age, p.gender, p.bio, p.phonenumber, p.website, p.created_at, p.updated_at,
			  u.id_user, u.name, u.email, u.is_active, u.created_at, u.updated_at
			  FROM profiles p
			  JOIN users u ON p.id_user = u.id_user
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
