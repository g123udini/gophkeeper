package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/g123udini/gophkeeper/internal/client/command"
	"github.com/g123udini/gophkeeper/internal/common/logger"
	"go.uber.org/zap"
)

type CommandFunc func(ctx context.Context, args []string) tea.Cmd

type CommandRegistry map[string]CommandFunc

func Run(ctx context.Context, registry CommandRegistry) {
	reader := bufio.NewScanner(os.Stdin)

	logger.Logger.Info("GophKeeper started")
	fmt.Print("> ")

	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if len(line) == 0 {
			fmt.Print("> ")
			continue
		}

		parts := strings.Fields(line)
		cmdName := parts[0]
		args := parts[1:]

		cmdFactory, ok := registry[cmdName]
		if !ok {
			logger.Logger.Error("Unknown command", zap.String("command", cmdName))
			fmt.Print("> ")
			continue
		}

		cmd := cmdFactory(ctx, args)
		msg := command.Run(cmd)

		switch m := msg.(type) {
		case nil:
		case command.ErrorMsg:
			logger.Logger.Error("Command execution error", zap.Error(m.Err))
		case command.LoginSuccessMsg:
			fmt.Println("login successful")
		case command.RegisterSuccessMsg:
			fmt.Println("register successful")
		case command.GetSuccessMsg:
			fmt.Println(m.Value)
		case command.SetSuccessMsg:
			fmt.Println("ok")
		default:
			fmt.Printf("%v\n", m)
		}

		fmt.Print("> ")
	}

	if err := reader.Err(); err != nil {
		logger.Logger.Fatal("STDIN read error", zap.Error(err))
	}
}
