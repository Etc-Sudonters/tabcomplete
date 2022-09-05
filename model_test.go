package tabcomplete

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func TestModel_RejectsMessageWithDifferentId(t *testing.T) {
	messageKinds := []interface{}{
		clear{},
		moveNext{},
		movePrev{},
		tabErr{},
		completed{},
	}

	for _, k := range messageKinds {
		kind := k

		t.Run(fmt.Sprintf("%T", kind), func(t *testing.T) {

			model, err := NewTabCompleter(UseCompleter(&TestTabCompleter{}))
			require.Nil(t, err)

			_, cmd := model.Update(Message{
				kind: kind,
				id:   model.id + 1,
			})

			require.Nil(t, cmd)
		})
	}
}

func TestModel_UpdatesViewWhenTabResultsArrive(t *testing.T) {
	completer := &TestTabCompleter{
		Entries: []string{"shattered faith", "kenosis", "verminous barrier", "irene"},
	}
	sep := " "
	model, err := NewTabCompleter(UseCompleter(completer), WithSeparator(sep, lipgloss.NewStyle()))

	require.Nil(t, err)

	require.Equal(t, EMPTY_DISPLAY, model.View())

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	expectedDisplay := DisplayViewHelper{
		Expected:  completer.Entries,
		Separator: sep,
	}

	require.Equal(t, expectedDisplay.String(), model.View())

}

func TestModel_ShowsNoResultsWhenClearMessageArrives(t *testing.T) {
	completer := &TestTabCompleter{
		Entries: []string{"dont", "see", "me"},
	}
	sep := " "
	model, err := NewTabCompleter(UseCompleter(completer), WithSeparator(sep, lipgloss.NewStyle()))
	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	require.Equal(t, strings.Join(completer.Entries, sep), model.View())

	model = model.setClear()
	require.Equal(t, EMPTY_DISPLAY, model.View())
}

func TestModel_Highlights_CurrentElement(t *testing.T) {
	completer := &TestTabCompleter{
		Entries: []string{"Old Paradise", "New Breakfast Habit", "Last Line Blues"},
	}
	sep := "|"

	blurredStyle := lipgloss.NewStyle().Strikethrough(true)

	model, err := NewTabCompleter(
		UseCompleter(completer),
		WithSeparator(sep, lipgloss.NewStyle()),
		BlurredStyle(blurredStyle),
	)

	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	expectedDisplay := &DisplayViewHelper{
		Expected:     completer.Entries,
		FocusedIndex: 0,
		BlurredStyle: blurredStyle,
		Separator:    sep,
	}

	require.Equal(t, expectedDisplay.String(), model.View())
}

func TestModel_MovesNextThrough_SelectionList(t *testing.T) {
	completer := &TestTabCompleter{
		Entries: []string{"Osman's Dream", "Inauspicious Prayer", "In Web"},
	}

	sep := "|"

	blurredStyle := lipgloss.NewStyle().Strikethrough(true)

	model, err := NewTabCompleter(
		UseCompleter(completer),
		WithSeparator(sep, lipgloss.NewStyle()),
		BlurredStyle(blurredStyle),
		MaxCandidatesToDisplay(3),
	)

	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	for i := range completer.Entries {
		n := i

		t.Run(completer.Entries[n], func(t *testing.T) {
			m := model.moveNextN(n)
			expectedDisplay := DisplayViewHelper{
				Expected:     completer.Entries,
				Separator:    sep,
				FocusedIndex: n,
				BlurredStyle: blurredStyle,
			}
			require.Equal(t, expectedDisplay.String(), m.View())
		})
	}

}

func TestModel_MovesPrevThrough_SelectionList(t *testing.T) {
	completer := &TestTabCompleter{
		Entries: []string{"Glass Shards", "Luminous Jar", "Warm Bed"},
	}

	sep := "|"

	blurredStyle := lipgloss.NewStyle().Strikethrough(true)

	model, err := NewTabCompleter(
		UseCompleter(completer),
		WithSeparator(sep, lipgloss.NewStyle()),
		BlurredStyle(blurredStyle),
		MaxCandidatesToDisplay(len(completer.Entries)),
	)
	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)
	model = model.moveNextN(len(completer.Entries))

	for i := range completer.Entries {
		rollback := i
		expectedIndex := len(completer.Entries) - 1 - rollback

		t.Run(completer.Entries[expectedIndex], func(t *testing.T) {
			m := model.movePrevN(rollback)
			expectedDisplay := DisplayViewHelper{
				Expected:     completer.Entries,
				Separator:    sep,
				FocusedIndex: expectedIndex,
				BlurredStyle: blurredStyle,
			}

			require.Equal(t, expectedDisplay.String(), m.View())
		})
	}
}

func TestPagesThroughCompletedCandidates(t *testing.T) {
	completer := &TestTabCompleter{
		Entries: []string{
			"Warm Bed", "Sea of Disease", "Heart of the Inferno",
			"Theta", "Untitled", "Frequency",
			"About Damn Time", "Death Wish Blues", "For The Jeers",
			"Ocean Of Malice",
		},
	}

	sep := " "

	blurredStyle := lipgloss.NewStyle().Strikethrough(true)

	model, err := NewTabCompleter(
		UseCompleter(completer),
		WithSeparator(sep, lipgloss.NewStyle()),
		BlurredStyle(blurredStyle),
		MaxCandidatesToDisplay(3),
	)
	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)
	model = model.moveNextN(5)

	expectedDisplay := DisplayViewHelper{
		Expected:     []string{"Theta", "Untitled", "Frequency"},
		FocusedIndex: 2,
		Separator:    sep,
		BlurredStyle: blurredStyle,
	}

	require.Equal(t, expectedDisplay.String(), model.View())
}

func TestShowsEachPageOfResults(t *testing.T) {
	perPage := 4
	completer := &TestTabCompleter{
		Entries: []string{
			"In The Way", "The Future Says Thank You", "They Fear Us", "Camera Eats First",
			"Cremation Party", "Number Five", "Fluorescent", "You Should Have Gone Back",
			"Hold, Be Held",
		},
	}

	sep := " "

	blurredStyle := lipgloss.NewStyle().Strikethrough(true)

	model, err := NewTabCompleter(
		UseCompleter(completer),
		WithSeparator(sep, lipgloss.NewStyle()),
		BlurredStyle(blurredStyle),
		MaxCandidatesToDisplay(perPage),
	)

	require.Nil(t, err)
	model = model.setCompleteUsing(ARBITRARY_VALUE)

	pages := [][]string{
		{"In The Way", "The Future Says Thank You", "They Fear Us", "Camera Eats First"},
		{"Cremation Party", "Number Five", "Fluorescent", "You Should Have Gone Back"},
		{"Hold, Be Held"},
	}

	for _, page := range pages {
		for i := range page {
			expectedDisplay := DisplayViewHelper{
				Expected:     page,
				FocusedIndex: i,
				Separator:    sep,
				BlurredStyle: blurredStyle,
			}

			require.Equal(t, expectedDisplay.String(), model.View())
			model = model.moveNextN(1)
		}
	}
}

func TestModel_CanSelectEachCandidateFromList(t *testing.T) {
	candidates := []string{
		"Oh What The Future Holds", "Pandora", "Far from Heaven",
		"In Shadows", "Two Towers", "A Higher Level of Hate",
		"Collateral Damage", "Savages", "Conditional Healing",
		"The Man That I Was Not",
	}
	sep := " "
	perPage := 3
	blurredStyle := lipgloss.NewStyle().Strikethrough(true)

	model, err := NewTabCompleter(
		UseCompleter(&TestTabCompleter{
			Entries: candidates,
		}),
		WithSeparator(sep, lipgloss.NewStyle()),
		BlurredStyle(blurredStyle),
		MaxCandidatesToDisplay(perPage),
	)

	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	for _, expectedCandidate := range candidates {
		actualCandidate, _ := model.SelectCurrent()
		require.Equal(t, expectedCandidate, actualCandidate)
		model = model.moveNextN(1)
	}
}

func TestModel_EmptyDisplay_WithNoCandidates(t *testing.T) {
	model, err := NewTabCompleter(
		UseCompleter(&TestTabCompleter{}),
		WithSeparator(" ", lipgloss.NewStyle()),
		MaxCandidatesToDisplay(3),
	)

	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	require.Equal(t, EMPTY_DISPLAY, model.View())
}

func TestModel_NoCurrentToSelect_WithNoCandidates(t *testing.T) {
	model, err := NewTabCompleter(
		UseCompleter(&TestTabCompleter{}),
		WithSeparator(" ", lipgloss.NewStyle()),
		MaxCandidatesToDisplay(3),
	)

	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	actualSelected, err := model.SelectCurrent()

	require.NotNil(t, err)
	require.ErrorIs(t, err, ErrNoCandidates)
	require.Equal(t, "", actualSelected)
}

func TestModel_SuccessfulCompleteClearsError(t *testing.T) {
	candidates := []string{
		"some", "results", "cool",
	}

	model, err := NewTabCompleter(UseCompleter(&TestTabCompleter{
		Entries: candidates,
	}),
	)

	require.Nil(t, err)
	model.Error = &TabError{
		Input: ARBITRARY_VALUE,
		Err:   ErrNoCandidates,
	}

	model = model.setCompleteUsing(ARBITRARY_VALUE)

	require.Nil(t, model.Error)
}

func TestModel_ErrorResult_ClearsCompletion(t *testing.T) {
	completer := &TestTabCompleter{
		Error: ErrCouldNotExpandHome,
	}

	model, err := NewTabCompleter(UseCompleter(completer))
	require.Nil(t, err)

	model = model.setCompleteUsing(ARBITRARY_VALUE)
	require.NotNil(t, model.Error)
	require.Equal(t, EMPTY_DISPLAY, model.View())

	candidates := []string{"dennis", "arthur", "lancelot"}
	completer.Entries = candidates
	completer.Error = nil
	model = model.setCompleteUsing(ARBITRARY_VALUE)

	require.Nil(t, model.Error)
	expectedDisplay := DisplayViewHelper{
		Expected:     candidates,
		FocusedIndex: 0,
		Separator:    " ",
	}
	require.Equal(t, expectedDisplay.String(), model.View())
}

func (m Model) setClear() Model {
	m, _ = m.Update(m.Clear()())
	return m
}

func (m Model) setCompleteUsing(input string) Model {
	m, _ = m.Update(m.Complete(input)())
	return m
}

func (m Model) moveNextN(n int) Model {
	for j := 0; j < n; j++ {
		m, _ = m.Update(m.MoveNext()())
	}

	return m
}

func (m Model) movePrevN(n int) Model {
	for j := 0; j < n; j++ {
		m, _ = m.Update(m.MovePrev()())
	}

	return m
}

type TestTabCompleter struct {
	Entries []string
	Error   error
}

func (t TestTabCompleter) Complete(string) ([]string, error) {
	return t.Entries, t.Error
}

func (t TestTabCompleter) Join(current, selected string) string {
	return current + " " + selected
}

type DisplayViewHelper struct {
	Expected     []string
	Separator    string
	FocusedIndex int
	FocusedStyle lipgloss.Style
	BlurredStyle lipgloss.Style
}

func (d DisplayViewHelper) String() string {
	expectedDisplay := make([]string, len(d.Expected))

	for i, entry := range d.Expected {
		if i == d.FocusedIndex {
			expectedDisplay[i] = d.FocusedStyle.Render(entry)
		} else {
			expectedDisplay[i] = d.BlurredStyle.Render(entry)
		}
	}

	return strings.Join(expectedDisplay, d.Separator)
}

const ARBITRARY_VALUE = ""
const EMPTY_DISPLAY = ""
