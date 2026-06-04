package domain

const ProviderPassword = "password"

type AuthIdentity struct {
	UserID       UserID
	Provider     string
	PasswordHash string
}
