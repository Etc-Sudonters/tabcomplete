package tabcomplete

type TabCompletion interface {
	Complete(string) ([]string, error)
	Join(string, string) string
}

type TabError struct {
	Input string
	Err   error
}

// TODO(ANR): Do we need to also tag the state w/ an Id?
type tabCompleteState struct {
	candidates      []string
	displayView     []string
	candidateCursor int
	displayCursor   int
}

func newTabState(maxCandidatesToDisplay int, candidates []string) *tabCompleteState {
	ts := &tabCompleteState{
		candidates: candidates,
	}

	ts.createDisplayList(maxCandidatesToDisplay)
	return ts
}

func (s *tabCompleteState) createDisplayList(maxCandidatesToDisplay int) {
	if maxCandidatesToDisplay > len(s.candidates) {
		s.displayView = s.candidates
		return
	}

	s.displayView = s.candidates[0:maxCandidatesToDisplay]
}

func (s *tabCompleteState) moveNext() {
	if s.candidateCursor >= len(s.candidates)-1 {
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
