package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/is-matrix-ops/api-go/internal/auth"
	"github.com/is-matrix-ops/api-go/internal/matrix"
	pkgdb "github.com/is-matrix-ops/api-go/pkg/db"
	"github.com/is-matrix-ops/api-go/pkg/middleware"
)

func main() {
	_ = godotenv.Load()

	db, err := pkgdb.NewPool()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	matrixRepo := matrix.NewRepository(db)
	matrixSvc := &matrix.Service{}
	matrixHandler := matrix.NewHandler(matrixSvc, matrixRepo)

	app := fiber.New()
	app.Use(middleware.Recovery())
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("ALLOWED_ORIGIN"),
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	v1 := app.Group("/api/v1")
	v1.Post("/auth/login", authHandler.Login)
	v1.Post("/auth/refresh", authHandler.Refresh)
	v1.Post("/auth/logout", middleware.JWT(), authHandler.Logout)
	v1.Post("/matrix/qr", middleware.JWT(), matrixHandler.ComputeQR)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
