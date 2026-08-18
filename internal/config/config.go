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
	CertDir    string
	ServerName string
}

const (
	defaultSigningKey string = "default_signing_key"
	defaultCertDir    string = "./../../cert/"
	defaultServerName string = "localhost"
)

func ParseFlags() Config {
	var appAddr string
	var dbConnectionString string
	var signingKey = defaultSigningKey
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

	return Config{
		AppAddr:    appAddr,
		DBConnStr:  dbConnectionString,
		SigningKey: signingKey,
		GRPCAddr:   grpcAddr,
		CertDir:    certDir,
		ServerName: serverName,
	}
}
