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
	state                  CandidateNavigator
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
	if msg, ok := msg.(Message); ok {

		if msg.id != m.id {
			return m, nil
		}

		switch msg := msg.kind.(type) {
		case clear:
			m.state = nil
			m.Error = nil
		case completed:
			m.state = newPagedCandidateNavigator(m.maxCandidatesToDisplay, msg.candidates)
			m.Error = nil
		case tabErr:
			m.state = nil
			m.Error = &TabError{
				Input: msg.input,
				Err:   msg.err,
			}
		case moveNext:
			if m.state != nil {
				m.state.MoveCursorNext()
			}
		case movePrev:
			if m.state != nil {
				m.state.MoveCursorPrev()
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	view := ""

	if m.state != nil {
		currentCandidates := m.state.CurrentDisplay()
		displayedCandidates := make([]string, len(currentCandidates))
		for i, candidate := range currentCandidates {
			if candidate.Current() {
				displayedCandidates[i] = m.display.tabFocusStyle.Render(candidate.String())
			} else {
				displayedCandidates[i] = m.display.tabBlurStyle.Render(candidate.String())
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
	if m.state == nil {
		return "", nil, ErrNoCandidates
	}

	cmd := func() tea.Msg {
		return Message{
			kind: clear{},
			id:   m.id,
		}
	}

	return m.state.SelectCurrent(), cmd, nil
}

func (m Model) HasCandidates() bool {
	return m.state != nil
}
