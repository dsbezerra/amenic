package moviescore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRottenSearch(t *testing.T) {
	query := "iron man"

	rotten := NewRottenTomatoes()
	result, err := rotten.Search(query)
	assert.NoError(t, err)

	expectedMovieInList := ResultItem{
		ID:    "/m/iron_man",
		Title: "Iron Man",
		Year:  2008,
	}
	assert.NotEmpty(t, result.Items)

	expectedFound := false
	for _, movie := range result.Items {
		expectedFound = isResultItemEqual(movie, expectedMovieInList)
		if expectedFound {
			break
		}
	}
	assert.Equal(t, true, expectedFound)
}

func TestRottenScore(t *testing.T) {
	path := "/m/sharknado_2013"
	rotten := NewRottenTomatoes()
	result, err := rotten.Score(path)
	assert.NoError(t, err)
	assert.Len(t, result.Items, 1)

	item := result.Items[0]
	assert.NotEqual(t, item.Score, 0.0)
	assert.True(t, isScoreClassOneOf(item.ScoreClass))
}

func TestEnsureHasPathM(t *testing.T) {
	expected := "/m/sharknado_2013"

	path := "sharknado_2013"
	path = ensurePathHasM(path)
	assert.Equal(t, expected, path)

	path = "/sharknado_2013"
	path = ensurePathHasM(path)
	assert.Equal(t, expected, path)

	path = "m/sharknado_2013"
	path = ensurePathHasM(path)
	assert.Equal(t, expected, path)

	path = "/m/sharknado_2013"
	path = ensurePathHasM(path)
	assert.Equal(t, expected, path)
}

func isResultItemEqual(a ResultItem, b ResultItem) bool {
	return (a.Title == b.Title &&
		a.ID == b.ID &&
		a.Year == b.Year)
}

func isScoreClassOneOf(scoreClass string) bool {
	return scoreClass == "rotten" || scoreClass == "fresh" || scoreClass == "certified_fresh"
}
