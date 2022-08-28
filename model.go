package tabcomplete

import (
	"errors"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	lastID                     int
	idMtx                      sync.Mutex
	ErrNoTabCompletionProvided = errors.New("tab completion must be provided")
	ErrNoCandidates            = errors.New("no candidates to select")
)

func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

type tabDisplay struct {
	separator      string
	tabFocusStyle  lipgloss.Style
	tabBlurStyle   lipgloss.Style
	separatorStyle lipgloss.Style
}

type Model struct {
	tabCompletion          TabCompletion
	maxCandidatesToDisplay int
	display                *tabDisplay
	state                  *tabCompleteState
	id                     int
	Error                  *TabError
}

func NewTabCompleter(opts ...ConfigureModel) (Model, error) {
	m := &Model{
		display: &tabDisplay{
			separator: " ",
		},
		id: nextID(),
	}

	for i := range opts {
		opts[i](m)
	}

	withMinCandidates(m)

	if m.tabCompletion == nil {
		return *m, ErrNoTabCompletionProvided
	}

	return *m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd = nil
	if msg, ok := msg.(Message); ok {

		if msg.id != m.id {
			return m, nil
		}

		switch msg := msg.kind.(type) {
		case clear:
			m.state = nil
			m.Error = nil
		case completed:
			m.state = newTabState(m.maxCandidatesToDisplay, msg.input, msg.candidates)
			m.Error = nil
		case tabErr:
			m.state = nil
			m.Error = &TabError{
				Input: msg.input,
				Err:   msg.err,
			}
		case moveNext:
			if m.state != nil {
				m.state.moveNext()
			}
		case movePrev:
			if m.state != nil {
				m.state.movePrev()
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	view := ""

	if m.state != nil && len(m.state.candidates) > 0 {
		displayedCandidates := make([]string, len(m.state.displayView))
		for i, candidate := range m.state.displayView {
			if i == m.state.displayCursor {
				displayedCandidates[i] = m.display.tabFocusStyle.Render(candidate)
			} else {
				displayedCandidates[i] = m.display.tabBlurStyle.Render(candidate)
			}
		}

		view = strings.Join(displayedCandidates, m.display.separatorStyle.Render(m.display.separator))
	}

	return view
}

func (m Model) Complete(input string) tea.Cmd {
	return func() tea.Msg {
		candidates, err := m.tabCompletion.Complete(input)
		if err != nil {
			return Message{
				kind: tabErr{
					input: input,
					err:   err,
				},
				id: m.id,
			}
		}

		return Message{
			kind: completed{
				input:      input,
				candidates: candidates,
			},
			id: m.id,
		}
	}
}

func (m Model) JoinCandidate(input, candidate string) string {
	return m.tabCompletion.Join(input, candidate)
}

func (m Model) Clear() tea.Cmd {
	return func() tea.Msg {
		return Message{
			kind: clear{},
			id:   m.id,
		}
	}
}

func (m Model) MoveNext() tea.Cmd {
	return func() tea.Msg {
		return Message{
			kind: moveNext{},
			id:   m.id,
		}
	}
}

func (m Model) MovePrev() tea.Cmd {
	return func() tea.Msg {
		return Message{
			kind: movePrev{},
			id:   m.id,
		}
	}
}

func (m Model) SelectCurrent() (string, tea.Cmd, error) {
	if m.state == nil || len(m.state.candidates) == 0 {
		return "", nil, ErrNoCandidates
	}

	cmd := func() tea.Msg {
		return Message{
			kind: clear{},
			id:   m.id,
		}
	}

	return m.state.selectCurrent(), cmd, nil
}

func (m Model) HasCandidates() bool {
	return m.state != nil && len(m.state.candidates) > 0
}
