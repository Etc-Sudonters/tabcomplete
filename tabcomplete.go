package tabcomplete

type TabCompletion interface {
	Complete(string) ([]string, error)
	Join(string, string) string
}

type TabError struct {
	Input string
	Err   error
}

type Candidate struct {
	value   string
	current bool
}

func (c Candidate) String() string {
	return c.value
}

func (c Candidate) Current() bool {
	return c.current
}

type CandidateNavigator interface {
	CurrentDisplay() []Candidate
	MoveCursorNext()
	MoveCursorPrev()
	SelectCurrent() string
}

type pagedCandidateNavigator struct {
	allCandidates          []string
	totalPages             int
	currentPage            int
	maxCandidatesToDisplay int
	displayCursorIndex     int
}

func newPagedCandidateNavigator(maxCandidatesToDisplay int, allCandidates []string) *pagedCandidateNavigator {
	pcn := &pagedCandidateNavigator{
		allCandidates:          allCandidates,
		maxCandidatesToDisplay: maxCandidatesToDisplay,
		totalPages:             calcTotalPages(maxCandidatesToDisplay, len(allCandidates)),
	}

	return pcn
}

func (p pagedCandidateNavigator) CurrentDisplay() []Candidate {
	leftIndex := max(0, p.maxCandidatesToDisplay*p.currentPage)
	rightIndex := min(len(p.allCandidates), p.maxCandidatesToDisplay*(p.currentPage+1))

	displayView := p.allCandidates[leftIndex:rightIndex]
	candidates := make([]Candidate, len(displayView))

	for i, entry := range displayView {
		candidates[i] = Candidate{
			value:   entry,
			current: i == p.displayCursorIndex,
		}
	}

	return candidates

}

func (p *pagedCandidateNavigator) MoveCursorNext() {
	leftIndex := max(0, p.maxCandidatesToDisplay*p.currentPage)
	rightIndex := min(len(p.allCandidates), p.maxCandidatesToDisplay*(p.currentPage+1))

	pageSize := rightIndex - leftIndex

	if p.displayCursorIndex+1 < pageSize {
		p.displayCursorIndex++
	} else if p.currentPage+1 < p.totalPages {
		p.currentPage++
		p.displayCursorIndex = 0
	}
}

func (p *pagedCandidateNavigator) MoveCursorPrev() {
	if p.displayCursorIndex > 0 {
		p.displayCursorIndex--
	} else if p.currentPage > 0 {
		p.currentPage--
		leftIndex := max(0, p.maxCandidatesToDisplay*p.currentPage)
		rightIndex := min(len(p.allCandidates), p.maxCandidatesToDisplay*(p.currentPage+1))
		pageSize := rightIndex - leftIndex
		p.displayCursorIndex = pageSize - 1
	}
}

func (p pagedCandidateNavigator) SelectCurrent() string {
	currentIndex := p.displayCursorIndex + (p.maxCandidatesToDisplay * p.currentPage)
	clamped := max(0, min(len(p.allCandidates), currentIndex))
	return p.allCandidates[clamped]
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func calcTotalPages(perPage, totalLen int) int {
	pages := totalLen / perPage
	if totalLen%perPage > 0 {
		pages++
	}

	return pages
}
