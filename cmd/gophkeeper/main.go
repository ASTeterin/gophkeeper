package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ASTeterin/loyalty/internal/cookie"
	"github.com/ASTeterin/loyalty/internal/handler"
	db "github.com/ASTeterin/loyalty/internal/repository"
	"github.com/ASTeterin/loyalty/internal/service"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"path/filepath"
	"time"

	appConfig "github.com/ASTeterin/loyalty/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	config := appConfig.ParseFlags()
	var dbConn *sql.DB
	dbConn, err := sql.Open("pgx", config.DBConnStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	migrateDB(dbConn)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	repo := db.NewUserRepository(dbConn)
	userService := service.NewUserService(repo)

	h := handler.NewHandler(userService)

	r := gin.Default()
	r.Use(cookie.CookieHandler(config.SigningKey))

	r.POST("/api/user/register", func(c *gin.Context) {
		h.Register(ctx, c)
	})
	r.POST("/api/user/login", func(c *gin.Context) {
		h.Authenticate(ctx, c)
	})
	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func migrateDB(conn *sql.DB) {
	driver, err := postgres.WithInstance(conn, &postgres.Config{
		SchemaName: "public",
	})
	if err != nil {
		log.Fatal(err)
	}

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	migrationsPath := filepath.Join(exeDir, "..", "..", "migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Fatalf("Migrations directory not found: %s", migrationsPath)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	}
}
