package tabcomplete

import (
	"errors"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	lastID int
	idMtx  sync.Mutex
)

func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

type TabCompletion interface {
	Complete(string) ([]string, error)
	Rank(string, []string) []string
	Join(string, string) string
}

type TabCompleterOptions struct {
	TabCompletion          TabCompletion
	MaxCandidatesToDisplay int
	Separator              string
	Width, Height          int
	TabFocusStyle          lipgloss.Style
	TabBlurStyle           lipgloss.Style
	InputFocusStyle        lipgloss.Style
	InputBlurStyle         lipgloss.Style
	ConfigureTextInput     func(*textinput.Model)
}

type TabError struct {
	Input string
	Err   error
}

type Model struct {
	tabCompletion          TabCompletion
	maxCandidatesToDisplay int
	display                *tabDisplay
	state                  *tabCompleteState
	input                  textinput.Model
	id                     int
	focus                  bool
	Error                  *TabError
}

type tabDisplay struct {
	separator       string
	width, height   int
	tabFocusStyle   lipgloss.Style
	tabBlurStyle    lipgloss.Style
	inputFocusStyle lipgloss.Style
	inputBlurStyle  lipgloss.Style
}

// TODO(ANR): Do we need to also tag the state w/ an Id?
type tabCompleteState struct {
	initial         string
	candidates      []string
	displayView     []string
	candidateCursor int
	displayCursor   int
}

func (s *tabCompleteState) createDisplayList(maxCandidatesToDisplay int) {
	if maxCandidatesToDisplay > len(s.candidates) {
		s.displayView = s.candidates
		return
	}

	s.displayView = s.candidates[0:maxCandidatesToDisplay]
}

func (s *tabCompleteState) moveNext() {
	if s.candidateCursor == len(s.candidates)-1 {
		return
	}

	s.candidateCursor++
	s.displayCursor++

	// rolled off edge, but we know there's at least one more element after
	if s.displayCursor >= len(s.displayView) {
		s.displayCursor = len(s.displayView) - 1
		s.displayView = append(
			s.displayView[1:],
			s.candidates[s.candidateCursor],
		)
	}
}
func (s *tabCompleteState) movePrev() {
	if s.candidateCursor == 0 {
		return
	}

	s.candidateCursor--
	s.displayCursor--

	// rolled off edge of display
	if s.displayCursor < 0 {
		s.displayCursor = 0
		s.displayView = append(
			[]string{s.candidates[s.candidateCursor]},
			s.displayView[:len(s.displayView)-1]...,
		)
	}
}
func (s *tabCompleteState) selectCurrent() string {
	return s.candidates[s.candidateCursor]
}

var ErrNoTabCompletionProvided = errors.New("tab completion must be provided")

func NewTabCompleter(opts TabCompleterOptions) (m Model, err error) {
	if opts.TabCompletion == nil {
		err = ErrNoTabCompletionProvided
		return
	}

	input := textinput.NewModel()

	if opts.ConfigureTextInput != nil {
		opts.ConfigureTextInput(&input)
	}

	display := &tabDisplay{
		separator:       opts.Separator,
		width:           opts.Width,
		height:          opts.Height,
		tabFocusStyle:   opts.TabFocusStyle,
		tabBlurStyle:    opts.TabBlurStyle,
		inputFocusStyle: opts.InputFocusStyle,
		inputBlurStyle:  opts.TabBlurStyle,
	}

	if display.separator == "" {
		display.separator = " "
	}

	if opts.MaxCandidatesToDisplay < 1 {
		opts.MaxCandidatesToDisplay = 5
	}

	m = Model{
		tabCompletion:          opts.TabCompletion,
		maxCandidatesToDisplay: opts.MaxCandidatesToDisplay,
		state:                  nil,
		display:                display,
		input:                  input,
		id:                     nextID(),
	}
	return
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Focus() tea.Cmd {
	m.input.PromptStyle = m.display.inputFocusStyle
	m.input.TextStyle = m.display.inputFocusStyle
	m.focus = true
	return m.input.Focus()
}

func (m *Model) Blur() {
	m.focus = false
	m.state = nil
	m.input.Blur()
	m.input.PromptStyle = m.display.inputBlurStyle
	m.input.TextStyle = m.display.inputBlurStyle
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// we're not active
	if !m.focus {
		return m, nil
	}
	switch msg := msg.(type) {
	case Message:
		// not us
		if msg.id != m.id {
			return m, nil
		}

		cmd := m.handleTabMessage(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String():
			return m, tea.Quit
		case tea.KeyEsc.String():
			m.state = nil
			return m, nil
		case tea.KeyTab.String():
			curVal := m.input.Value()
			return m, func() tea.Msg {
				return Message{
					kind: started{curVal},
					id:   m.id,
				}
			}
		case tea.KeyShiftTab.String():
			if m.state != nil {
				m.state.movePrev()
			}
			return m, nil
		case tea.KeyLeft.String():
			if m.state != nil {
				m.state.movePrev()
				return m, nil
			}
		case tea.KeyRight.String():
			if m.state != nil {
				m.state.moveNext()
				return m, nil
			}
		case tea.KeyEnter.String():
			if m.state != nil {
				selectedVal := m.state.selectCurrent()
				curVal := m.input.Value()
				return m, func() tea.Msg {
					return Message{
						id: m.id,
						kind: selected{
							current:  curVal,
							selected: selectedVal,
						},
					}
				}
			}
		}
	}

	curVal := m.input.Value()
	newInput, cmd := m.input.Update(msg)
	if newInput.Value() != curVal {
		cmd = tea.Batch(cmd, func() tea.Msg {
			return Message{
				id:   m.id,
				kind: canceled{},
			}
		})
	}
	m.input = newInput

	return m, cmd
}

func (m *Model) handleTabMessage(msg Message) tea.Cmd {
	switch msg := msg.kind.(type) {
	case tabErr:
		m.state = nil
		m.Error = &TabError{
			Input: msg.input,
			Err:   msg.err,
		}
		return nil
	case started:
		if m.state != nil {
			m.state.moveNext()
			return nil
		}
		m.Error = nil
		return func() tea.Msg {
			candidates, err := m.tabCompletion.Complete(msg.input)

			if err != nil {
				return Message{
					id: m.id,
					kind: tabErr{
						input: msg.input,
						err:   err,
					},
				}
			}

			return Message{
				id: m.id,
				kind: completed{
					input:      msg.input,
					candidates: candidates,
				},
			}
		}

	case completed:
		if len(msg.candidates) == 0 {
			return nil
		}

		ranked := m.tabCompletion.Rank(msg.input, msg.candidates)

		m.state = &tabCompleteState{
			initial:         msg.input,
			candidates:      ranked,
			candidateCursor: 0,
			displayCursor:   0,
		}
		m.state.createDisplayList(m.maxCandidatesToDisplay)
		return nil

	case selected:
		// probably canceled
		if m.state == nil {
			return nil
		}
		joined := m.tabCompletion.Join(msg.current, msg.selected)
		m.input.SetValue(joined)
		m.input.SetCursor(len(m.input.Value()))
		m.state = nil
		return nil
	case canceled:
		m.state = nil
		m.Error = nil
		return nil
	}

	return nil
}

func (m Model) View() string {
	view := strings.Builder{}
	view.WriteString(m.input.View() + "\n")

	if m.state != nil {
		for i, candidate := range m.state.displayView {
			if i == m.state.displayCursor {
				view.WriteString(m.display.tabFocusStyle.Render(candidate))
			} else {
				view.WriteString(m.display.tabBlurStyle.Render(candidate))
			}

			view.WriteString(m.display.separator)
		}
	}

	return view.String()
}

type Message struct {
	kind interface{}
	id   int
}

type completed struct {
	input      string
	candidates []string
}

type started struct {
	input string
}

type selected struct {
	selected string
	current  string
}

type canceled struct{}

type tabErr struct {
	input string
	err   error
}
