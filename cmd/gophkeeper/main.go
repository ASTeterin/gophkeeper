package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	grpcPkg "google.golang.org/grpc"

	pb "github.com/ASTeterin/gophkeeper/api"
	"github.com/ASTeterin/gophkeeper/internal/cookie"
	"github.com/ASTeterin/gophkeeper/internal/grpc"
	"github.com/ASTeterin/gophkeeper/internal/handler"
	db "github.com/ASTeterin/gophkeeper/internal/repository"
	"github.com/ASTeterin/gophkeeper/internal/service"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"

	appConfig "github.com/ASTeterin/gophkeeper/internal/config"
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
	g, ctx := errgroup.WithContext(ctx)
	defer cancel()

	userRepo := db.NewUserRepository(dbConn)
	dataRepo := db.NewPrivateDataRepo(dbConn)
	userService := service.NewUserService(userRepo)
	dataService := service.NewPrivateDataService(dataRepo)

	uh := handler.NewUserHandler(userService)
	dh := handler.NewPrivateDataHandler(dataService)

	r := gin.Default()
	r.Use(cookie.CookieHandler(config.SigningKey))

	r.POST("/api/user/register", func(c *gin.Context) {
		uh.Register(c)
	})
	r.POST("/api/user/login", func(c *gin.Context) {
		uh.Authenticate(c)
	})
	r.POST("/api/data", func(c *gin.Context) {
		dh.Store(c)
	})
	r.GET("/api/data", func(c *gin.Context) {
		dh.GetAll(c)
	})
	r.GET("/api/data/:key", func(c *gin.Context) {
		dh.GetByKey(c)
	})
	r.DELETE("/api/data/:id", func(c *gin.Context) {
		dh.Delete(c)
	})
	r.POST("/api/data/sync", func(c *gin.Context) {
		dh.ReplaceAll(c)
	})

	grpcServer := grpcPkg.NewServer()
	grpcDataSvc := grpc.NewPrivateDataGRPCServer(dataService)
	pb.RegisterPrivateDataServiceServer(grpcServer, grpcDataSvc)

	grpcUserSvc := grpc.NewUserGRPCServer(userService)
	pb.RegisterUserServiceServer(grpcServer, grpcUserSvc)

	host, _, err := net.SplitHostPort(config.AppAddr)
	if err != nil {
		host = ""
	}
	grpcAddr := net.JoinHostPort(host, config.GRPCAddr)

	g.Go(func() error {
		log.Printf("HTTP server listening on %s", config.AppAddr)
		httpLis, err := net.Listen("tcp", config.AppAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", config.AppAddr, err)
		}

		go func() {
			if err := r.RunListener(httpLis); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}()

		<-ctx.Done()
		return nil
	})

	g.Go(func() error {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
		}
		log.Printf("gRPC server listening on %s", grpcAddr)

		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
		<-ctx.Done()
		grpcServer.GracefulStop()
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("Servers failed: %v", err)
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
