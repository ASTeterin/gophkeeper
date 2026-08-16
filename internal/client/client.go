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

var (
	version   = "dev"
	buildDate = "unknown"
)

type App struct {
	config   *config.Config
	client   *grpc.Client
	store    *LocalStore
	isOnline bool
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
		store:  NewLocalStore("./local_store.json"), // Путь к локальному файлу
	}
}

func (a *App) Run() error {
	// Пытаемся подключиться к серверу
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
	a.store.Add(key, "", []byte(data))
	fmt.Printf("Data added for key: %s\n", key)

	if a.isOnline {
		err := a.client.AddData(context.Background(), key, "", []byte(data))
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
	fmt.Printf("Data for key %s: %s\n", key, string(item.Data))
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
