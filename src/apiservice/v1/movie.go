package v1

import (
	"strings"

	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// MovieService ...
type MovieService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeMovies ...
func (r *RESTService) ServeMovies(rg *gin.RouterGroup) {
	s := &MovieService{r.data, r.emitter}

	client := rg.Group("/movies", rest.ClientAuth(r.data))
	client.GET("/movie/:id", s.Get)
	client.GET("/movie/:id/showtimes", s.GetSessions)
	client.GET("/movie/:id/sessions", s.GetSessions)
	client.GET("/now_playing", s.GetNowPlaying)
	client.GET("/upcoming", s.GetUpcoming)

	admin := rg.Group("/movies", rest.AdminAuth(r.data))
	admin.GET("", s.GetAll)
	admin.GET("/count", s.Count)
	admin.PUT("/movie/:id", s.Update)
	admin.DELETE("/movie/:id", s.Delete)
}

// GetNowPlaying gets all now playing movies
func (s *MovieService) GetNowPlaying(c *gin.Context) {
	// NOTE: We use SessionQuery because in order to retrieve now playing movies
	// we need to perform an agreggation on Sessions collection
	movies, err := s.data.GetNowPlayingMovies(BuildSessionQuery(s.data, c))
	apiutil.SendSuccessOrError(c, movies, err)
}

// GetUpcoming gets all upcoming movies
func (s *MovieService) GetUpcoming(c *gin.Context) {
	query := c.MustGet("query_options").(map[string]string)
	movies, err := s.data.GetUpcomingMovies(s.data.BuildMovieQuery(query))
	apiutil.SendSuccessOrError(c, movies, err)
}

// Get gets the movie corresponding the requested ID.
func (s *MovieService) Get(c *gin.Context) {
	movie, err := s.data.GetMovie(c.Param("id"), BuildMovieQuery(s.data, c))
	apiutil.SendSuccessOrError(c, movie, err)
}

// GetAll gets all movies.
func (s *MovieService) GetAll(c *gin.Context) {
	movies, err := s.data.GetMovies(BuildMovieQuery(s.data, c))
	apiutil.SendSuccessOrError(c, movies, err)
}

// GetSessions gets all showtimes for a given movie
func (s *MovieService) GetSessions(c *gin.Context) {
	query := BuildSessionQuery(s.data, c).
		AddCondition("movieId", c.Param("id"))
	if cinema := c.Query("cinema"); cinema != "" {
		query.AddCondition("cinemaId", cinema)
	}
	showtimes, err := s.data.GetSessions(query)
	apiutil.SendSuccessOrError(c, showtimes, err)
}

// Count returns the total count of Movie matching the given query
func (s *MovieService) Count(c *gin.Context) {
	count, err := s.data.CountMovies(BuildMovieQuery(s.data, c))
	apiutil.SendSuccessOrError(c, count, err)
}

// Update apply to movie with the given ID the given body data
func (s *MovieService) Update(c *gin.Context) {
	movie := models.Movie{}
	err := c.ShouldBindJSON(&movie)
	if err != nil {
		apiutil.SendBadRequest(c)
		return
	}
	_, err = s.data.UpdateMovie(c.Param("id"), movie)
	apiutil.SendSuccessOrError(c, movie, err)
}

// Delete the movie with the given ID
func (s *MovieService) Delete(c *gin.Context) {
	err := s.data.DeleteMovie(c.Param("id"))
	// TODO: Emit movie deleted event if successful
	apiutil.SendSuccessOrError(c, 1, err)
}

// BuildMovieQuery builds movie query from request query string
func BuildMovieQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildMovieQuery(query)
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
