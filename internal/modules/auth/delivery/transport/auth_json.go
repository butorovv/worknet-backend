package transport

import "github.com/butorovv/worknet-backend/internal/modules/auth/usecase"

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

func ToAuthResponse(result usecase.AuthResult) AuthResponse {
	return AuthResponse{User: ToUserResponse(result.User)}
}

func ToUserResponse(user usecase.CurrentUser) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
	}
}
