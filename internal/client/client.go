package client

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ASTeterin/gophkeeper/internal/config"
	"github.com/ASTeterin/gophkeeper/internal/grpc"
)

type App struct {
	config *config.Config
	client *grpc.Client
}

func New(cfg *config.Config) *App {
	return &App{config: cfg}
}

func (a *App) Run() error {
	host, _, err := net.SplitHostPort(a.config.AppAddr)
	if err != nil {
		host = ""
	}
	addr := net.JoinHostPort(host, a.config.GRPCAddr)

	cl, err := grpc.NewClient(addr)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer cl.Close()
	a.client = cl

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
		case "register":
			if len(args) < 3 {
				fmt.Println("Usage: register <login> <password>")
				continue
			}
			a.handleRegister(args[1], args[2])
		case "login":
			if len(args) < 3 {
				fmt.Println("Usage: login <login> <password>")
				continue
			}
			a.handleLogin(args[1], args[2])
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
	fmt.Println("  register <login> <pass>  Register a new user")
	fmt.Println("  login <login> <pass>     Login existing user")
	fmt.Println("  add <key> <data>         Add private data")
	fmt.Println("  get <key> 		        Get private data by key")
	fmt.Println("  list                     List all data keys")
	fmt.Println("  sync                     Sync data (demo: clears and adds sample)")
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
	err := a.client.AddData(context.Background(), key, "", []byte(data))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Data added for key: %s\n", key)
	}
}

func (a *App) handleGet(key string) {
	data, err := a.client.GetByKey(context.Background(), key)
	fmt.Println("!!!!!!!!", data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if data == nil {
		fmt.Println("No data found.")
		return
	}
	fmt.Printf("Data for key %s: %s\n", key, string(data.Data))
}

func (a *App) handleList() {
	items, err := a.client.GetAllData(context.Background())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(items) == 0 {
		fmt.Println("No data found.")
		return
	}
	for _, item := range items {
		fmt.Printf("Key: %s | Desc: %s\n", item.DataKey, item.Description)
	}
}

func (a *App) handleSync() {
	// Demo sync: replaces all data with a single item
	items := []grpc.SyncItem{
		{Key: "synced_key", Description: "Synced via CLI", Data: []byte("synced_value")},
	}
	err := a.client.SyncData(context.Background(), items)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Data synced successfully.")
	}
}
