package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	AppAddr    string
	GRPCAddr   string
	DBConnStr  string
	SigningKey string
	CertDir    string
	ServerName string
}

const (
	defaultCertDir    string = "./../../cert/"
	defaultServerName string = "localhost"
)

func ParseFlags() Config {
	var appAddr string
	var dbConnectionString string
	var signingKey string
	var grpcAddr string
	var certDir = defaultCertDir
	var serverName = defaultServerName

	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&grpcAddr, "g", ":8081", "port to run grpc server")
	flag.StringVar(&dbConnectionString, "d", "postgres://admin:1234@localhost:5432/gophkeeper?sslmode=disable", "database DSN")
	flag.Parse()

	if a, exist := os.LookupEnv("RUN_ADDRESS"); exist {
		appAddr = a
	}
	if a, exist := os.LookupEnv("GRPC_ADDRESS"); exist {
		grpcAddr = a
	}
	if d, exist := os.LookupEnv("DATABASE_URI"); exist {
		dbConnectionString = d
	}
	if k, exist := os.LookupEnv("SIGNING_KEY"); exist {
		signingKey = k
	}
	if d, exist := os.LookupEnv("CERT_DIR"); exist {
		certDir = d
	}
	if s, exist := os.LookupEnv("SERVER_NAME"); exist {
		serverName = s
	}

	if signingKey == "" {
		log.Fatal("SIGNING_KEY environment variable is required")
	}

	return Config{
		AppAddr:    appAddr,
		DBConnStr:  dbConnectionString,
		SigningKey: signingKey,
		GRPCAddr:   grpcAddr,
		CertDir:    certDir,
		ServerName: serverName,
	}
}

func ParseClientFlags() Config {
	var appAddr string
	var grpcAddr string
	var certDir = defaultCertDir
	var serverName = defaultServerName

	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&grpcAddr, "g", ":8081", "port to run grpc server")
	flag.Parse()

	if a, exist := os.LookupEnv("RUN_ADDRESS"); exist {
		appAddr = a
	}
	if g, exist := os.LookupEnv("GRPC_ADDRESS"); exist {
		grpcAddr = g
	}
	if d, exist := os.LookupEnv("CERT_DIR"); exist {
		certDir = d
	}
	if s, exist := os.LookupEnv("SERVER_NAME"); exist {
		serverName = s
	}

	return Config{
		AppAddr:    appAddr,
		GRPCAddr:   grpcAddr,
		CertDir:    certDir,
		ServerName: serverName,
	}
}
