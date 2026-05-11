package auth_test

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/is-matrix-ops/api-go/internal/auth"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) GetUserByEmail(email string) (*auth.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *mockRepo) SaveRefreshToken(userID, token string, expiresAt time.Time) error {
	return m.Called(userID, token, expiresAt).Error(0)
}
func (m *mockRepo) GetRefreshToken(token string) (*auth.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.RefreshToken), args.Error(1)
}
func (m *mockRepo) DeleteRefreshToken(token string) error {
	return m.Called(token).Error(0)
}

func init() {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY_MINUTES", "10")
	os.Setenv("JWT_REFRESH_EXPIRY_DAYS", "7")
}

func hashedPassword(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), 4) // cost 4 for speed in tests
	require.NoError(t, err)
	return string(h)
}

func TestLogin_ValidCredentials(t *testing.T) {
	repo := &mockRepo{}
	svc := auth.NewService(repo)

	hash := hashedPassword(t, "secret")
	user := &auth.User{ID: "uid-1", Email: "a@b.com", Password: hash}
	repo.On("GetUserByEmail", "a@b.com").Return(user, nil)
	repo.On("SaveRefreshToken", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	pair, err := svc.Login("a@b.com", "secret")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.True(t, pair.ExpiresAt.After(time.Now()))
	repo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockRepo{}
	svc := auth.NewService(repo)

	hash := hashedPassword(t, "correct")
	user := &auth.User{ID: "uid-1", Email: "a@b.com", Password: hash}
	repo.On("GetUserByEmail", "a@b.com").Return(user, nil)

	_, err := svc.Login("a@b.com", "wrong")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := auth.NewService(repo)
	repo.On("GetUserByEmail", "x@x.com").Return(nil, sql.ErrNoRows)

	_, err := svc.Login("x@x.com", "any")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestRefresh_ValidToken(t *testing.T) {
	repo := &mockRepo{}
	svc := auth.NewService(repo)

	rt := &auth.RefreshToken{ID: "rt-1", UserID: "uid-1", Token: "old-token", ExpiresAt: time.Now().Add(time.Hour)}
	repo.On("GetRefreshToken", "old-token").Return(rt, nil)
	repo.On("DeleteRefreshToken", "old-token").Return(nil)
	repo.On("SaveRefreshToken", "uid-1", mock.Anything, mock.Anything).Return(nil)

	pair, err := svc.Refresh("old-token")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEqual(t, "old-token", pair.RefreshToken)
	repo.AssertExpectations(t)
}

func TestRefresh_InvalidToken(t *testing.T) {
	repo := &mockRepo{}
	svc := auth.NewService(repo)
	repo.On("GetRefreshToken", "bad-token").Return(nil, sql.ErrNoRows)

	_, err := svc.Refresh("bad-token")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestLogout_DeletesToken(t *testing.T) {
	repo := &mockRepo{}
	svc := auth.NewService(repo)
	repo.On("DeleteRefreshToken", "tok").Return(nil)

	err := svc.Logout("tok")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestLogout_PropagatesError(t *testing.T) {
	repo := &mockRepo{}
	svc := auth.NewService(repo)
	repo.On("DeleteRefreshToken", "tok").Return(errors.New("db error"))

	err := svc.Logout("tok")
	assert.Error(t, err)
}
