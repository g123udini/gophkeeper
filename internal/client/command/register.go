package command

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

type UserRegistrar interface {
	Register(ctx context.Context, login, password string, masterPassword []byte) (string, error)
}

func Register(
	ctx context.Context,
	client UserRegistrar,
	masterPassword []byte,
	args []string,
) tea.Cmd {
	return func() tea.Msg {
		if len(args) < 2 {
			return ErrorMsg{Err: errors.New("args: <login> <password>")}
		}

		_, err := client.Register(ctx, args[0], args[1], masterPassword)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return RegisterSuccessMsg{}
	}
}
