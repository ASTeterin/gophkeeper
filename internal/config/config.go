package config

import (
	"flag"
	"os"
)

type Config struct {
	AppAddr    string
	GRPCAddr   string
	DBConnStr  string
	SigningKey string
}

const (
	defaultSigningKey string = "default_signing_key"
)

func ParseFlags() Config {
	var appAddr string
	var dbConnectionString string
	var signingKey = defaultSigningKey
	var grpcAddr string
	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&grpcAddr, "g", ":8081", "port to run grpc server")
	flag.StringVar(&dbConnectionString, "d", "postgres://admin:1234@localhost:5432/gophkeeper?sslmode=disable", "database DSN")
	flag.Parse()

	if envAppAddr, exist := os.LookupEnv("RUN_ADDRESS"); exist {
		appAddr = envAppAddr
	}
	if envGRPCAddr, exist := os.LookupEnv("GRPC_ADDRESS"); exist {
		grpcAddr = envGRPCAddr
	}
	if envDBConnectionStr, exist := os.LookupEnv("DATABASE_URI"); exist {
		dbConnectionString = envDBConnectionStr
	}
	if key, exist := os.LookupEnv("SIGNING_KEY"); exist {
		signingKey = key
	}

	return Config{
		AppAddr:    appAddr,
		DBConnStr:  dbConnectionString,
		SigningKey: signingKey,
		GRPCAddr:   grpcAddr,
	}
}
