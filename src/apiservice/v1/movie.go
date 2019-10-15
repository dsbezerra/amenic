/** DEPRECATED */

package v1

import (
	"strings"
	"time"

	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/dsbezerra/amenic/src/lib/util/scheduleutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Movie struct {
	ID            primitive.ObjectID `json:"_id"`
	ImdbID        string             `json:"imdb_id,omitempty"`
	Title         string             `json:"title,omitempty"`
	OriginalTitle string             `json:"original_title,omitempty"`
	PosterURL     string             `json:"poster_url,omitempty"`
	BackdropURL   string             `json:"backdrop_url,omitempty"`
	Synopsis      string             `json:"synopsis,omitempty"`
	Trailer       string             `json:"trailer_id,omitempty"`
	Cast          []string           `json:"cast,omitempty"`
	Genres        []string           `json:"genres,omitempty"`
	Rating        int                `json:"rating,omitempty"`
	Runtime       int                `json:"runtime,omitempty"`
	Distributor   string             `json:"distributor,omitempty"`
	ReleaseDate   *time.Time         `json:"release_date,omitempty"`
	Showtimes     []Showtime         `json:"showtimes,omitempty"`
	Scores        []models.Score     `json:"scores,omitempty"`
}

// MovieService ...
type MovieService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeMovies ...
func (r *RESTService) ServeMovies(rg *gin.RouterGroup) {
	s := &MovieService{r.data, r.emitter}

	client := rg.Group("/movies", rest.ClientAuth(r.data))
	client.GET("/:id", s.Get)
	client.GET("/:id/showtimes", s.GetSessions)
}

// Get gets the movie corresponding the requested ID.
func (s *MovieService) Get(c *gin.Context) {
	movie, err := s.data.GetMovie(c.Param("id"), s.data.DefaultQuery())
	if movie == nil {
		apiutil.SendNotFound(c)
		return
	}

	scores, _ := s.data.GetScores(s.data.DefaultQuery().AddCondition("movieId", movie.ID))
	movie.Scores = scores

	query := s.data.DefaultQuery()

	period := scheduleutil.GetWeekPeriod(nil)
	period.End = period.End.AddDate(0, 0, 1)
	query.AddCondition("startTime", bson.M{"$gte": period.Start}).
		AddCondition("movieId", movie.ID).
		AddCondition("startTime", bson.M{"$lt": period.End}).
		AddInclude("theater", "movie").
		SetLimit(-1)

	// If we are not sorting let's set the default sort
	query.SetSort("+movieSlug", "+version", "+format", "+startTime")

	sessions, _ := s.data.GetSessions(query)
	movie.Sessions = sessions

	apiutil.SendSuccessOrError(c, mapToMovie(movie), err)
}

// GetSessions gets all showtimes for a given movie
func (s *MovieService) GetSessions(c *gin.Context) {
	query := BuildSessionQuery(s.data, c)
	if ID, err := primitive.ObjectIDFromHex(c.Param("id")); err != nil {
		apiutil.SendBadRequest(c)
		return
	} else {
		query.AddCondition("movieId", ID).
			AddInclude("theater", "movie")
	}
	sessions, err := s.data.GetSessions(query)
	apiutil.SendSuccessOrError(c, mapSessionsTo(sessions), err)
}

// BuildMovieQuery builds movie query from request query string
func BuildMovieQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildMovieQuery(query)
}

func mapToMovie(movie *models.Movie) *Movie {
	if movie == nil {
		return nil
	}

	return &Movie{
		ID:            movie.ID,
		ImdbID:        movie.ImdbID,
		PosterURL:     movie.PosterURL,
		BackdropURL:   movie.BackdropURL,
		Title:         movie.Title,
		OriginalTitle: movie.OriginalTitle,
		Synopsis:      movie.Synopsis,
		Trailer:       movie.Trailer,
		Cast:          movie.Cast,
		Genres:        movie.Genres,
		Rating:        movie.Rating,
		Runtime:       movie.Runtime,
		Distributor:   movie.Distributor,
		ReleaseDate:   movie.ReleaseDate,
		Scores:        movie.Scores,
		Showtimes:     mapSessionsTo(movie.Sessions),
	}
}

// NOTE: isolate to another file
func applyAutoFormatForCloudinaryImage(url string) string {
	base := "res.cloudinary.com/dyrib46is/image/upload"
	index := strings.Index(url, base)
	if index > -1 {
		baseLen := len(base)
		start := url[0 : index+baseLen]
		end := url[index+baseLen:]
		return start + "/f_auto" + end
	}

	return url
}
