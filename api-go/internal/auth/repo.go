package auth

import "time"

// Repo is the interface the Service depends on, enabling unit testing with mocks.
type Repo interface {
	GetUserByEmail(email string) (*User, error)
	SaveRefreshToken(userID, token string, expiresAt time.Time) error
	GetRefreshToken(token string) (*RefreshToken, error)
	DeleteRefreshToken(token string) error
}
