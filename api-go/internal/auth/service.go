package auth

import (
	"database/sql"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid or expired refresh token")

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(email, password string) (*TokenPair, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokenPair(user.ID, user.Email)
}

func (s *Service) Refresh(refreshToken string) (*TokenPair, error) {
	rt, err := s.repo.GetRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if err := s.repo.DeleteRefreshToken(rt.Token); err != nil {
		return nil, err
	}
	return s.issueTokenPair(rt.UserID, "")
}

func (s *Service) Logout(refreshToken string) error {
	return s.repo.DeleteRefreshToken(refreshToken)
}

func (s *Service) issueTokenPair(userID, email string) (*TokenPair, error) {
	expiryMinutes := 10
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expiryMinutes = n
		}
	}
	expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	refreshDays := 7
	if v := os.Getenv("JWT_REFRESH_EXPIRY_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			refreshDays = n
		}
	}
	refreshExpiry := time.Now().Add(time.Duration(refreshDays) * 24 * time.Hour)
	newRefreshToken, err := generateUUID()
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveRefreshToken(userID, newRefreshToken, refreshExpiry); err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: accessToken, RefreshToken: newRefreshToken, ExpiresAt: expiresAt}, nil
}
