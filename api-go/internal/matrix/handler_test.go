package matrix_test

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
	"github.com/is-matrix-ops/api-go/internal/matrix"
	pkgdb "github.com/is-matrix-ops/api-go/pkg/db"
	"github.com/is-matrix-ops/api-go/pkg/middleware"
)

func setupMatrixApp(t *testing.T) (*fiber.App, string) {
	t.Helper()
	db, err := pkgdb.NewPool()
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Exec("TRUNCATE qr_computations, refresh_tokens CASCADE")
		db.Close()
	})

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	matrixRepo := matrix.NewRepository(db)
	matrixSvc := &matrix.Service{}
	matrixHandler := matrix.NewHandler(matrixSvc, matrixRepo)

	app := fiber.New()
	app.Post("/api/v1/auth/login", authHandler.Login)
	app.Post("/api/v1/matrix/qr", middleware.JWT(), matrixHandler.ComputeQR)

	// Get a valid token
	b, _ := json.Marshal(map[string]string{
		"email":    os.Getenv("SEED_EMAIL"),
		"password": os.Getenv("SEED_PASSWORD"),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	var pair map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pair)
	token, _ := pair["access_token"].(string)

	return app, token
}

func postMatrix(app *fiber.App, body interface{}, token string) *http.Response {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/qr", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

func TestQR_ValidMatrix(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app, token := setupMatrixApp(t)
	resp := postMatrix(app, map[string]interface{}{
		"matrix": [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.NotNil(t, result["q"])
	assert.NotNil(t, result["r"])
}

func TestQR_NoJWT(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app, _ := setupMatrixApp(t)
	resp := postMatrix(app, map[string]interface{}{
		"matrix": [][]int{{1, 2}, {3, 4}},
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestQR_JaggedMatrix(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	app, token := setupMatrixApp(t)
	resp := postMatrix(app, map[string]interface{}{
		"matrix": []interface{}{[]int{1, 2}, []int{3}},
	}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
