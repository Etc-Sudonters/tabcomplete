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
	tc    tabcomplete.Model
	input textinput.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String():
			return m, tea.Quit
		case tea.KeyTab.String():
			if m.tc.HasCandidates() {
				return m, m.tc.MoveNext()
			}
			return m, m.tc.Complete(m.input.Value())
		case tea.KeyShiftTab.String():
			if m.tc.HasCandidates() {
				return m, m.tc.MovePrev()
			}
		case tea.KeyEnter.String():
			if m.tc.HasCandidates() {
				candidate, cmd, err := m.tc.SelectCurrent()
				if err != nil {
					return m, nil
				}

				m.input.SetValue(m.tc.JoinCandidate(m.input.Value(), candidate))
				return m, cmd
			}
		}
	case tabcomplete.Message:
		model, cmd := m.tc.Update(msg)
		m.tc = model
		return m, cmd
	}

	model, cmd := m.input.Update(msg)
	m.input = model
	return m, cmd
}

func (m model) View() string {
	view := strings.Builder{}
	view.WriteString("Enter a path and hit tab\n")
	view.WriteString(m.input.View())
	view.WriteString("\n")
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
		tabcomplete.UseFileSystemCompleter(),
		tabcomplete.MaxCandidatesToDisplay(10),
		tabcomplete.WithSeparator(" | ", lipgloss.NewStyle()),
		tabcomplete.BlurredStyle(lipgloss.NewStyle()),
		tabcomplete.FocusedStyle(lipgloss.NewStyle().Background(lipgloss.Color("#8250df"))),
	)
	input := textinput.New()
	input.Prompt = "Path: "
	input.Focus()
	if err != nil {
		log.Fatal(err)
	}
	p := tea.NewProgram(model{tc, input})
	if err = p.Start(); err != nil {
		log.Fatal(err)
	}
}
