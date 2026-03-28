package command

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

type UserAuthenticator interface {
	Login(ctx context.Context, login, password string, masterPassword []byte) (string, error)
}

func Login(
	ctx context.Context,
	client UserAuthenticator,
	masterPassword []byte,
	args []string,
) tea.Cmd {
	return func() tea.Msg {
		if len(args) < 2 {
			return ErrorMsg{Err: errors.New("args: <login> <password>")}
		}

		_, err := client.Login(ctx, args[0], args[1], masterPassword)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return LoginSuccessMsg{}
	}
}
