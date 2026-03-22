package command

import tea "github.com/charmbracelet/bubbletea"

type ErrorMsg struct {
	Err error
}

func (m ErrorMsg) Error() string {
	if m.Err == nil {
		return ""
	}

	return m.Err.Error()
}

func (m ErrorMsg) Unwrap() error {
	return m.Err
}

type LoginSuccessMsg struct{}
type RegisterSuccessMsg struct{}

type GetSuccessMsg struct {
	Value string
}

type SetSuccessMsg struct{}

func Run(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}

	return cmd()
}
