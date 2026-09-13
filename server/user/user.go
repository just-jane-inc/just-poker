package user

// UserDTO describes an individual user
type UserDTO struct {
	// the id of the user
	UserID string `json:"user_id"`

	// the users display name
	DisplayName string `json:"display_name"`

	// the type of user
	UserType string `json:"user_type"`

	// the id of the twitch user that owns this poker user
	TwitchID string `json:"twitch_id"`
}
