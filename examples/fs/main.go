package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/etc-sudonters/tabcomplete"
)

type model struct {
	tc tabcomplete.Model
}

func (m model) Init() tea.Cmd {
	return m.tc.Focus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String():
			return m, tea.Quit
		}
	}

	model, cmd := m.tc.Update(msg)
	m.tc = model.(tabcomplete.Model)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf("You can use ~ for your homedir:\nEnter a path and hit tab\n%s", m.tc.View())
}

func main() {
	tc, err := tabcomplete.NewTabCompleter(tabcomplete.TabCompleterOptions{
		TabCompletion:          tabcomplete.NewFileSystemTabCompletion(),
		MaxCandidatesToDisplay: 3,
		Separator:              " ",
		TabFocusStyle:          lipgloss.NewStyle().Background(lipgloss.Color("#8250df")),
		TabBlurStyle:           lipgloss.NewStyle(),
		InputFocusStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("#8250df")),
		InputBlurStyle:         lipgloss.NewStyle(),
		ConfigureTextInput: func(m *textinput.Model) {
			m.Prompt = "Start typing: "
		},
	})
	tc.Focus()
	if err != nil {
		log.Fatal(err)
	}
	p := tea.NewProgram(model{tc})
	if err = p.Start(); err != nil {
		log.Fatal(err)
	}
}
