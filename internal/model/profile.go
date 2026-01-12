package model

type Profile struct {
	ProfileId   int64  `json:"id_profile"`
	UserID      int64  `json:"id_user"`
	Age         uint8  `json:"age"`
	Gender      string `json:"gender"`
	Bio         string `json:"bio"`
	PhoneNumber string `json:"phone_number"`
	Website     string `json:"website"`
	CreateAt    string `json:"created_at"`
	UpdateAt    string `json:"updated_at"`
	User        *User  `json:"user,omitempty"`
}

type CreateProfileRequest struct {
	UserID      int64  `json:"id_user"`
	Age         uint8  `json:"age"`
	Gender      string `json:"gender"`
	Bio         string `json:"bio"`
	PhoneNumber string `json:"phone_number"`
	Website     string `json:"website"`
}

type UpdateProfileRequest struct {
	ProfileId   int64  `json:"id_profile"`
	Age         uint8  `json:"age"`
	Gender      string `json:"gender"`
	Bio         string `json:"bio"`
	PhoneNumber string `json:"phone_number"`
	Website     string `json:"website"`
}
