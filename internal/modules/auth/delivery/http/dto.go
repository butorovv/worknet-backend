package authhttp

import "github.com/butorovv/worknet-backend/internal/modules/auth/usecase"

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User userResponse `json:"user"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

func authResultResponse(result usecase.AuthResult) authResponse {
	return authResponse{User: userResponseFromCurrent(result.User)}
}

func userResponseFromCurrent(user usecase.CurrentUser) userResponse {
	return userResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
	}
}
