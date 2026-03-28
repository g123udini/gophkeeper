package command

import (
	"context"
	"errors"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/g123udini/gophkeeper/internal/client/aes"
	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/client/value"
)

type DataUpserter interface {
	Upsert(ctx context.Context, data *model.UserData) error
}

func Set(
	ctx context.Context,
	dataManager DataUpserter,
	masterPassword []byte,
	args []string,
) tea.Cmd {
	return func() tea.Msg {
		if len(args) < 3 {
			return ErrorMsg{Err: errors.New("args: <key> <type> <args...>")}
		}

		val, err := value.FromUserInput(args[1], args[2:])
		if err != nil {
			return ErrorMsg{Err: err}
		}

		raw, err := val.ToBytes()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		encRaw, err := aes.Encrypt(masterPassword, raw)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		err = dataManager.Upsert(ctx, &model.UserData{
			DataKey:   args[0],
			DataValue: encRaw,
			UpdatedAt: time.Now(),
			DeletedAt: time.Unix(0, 0),
		})
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return SetSuccessMsg{}
	}
}
