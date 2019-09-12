package mongolayer

import (
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMovie(t *testing.T) {
	data, err := getTestingMongoDAL()
	defer data.Close()
	assert.NoError(t, err)

	doc := models.Movie{
		ID:    primitive.NewObjectID(),
		Title: "some-movie-title",
	}

	opts := DefaultOptions("")

	// Insert
	err = data.InsertMovie(doc)
	assert.NoError(t, err)

	// Find one with conditions
	movie, err := data.FindMovie(DefaultOptions("").AddCondition("title", doc.Title))
	assert.NoError(t, err)
	assert.NotEmpty(t, movie)

	// Find one by id
	movie, err = data.GetMovie(doc.ID.Hex(), opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, movie)

	// Find many
	movies, err := data.GetMovies(opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, movies)

	// Update
	update := doc
	update.Title = "updated-title"
	modified, err := data.UpdateMovie(doc.ID.Hex(), update)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), modified)

	movie, err = data.GetMovie(doc.ID.Hex(), opts)
	assert.NoError(t, err)
	assert.Equal(t, movie.Title, update.Title)

	// Delete
	err = data.DeleteMovie(doc.ID.Hex())
	assert.NoError(t, err)

	// GetNowPlaying
	movies, err = data.GetNowPlayingMovies(data.DefaultQuery())
	assert.NoError(t, err)
	assert.NotEmpty(t, movies)
}
