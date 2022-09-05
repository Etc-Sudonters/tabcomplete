package tabcomplete

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPagedCandidateNavigator_Starts_WithFirstCandidate_Selected(t *testing.T) {
	perPageSize := 3
	candidates := []string{
		"Aftermath", "Liminal Rite", "Artificial Brain",
		"The Masquerade", "False Light", "Phototroph",
		"Hostile Architecture", "Ars Moriendi", "Elegy",
		"All That Was Promised",
	}

	pcn := newPagedCandidateNavigator(perPageSize, candidates)

	current := pcn.CurrentDisplay()

	require.Equal(t, perPageSize, len(current))

	firstCandidate := current[0]
	assert.True(t, firstCandidate.Current())
	assert.Equal(t, "Aftermath", firstCandidate.String())
}

func TestPagedCandidateNavigator_CanMoveThroughCurrentDisplayView_WithoutChangingPage(t *testing.T) {
	perPageSize := 3
	candidates := []string{
		"Aftermath", "Liminal Rite", "Artificial Brain",
		"The Masquerade", "False Light", "Phototroph",
		"Hostile Architecture", "Ars Moriendi", "Elegy",
		"All That Was Promised",
	}

	pcn := newPagedCandidateNavigator(perPageSize, candidates)

	for i := 0; i < perPageSize; i++ {
		current := pcn.CurrentDisplay()
		require.NotEmpty(t, current)
		candidate := current[i]
		assert.True(t, candidate.Current(), "Expected %d %s to be current item but was not", i, candidate.String())
		assert.Equal(t, candidates[i], candidate.String())
		pcn.MoveCursorNext()
	}
}

func TestPagedCandidateNavigator_CanMoveToNextPage(t *testing.T) {
	perPageSize := 3
	candidates := []string{
		"Aftermath", "Liminal Rite", "Artificial Brain",
		"The Masquerade", "False Light", "Phototroph",
		"Hostile Architecture", "Ars Moriendi", "Elegy",
		"All That Was Promised",
	}

	pcn := newPagedCandidateNavigator(perPageSize, candidates)

	for i := 0; i < perPageSize; i++ {
		pcn.MoveCursorNext()
	}

	current := pcn.CurrentDisplay()

	require.Equal(t, perPageSize, len(current))
	firstCandidate := current[0]
	assert.True(t, firstCandidate.Current())
	assert.Equal(t, "The Masquerade", firstCandidate.String())
}

func TestPagedCandidateNavigator_Can_MoveNextThroughEntireCandidateList(t *testing.T) {
	perPageSize := 3
	candidates := []string{
		"Aftermath", "Liminal Rite", "Artificial Brain",
		"The Masquerade", "False Light", "Phototroph",
		"Hostile Architecture", "Ars Moriendi", "Elegy",
		"All That Was Promised",
	}

	pcn := newPagedCandidateNavigator(perPageSize, candidates)

	for i := 0; i < len(candidates); i++ {
		expectedCandidate := candidates[i]
		actualCandidate := pcn.SelectCurrent()
		require.Equal(t, expectedCandidate, actualCandidate)
		pcn.MoveCursorNext()
	}
}

func TestPagedCandidateNavigator_CapsDisplay_If_LessTotalCandidates_ThanMaxCandidates(t *testing.T) {
	candidates := []string{
		"otherness", "heroine",
	}
	perPage := 3

	pcn := newPagedCandidateNavigator(perPage, candidates)

	pcn.MoveCursorNext()
	pcn.MoveCursorNext()
	pcn.MoveCursorNext()

	require.Equal(t, "heroine", pcn.SelectCurrent())

	current := pcn.CurrentDisplay()

	require.Equal(t, 2, len(current))
	assert.Equal(t, "otherness", current[0].String())
	assert.False(t, current[0].Current())
	assert.Equal(t, "heroine", current[1].String())
	assert.True(t, current[1].Current())
}

func TestPagedCandidateNavigator_CanMovePreviouslyThroughCandidates(t *testing.T) {
	candidates := []string{
		"the hellfire club", "leather wings", "blue velvet",
	}
	perPage := 3

	pcn := newPagedCandidateNavigator(perPage, candidates)

	for i := 0; i < perPage; i++ {
		pcn.MoveCursorNext()
	}

	for i := perPage - 1; i >= 0; i-- {
		currentDisplay := pcn.CurrentDisplay()
		candidate := currentDisplay[i]
		assert.True(t, candidate.Current())
		assert.Equal(t, candidates[i], candidate.String())
		pcn.MoveCursorPrev()
	}
}

func TestPagedCandidateNavigator_CanMovePreviously_ThroughEntireCandidateList(t *testing.T) {
	candidates := []string{
		"The Darkest Burden", "Broken Maze", "Behind Closed Doors", "When Talking Fails, It's Time For Violence!",
		"Your Dystopian Hell", "Unrecognizable", "Hatred Transcending", "Doomsayer",
		"Pale Moonlight", "Seizures", "Voiceless Choir", "Grieve",
		"Sea of Disease", "Noxious Cloud", "Shattered Faith", "Desolate Landscapes",
		"Spiral Eyes", "Vicious Circle", "Weeping Willow", "All Will Wither",
		"Glass Shards",
	}
	perPage := 4
	pcn := newPagedCandidateNavigator(perPage, candidates)

	for range candidates {
		pcn.MoveCursorNext()
	}

	require.Equal(t, candidates[len(candidates)-1], pcn.SelectCurrent())

	for i := len(candidates) - 1; i >= 0; i-- {
		expectedCandidate := candidates[i]
		actualCandidate := pcn.SelectCurrent()
		require.Equal(t, expectedCandidate, actualCandidate)
		pcn.MoveCursorPrev()
	}
}
