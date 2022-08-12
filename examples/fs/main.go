package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/etc-sudonters/tabcomplete"
)

var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

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
	view := strings.Builder{}
	view.WriteString("Enter a path and hit tab\n")
	view.WriteString(m.tc.View())
	view.WriteString("\n")
	if m.tc.Error != nil {
		view.WriteString(errStyle.Render(fmt.Sprintf("%s: %s", m.tc.Error.Input, m.tc.Error.Err)))
	}
	view.WriteString("\n")
	view.WriteString("Hint: You can use ~ for your home directory")
	return view.String()
}

func main() {
	tc, err := tabcomplete.NewTabCompleter(
		tabcomplete.AlwaysRenderTabLine(),
		tabcomplete.UseFileSystemCompleter(),
		tabcomplete.MaxCandidatesToDisplay(10),
		tabcomplete.WithSeparator(" | "),
		tabcomplete.BlurredStyle(lipgloss.NewStyle()),
		tabcomplete.FocusedStyle(lipgloss.NewStyle().Background(lipgloss.Color("#8250df"))),
		tabcomplete.ConfigureTextInput(func(m *textinput.Model) {
			m.Prompt = "Start Typing: "
		}),
	)
	tc.Focus()
	if err != nil {
		log.Fatal(err)
	}
	p := tea.NewProgram(model{tc})
	if err = p.Start(); err != nil {
		log.Fatal(err)
	}
}
