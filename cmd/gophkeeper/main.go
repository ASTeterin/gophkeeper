package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"google.golang.org/grpc/credentials"

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

	dbConn, err := sql.Open("pgx", config.DBConnStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	if err := migrateDB(dbConn); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Received shutdown signal...")
		cancel()
	}()

	userRepo := db.NewUserRepository(dbConn)
	dataRepo := db.NewPrivateDataRepo(dbConn)
	userService := service.NewUserService(userRepo)
	dataService := service.NewPrivateDataService(dataRepo)

	uh := handler.NewUserHandler(userService)
	dh := handler.NewPrivateDataHandler(dataService)

	r := initRouter(uh, dh, config.SigningKey)

	creds, err := credentials.NewServerTLSFromFile(config.CertDir+"cert.pem", config.CertDir+"key.pem")
	if err != nil {
		log.Fatalf("Failed to generate credentials: %v", err)
	}

	grpcServer := grpcPkg.NewServer(grpcPkg.Creds(creds))
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
		httpLis, err := net.Listen("tcp", config.AppAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", config.AppAddr, err)
		}
		log.Printf("HTTP server listening on %s", config.AppAddr)

		srv := &http.Server{
			Handler: r,
		}

		go func() {
			if err := srv.Serve(httpLis); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP server error: %v", err)
			}
		}()
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
			return err
		}
		log.Println("HTTP server stopped gracefully")
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
		log.Println("gRPC server stopped gracefully")
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("Servers failed: %v", err)
	}

	log.Println("Server exited successfully")
}

func migrateDB(conn *sql.DB) error {
	driver, err := postgres.WithInstance(conn, &postgres.Config{
		SchemaName: "public",
	})
	if err != nil {
		return err
	}

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	migrationsPath := filepath.Join(exeDir, "..", "..", "migrations")

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("Migrations directory not found: %s", migrationsPath)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func limitBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

func initRouter(uh handler.UserHandler, dh handler.PrivateDataHandler, sk string) *gin.Engine {
	r := gin.Default()
	r.Use(cookie.CookieHandler(sk))
	r.Use(limitBodySize(10 * 1024 * 1024))

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
	r.DELETE("/api/data/:key", func(c *gin.Context) {
		dh.Delete(c)
	})
	r.POST("/api/data/sync", func(c *gin.Context) {
		dh.ReplaceAll(c)
	})

	return r
}
