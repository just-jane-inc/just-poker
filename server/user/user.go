package user

type UserDTO struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	UserType    string `json:"user_type"`
	TwitchID    string `json:"twitch_id"`
}
