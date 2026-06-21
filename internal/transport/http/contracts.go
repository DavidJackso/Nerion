package thttp

import "go-template/internal/entity"

// Auth

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// User

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func toUserResponse(u *entity.User) userResponse {
	return userResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  string(u.Role),
	}
}
