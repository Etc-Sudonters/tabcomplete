package main

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/etc-sudonters/tabcomplete"
)

type testcompleter struct{}

func (ttc testcompleter) Complete(string) ([]string, error) {
	return []string{
		"The Million Violation",
		"Kenosis",
		"Lithopaedic",
		"Iosis",
		"Decollation",
		"Death Complex",
		"Casting Of The Self",
		"All That Was Promised",
		"Name Them Yet Build No Monument",
	}, nil
}

func (ttc testcompleter) Join(cur, selected string) string {
	return cur + " " + selected
}

type model struct {
	tc    tabcomplete.Model
	input textinput.Model
}

func (m model) Init() tea.Cmd {
	return m.input.Focus()
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
	viewBuilder := strings.Builder{}
	viewBuilder.WriteString(m.input.View())
	if m.tc.HasCandidates() {
		viewBuilder.WriteString("\n")
		viewBuilder.WriteString(m.tc.View())
	}

	return viewBuilder.String()
}

func main() {
	tc, _ := tabcomplete.NewTabCompleter(
		tabcomplete.UseCompleter(&testcompleter{}),
		tabcomplete.MaxCandidatesToDisplay(3),
		tabcomplete.BlurredStyle(lipgloss.NewStyle()),
		tabcomplete.FocusedStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#B95FF4")),
		),
	)
	input := textinput.New()
	input.Prompt = "Enter some text "
	input.Focus()
	p := tea.NewProgram(model{tc, input})
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
