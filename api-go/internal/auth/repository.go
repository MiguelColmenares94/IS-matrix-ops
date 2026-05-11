package auth

import (
	"database/sql"
	"time"
)

type User struct {
	ID        string
	Email     string
	Password  string
	CreatedAt time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserByEmail(email string) (*User, error) {
	row := r.db.QueryRow("SELECT * FROM get_user_by_email($1)", email)
	u := &User{}
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) SaveRefreshToken(userID, token string, expiresAt time.Time) error {
	_, err := r.db.Exec("CALL save_refresh_token($1, $2, $3)", userID, token, expiresAt)
	return err
}

func (r *Repository) GetRefreshToken(token string) (*RefreshToken, error) {
	row := r.db.QueryRow("SELECT * FROM get_refresh_token($1)", token)
	rt := &RefreshToken{}
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt); err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *Repository) DeleteRefreshToken(token string) error {
	_, err := r.db.Exec("CALL delete_refresh_token($1)", token)
	return err
}
