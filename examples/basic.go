package examples

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/etc-sudonters/tabcomplete"
)

type testcompleter struct{}

func (ttc testcompleter) Complete(string) []string {
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
	}
}

func (ttc testcompleter) Rank(_ string, candidates []string) []string {
	return candidates
}

func (ttc testcompleter) Join(cur, selected string) string {
	return cur + " " + selected
}

type model struct {
	tc tabcomplete.Model
}

func (m model) Init() tea.Cmd {
	return m.tc.Focus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.tc.Update(msg)
	m.tc = model.(tabcomplete.Model)
	return m, cmd
}

func (m model) View() string {
	return m.tc.View()
}

func main() {
	tc, err := tabcomplete.NewTabCompleter(tabcomplete.TabCompleterOptions{
		TabCompletion:          testcompleter{},
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
	if err != nil {
		log.Fatal(err)
	}
	p := tea.NewProgram(model{tc})
	if err = p.Start(); err != nil {
		log.Fatal(err)
	}
}
