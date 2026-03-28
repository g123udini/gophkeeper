package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/g123udini/gophkeeper/internal/client/cli"
	"github.com/g123udini/gophkeeper/internal/client/command"
	"github.com/g123udini/gophkeeper/internal/client/config"
	"github.com/g123udini/gophkeeper/internal/client/grpc"
	"github.com/g123udini/gophkeeper/internal/client/repository"
	"github.com/g123udini/gophkeeper/internal/client/service"
	"github.com/g123udini/gophkeeper/internal/client/synchronizer"
	"github.com/g123udini/gophkeeper/internal/common/logger"
	"github.com/golang-migrate/migrate/v4"
	sqlitemigrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type App struct {
	syncer   *synchronizer.Synchronizer
	registry cli.CommandRegistry
	db       *sql.DB
}

func New() (*App, error) {
	conf, err := config.ParseArgs()
	if err != nil {
		return nil, fmt.Errorf("can`t parse arguments: %w", err)
	}

	logger.Init("client", zap.InfoLevel.String())

	db, err := initDB(conf.DBPath)
	if err != nil {
		return nil, err
	}

	userDataRepo, err := repository.NewUserDataRepository(db)
	if err != nil {
		return nil, err
	}
	metaRepo, err := repository.NewMetaRepository(db)
	if err != nil {
		return nil, err
	}

	userDataManager := service.NewUserDataManager(userDataRepo)
	metaManager := service.NewMetaManager(metaRepo)

	masterPassword, err := readPassword("Enter master password: ")
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("can`t read master password: %w", err)
	}
	masterPasswordBytes := []byte(masterPassword)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ok, err := metaManager.MasterPasswordHashDefined(ctx)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("can`t check master password state: %w", err)
	}

	if !ok {
		if err := metaManager.SetMasterPassword(ctx, masterPassword); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("can`t set master password: %w", err)
		}
	} else {
		if err := metaManager.ValidateMasterPassword(ctx, masterPassword); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("invalid master password: %w", err)
		}
	}

	client, err := grpc.NewClient(conf.ServerAddr)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("can`t create client: %w", err)
	}

	syncer := synchronizer.New(
		client,
		userDataManager,
		metaManager,
		time.Duration(conf.SyncIntervalSec)*time.Second,
	)

	registry := cli.CommandRegistry{
		"get": func(ctx context.Context, args []string) tea.Cmd {
			return command.Get(ctx, userDataManager, masterPasswordBytes, args)
		},
		"set": func(ctx context.Context, args []string) tea.Cmd {
			return command.Set(ctx, userDataManager, masterPasswordBytes, args)
		},
		"login": func(ctx context.Context, args []string) tea.Cmd {
			return command.Login(ctx, client, masterPasswordBytes, args)
		},
		"register": func(ctx context.Context, args []string) tea.Cmd {
			return command.Register(ctx, client, masterPasswordBytes, args)
		},
	}

	return &App{
		syncer:   syncer,
		registry: registry,
		db:       db,
	}, nil
}

func initDB(path string) (*sql.DB, error) {
	if err := touchFilepath(path); err != nil {
		return nil, fmt.Errorf("can`t touch filepath: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can`t open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("can`t ping db: %w", err)
	}

	if err := runMigrations(db, "file://internal/client/migrations"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("can`t run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := sqlitemigrate.WithInstance(db, &sqlitemigrate.Config{})
	if err != nil {
		return fmt.Errorf("create sqlite migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func touchFilepath(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}

		return file.Close()
	}

	return nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	password, err := term.ReadPassword(uintptr(int(os.Stdin.Fd())))
	fmt.Println()
	if err != nil {
		return "", err
	}

	return string(password), nil
}

func (a *App) Run() {
	defer a.db.Close()
	defer a.syncer.Stop()

	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := a.syncer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Logger.Error("synchronizer stopped with error", zap.Error(err))
		}

		stop()
	}()

	go func() {
		defer wg.Done()
		cli.Run(ctx, a.registry)
		stop()
	}()

	wg.Wait()
}
