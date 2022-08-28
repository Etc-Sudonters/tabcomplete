package tabcomplete

import (
	"github.com/charmbracelet/lipgloss"
)

type ConfigureModel func(m *Model)

func UseCompleter(tc TabCompletion) ConfigureModel {
	return func(m *Model) {
		m.tabCompletion = tc
	}
}

func FocusedStyle(s lipgloss.Style) ConfigureModel {
	return func(m *Model) {
		m.display.tabFocusStyle = s
	}
}

func BlurredStyle(s lipgloss.Style) ConfigureModel {
	return func(m *Model) {
		m.display.tabBlurStyle = s
	}
}

func MaxCandidatesToDisplay(max int) ConfigureModel {
	return func(m *Model) {
		m.maxCandidatesToDisplay = max
	}
}

func WithSeparator(sep string, sepStyle lipgloss.Style) ConfigureModel {
	return func(m *Model) {
		m.display.separator = sep
		m.display.separatorStyle = sepStyle
	}
}

func withMinCandidates(m *Model) {
	if m.maxCandidatesToDisplay < 1 {
		m.maxCandidatesToDisplay = 5
	}
}
