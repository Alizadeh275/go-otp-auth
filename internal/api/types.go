package api

// UserResponse represents the user data returned by the API
type UserResponse struct {
	ID           int64  `json:"id"`
	Phone        string `json:"phone"`
	RegisteredAt string `json:"registered_at"`
}
