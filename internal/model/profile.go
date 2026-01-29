package model

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Profile struct {
	ProfileId   int64       `json:"id_profile"`
	UserID      int64       `json:"id_user"`
	Age         pgtype.Int8 `json:"age"`           // nullable: uses pgtype for NULL handling
	Gender      pgtype.Text  `json:"gender"`        // nullable: uses pgtype for NULL handling
	Bio         pgtype.Text  `json:"bio"`           // nullable: uses pgtype for NULL handling
	PhoneNumber pgtype.Text  `json:"phonenumber"`    // nullable: uses pgtype for NULL handling
	Website     pgtype.Text  `json:"website"`        // nullable: uses pgtype for NULL handling
	CreateAt    time.Time    `json:"created_at"`
	UpdateAt    time.Time    `json:"updated_at"`
	User        *User        `json:"user,omitempty"`
}

// CreateProfileRequest represents a request to create a profile
// All fields are optional (use pointers for nullable fields)
type CreateProfileRequest struct {
	UserID      int64   `json:"id_user"`
	Age         *int8   `json:"age,omitempty"`         // nullable
	Gender      *string `json:"gender,omitempty"`      // nullable
	Bio         *string `json:"bio,omitempty"`         // nullable
	PhoneNumber *string `json:"phonenumber,omitempty"`  // nullable
	Website     *string `json:"website,omitempty"`      // nullable
}

// UpdateProfileRequest represents a request to update a profile
// All fields are optional (use pointers for nullable fields)
type UpdateProfileRequest struct {
	ProfileId   int64   `json:"id_profile"`
	Age         *int8   `json:"age,omitempty"`         // nullable
	Gender      *string `json:"gender,omitempty"`      // nullable
	Bio         *string `json:"bio,omitempty"`         // nullable
	PhoneNumber *string `json:"phonenumber,omitempty"`  // nullable
	Website     *string `json:"website,omitempty"`      // nullable
}
