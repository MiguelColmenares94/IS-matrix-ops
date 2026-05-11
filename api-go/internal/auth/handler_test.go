package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/is-matrix-ops/api-go/internal/auth"
	pkgdb "github.com/is-matrix-ops/api-go/pkg/db"
	"github.com/is-matrix-ops/api-go/pkg/middleware"
)

func setupApp(t *testing.T) *fiber.App {
	t.Helper()
	db, err := pkgdb.NewPool()
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Exec("TRUNCATE refresh_tokens CASCADE")
		db.Close()
	})

	repo := auth.NewRepository(db)
	svc := auth.NewService(repo)
	h := auth.NewHandler(svc)

	app := fiber.New()
	app.Post("/api/v1/auth/login", h.Login)
	app.Post("/api/v1/auth/refresh", h.Refresh)
	app.Post("/api/v1/auth/logout", middleware.JWT(), h.Logout)
	return app
}

func post(app *fiber.App, path string, body interface{}, token string) *http.Response {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

func TestLogin_Success(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app := setupApp(t)
	resp := post(app, "/api/v1/auth/login", map[string]string{
		"email":    os.Getenv("SEED_EMAIL"),
		"password": os.Getenv("SEED_PASSWORD"),
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var pair map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pair)
	assert.NotEmpty(t, pair["access_token"])
	assert.NotEmpty(t, pair["refresh_token"])
}

func TestLogin_WrongPassword(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app := setupApp(t)
	resp := post(app, "/api/v1/auth/login", map[string]string{
		"email":    os.Getenv("SEED_EMAIL"),
		"password": "wrongpassword",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_MissingFields(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app := setupApp(t)
	resp := post(app, "/api/v1/auth/login", map[string]string{"email": "x@x.com"}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRefresh_Success(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app := setupApp(t)
	loginResp := post(app, "/api/v1/auth/login", map[string]string{
		"email":    os.Getenv("SEED_EMAIL"),
		"password": os.Getenv("SEED_PASSWORD"),
	}, "")
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	var pair map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&pair)

	resp := post(app, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": pair["refresh_token"].(string),
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRefresh_InvalidToken(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app := setupApp(t)
	resp := post(app, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "00000000-0000-0000-0000-000000000000",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogout_ThenRefreshFails(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app := setupApp(t)
	loginResp := post(app, "/api/v1/auth/login", map[string]string{
		"email":    os.Getenv("SEED_EMAIL"),
		"password": os.Getenv("SEED_PASSWORD"),
	}, "")
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	var pair map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&pair)

	logoutResp := post(app, "/api/v1/auth/logout", map[string]string{
		"refresh_token": pair["refresh_token"].(string),
	}, pair["access_token"].(string))
	assert.Equal(t, http.StatusNoContent, logoutResp.StatusCode)

	refreshResp := post(app, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": pair["refresh_token"].(string),
	}, "")
	assert.Equal(t, http.StatusUnauthorized, refreshResp.StatusCode)
}
