package command

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/g123udini/gophkeeper/internal/client/aes"
	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/client/value"
)

type UserDataGetter interface {
	Get(ctx context.Context, key string) (*model.UserData, error)
}

func Get(
	ctx context.Context,
	dataManager UserDataGetter,
	masterPassword []byte,
	args []string,
) tea.Cmd {
	return func() tea.Msg {
		if len(args) < 1 {
			return ErrorMsg{Err: errors.New("args: <key>")}
		}

		data, err := dataManager.Get(ctx, args[0])
		if err != nil {
			return ErrorMsg{Err: err}
		}

		raw, err := aes.Decrypt(masterPassword, data.DataValue)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		val, err := value.FromBytes(raw)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return GetSuccessMsg{Value: val.String()}
	}
}
