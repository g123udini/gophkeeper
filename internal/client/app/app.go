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
	"github.com/g123udini/gophkeeper/internal/client/cli"
	"github.com/g123udini/gophkeeper/internal/client/command"
	"github.com/g123udini/gophkeeper/internal/client/config"
	"github.com/g123udini/gophkeeper/internal/client/grpc"
	"github.com/g123udini/gophkeeper/internal/client/manager"
	"github.com/g123udini/gophkeeper/internal/client/repository"
	"github.com/g123udini/gophkeeper/internal/client/synchronizer"
	"github.com/g123udini/gophkeeper/internal/common/logger"
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

	if err := touchFilepath(conf.DBPath); err != nil {
		return nil, fmt.Errorf("can`t touch filepath: %w", err)
	}

	db, err := sql.Open("sqlite3", conf.DBPath)
	if err != nil {
		return nil, fmt.Errorf("can`t open db: %w", err)
	}

	userDataRepo, err := repository.NewUserDataRepository(db)
	if err != nil {
		return nil, fmt.Errorf("can`t create data repo: %w", err)
	}

	metaRepo, err := repository.NewMetaRepository(db)
	if err != nil {
		return nil, fmt.Errorf("can`t create meta repo: %w", err)
	}

	userDataManager := manager.NewUserDataManager(userDataRepo)
	metaManager := manager.NewMetaManager(metaRepo)
	masterPassword := []byte(conf.MasterPassword)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ok, err := metaManager.MasterPasswordHashDefined(ctx)
	if err != nil {
		return nil, err
	}

	if !ok {
		if err := metaManager.SetMasterPassword(ctx, conf.MasterPassword); err != nil {
			return nil, err
		}
	} else {
		if err := metaManager.ValidateMasterPassword(ctx, conf.MasterPassword); err != nil {
			return nil, fmt.Errorf("invalid master password: %w", err)
		}
	}

	client, err := grpc.NewClient(conf.ServerAddr)
	if err != nil {
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
			return command.Get(ctx, userDataManager, masterPassword, args)
		},
		"set": func(ctx context.Context, args []string) tea.Cmd {
			return command.Set(ctx, userDataManager, masterPassword, args)
		},
		"login": func(ctx context.Context, args []string) tea.Cmd {
			return command.Login(ctx, client, masterPassword, args)
		},
		"register": func(ctx context.Context, args []string) tea.Cmd {
			return command.Register(ctx, client, masterPassword, args)
		},
	}

	return &App{
		syncer:   syncer,
		registry: registry,
		db:       db,
	}, nil
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

func (a *App) Run() {
	defer a.db.Close()
	defer a.syncer.Stop()

	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	wg.Add(2)

	go func() {
		defer wg.Done()
		a.syncer.Start(ctx)
		stop()
	}()

	go func() {
		defer wg.Done()
		cli.Run(ctx, a.registry)
		stop()
	}()

	wg.Wait()
}
