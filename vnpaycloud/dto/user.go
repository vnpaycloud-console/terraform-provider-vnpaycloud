package dto

// User matches the backend User proto message. An org user; its id is the portal user id.
type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}

// ListUsersResponse matches the backend ListUsersResponse proto message.
type ListUsersResponse struct {
	Users []User `json:"users"`
}
