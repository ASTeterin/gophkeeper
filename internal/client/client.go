package client

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"

	"github.com/ASTeterin/gophkeeper/internal/config"
	"github.com/ASTeterin/gophkeeper/internal/grpc"
)

var (
	version   = "1.0.0 "
	buildDate = "unknown"
)

const (
	keySize    = 32 // AES-256
	iterations = 100000
)

type App struct {
	config    *config.Config
	client    *grpc.Client
	store     *LocalStore
	isOnline  bool
	masterKey []byte
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
		store:  NewLocalStore("./local_store.json"),
	}
}

func deriveKey(password string) []byte {
	salt := sha256.Sum256([]byte(password))
	return pbkdf2.Key([]byte(password), salt[:], iterations, keySize, sha256.New)
}

func (a *App) promptMasterPassword() error {
	fmt.Print("Enter master password for encryption: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read password")
	}
	password := scanner.Text()

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	a.masterKey = deriveKey(password)
	return nil
}

func (a *App) encryptData(plaintext []byte) (string, error) {
	if a.masterKey == nil {
		return "", fmt.Errorf("master key not initialized")
	}

	block, err := aes.NewCipher(a.masterKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a *App) decryptData(ciphertext string) ([]byte, error) {
	if a.masterKey == nil {
		return nil, fmt.Errorf("master key not initialized")
	}

	encoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	cipher, err := chacha20poly1305.New(a.masterKey)
	if err != nil {
		return nil, err
	}

	nonceSize := chacha20poly1305.NonceSize
	if len(encoded) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedBytes := encoded[:nonceSize], encoded[nonceSize:]

	plaintext, err := cipher.Open(nil, nonce, encryptedBytes, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (a *App) Run() error {
	if err := a.promptMasterPassword(); err != nil {
		return fmt.Errorf("password error: %v", err)
	}

	host, _, err := net.SplitHostPort(a.config.AppAddr)
	if err != nil {
		host = ""
	}
	addr := net.JoinHostPort(host, a.config.GRPCAddr)

	cl, err := grpc.NewClient(addr)
	if err != nil {
		fmt.Println("Warning: Server unavailable. Working in offline mode.")
		a.isOnline = false
	} else {
		a.client = cl
		defer cl.Close()
		a.isOnline = true
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Gophkeeper Client. Type 'help' for commands.")

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		args := strings.Fields(scanner.Text())
		if len(args) == 0 {
			continue
		}

		cmd := args[0]
		switch cmd {
		case "help":
			a.printHelp()
		case "version":
			fmt.Printf("Version: %s\nBuild Date: %s\n", version, buildDate)
		case "register", "login":
			if !a.isOnline {
				fmt.Println("Command requires server connection.")
				continue
			}
			if len(args) < 3 {
				fmt.Println("Usage: <command> <login> <password>")
				continue
			}
			if cmd == "register" {
				a.handleRegister(args[1], args[2])
			} else {
				a.handleLogin(args[1], args[2])
			}
		case "add":
			if len(args) < 3 {
				fmt.Println("Usage: add <key> <data>")
				continue
			}
			a.handleAdd(args[1], args[2])
		case "get":
			if len(args) < 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			a.handleGet(args[1])
		case "list":
			a.handleList()
		case "sync":
			a.handleSync()
		case "exit", "quit":
			return nil
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}
	return nil
}

func (a *App) printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  register <login> <pass>  Register a new user (online)")
	fmt.Println("  login <login> <pass>     Login existing user (online)")
	fmt.Println("  add <key> <data>         Add data (offline/online)")
	fmt.Println("  get <key>                Get data (offline/online)")
	fmt.Println("  list                     List all data (offline/online)")
	fmt.Println("  sync                     Sync with server (online)")
	fmt.Println("  version                  Show version")
	fmt.Println("  exit                     Exit")
}

func (a *App) handleRegister(login, pass string) {
	_, err := a.client.Register(context.Background(), login, pass)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("User registered successfully.")
	}
}

func (a *App) handleLogin(login, pass string) {
	_, err := a.client.Authenticate(context.Background(), login, pass)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Logged in successfully.")
	}
}

func (a *App) handleAdd(key, data string) {
	encryptedData, err := a.encryptData([]byte(data))
	if err != nil {
		fmt.Printf("Encryption error: %v\n", err)
		return
	}

	a.store.Add(key, "", []byte(encryptedData))
	fmt.Printf("Data added for key: %s\n", key)

	if a.isOnline {
		err := a.client.AddData(context.Background(), key, "", []byte(encryptedData))
		if err != nil {
			fmt.Printf("Sync error: %v\n", err)
		}
	}
}

func (a *App) handleGet(key string) {
	item, ok := a.store.Get(key)
	if !ok {
		fmt.Println("No data found.")
		return
	}

	decryptedData, err := a.decryptData(string(item.Data))
	if err != nil {
		fmt.Printf("Decryption error: %v\n", err)
		return
	}

	fmt.Printf("Data for key %s: %s\n", key, string(decryptedData))
}

func (a *App) handleList() {
	items := a.store.List()
	if len(items) == 0 {
		fmt.Println("No data found.")
		return
	}
	for _, item := range items {
		fmt.Printf("Key: %s | Desc: %s\n", item.Key, item.Description)
	}
}

func (a *App) handleSync() {
	if !a.isOnline {
		fmt.Println("Sync requires server connection.")
		return
	}

	items, err := a.client.GetAllData(context.Background())
	if err != nil {
		fmt.Printf("Sync error: %v\n", err)
		return
	}

	localItems := make([]*Item, len(items))
	for i, item := range items {
		localItems[i] = &Item{
			Key:         item.DataKey,
			Description: item.Description,
			Data:        item.Data,
			Dirty:       false,
		}
	}
	a.store.Sync(localItems)
	fmt.Println("Sync completed successfully.")
}
