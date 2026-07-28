package user

type UserDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	UserType    string `json:"user_type"`
}
