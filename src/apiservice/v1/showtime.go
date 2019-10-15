/** DEPRECATED */

package v1

import (
	"fmt"
	"sort"
	"time"

	"github.com/dsbezerra/amenic/src/lib/util/scheduleutil"
	"github.com/dsbezerra/amenic/src/lib/util/timeutil"

	"github.com/dsbezerra/amenic-api/showtimeutil"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

type Showtime struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	MovieID     primitive.ObjectID `json:"movie_id" bson:"movieId,omitempty"`
	CinemaID    string             `json:"cinema_id" bson:"cinemaId,omitempty"`
	Format      string             `json:"format" bson:"format"`
	Version     string             `json:"version" bson:"version"`
	OpeningTime string             `json:"opening_time,omitempty" bson:"time"`
	RoomNumber  int                `json:"room_number,omitempty" bson:"room"`
	Weekday     int                `json:"weekday" bson:"weekday"`
	Period      struct {
		Start time.Time `json:"start,omitempty" bson:"start,omitempty"`
		End   time.Time `json:"end,omitempty" bson:"end,omitempty"`
	} `json:"period,omitempty" bson:"period,omitempty"`
	Cinema *Cinema `json:"cinema,omitempty" bson:"cinema,omitempty"`
	Movie  *Movie  `json:"movie,omitempty" bson:"movie,omitempty"`
}

// ShowtimeService ...
type ShowtimeService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeShowtimes ...
func (r *RESTService) ServeShowtimes(rg *gin.RouterGroup) {
	s := &ShowtimeService{r.data, r.emitter}

	client := rg.Group("/showtimes", rest.ClientAuth(r.data))
	client.GET("", s.GetAll)
}

// GetAll gets all showtimes.
func (s *ShowtimeService) GetAll(c *gin.Context) {
	q := c.MustGet("query_options").(map[string]string)
	getSessions(c, q["cinema"], s.data, "{theater,movie{title\\nposter\\nrating}}")
}

func getSessions(c *gin.Context, cinema string, data persistence.DataAccessLayer, include string) {
	fq := make(map[string]string, 0)
	if cinema != "" {
		findTheaterQuery := data.DefaultQuery()
		if cinema == "cinemais-34" {
			findTheaterQuery.AddCondition("internalId", "34")
		} else {
			findTheaterQuery.AddCondition("internalId", cinema)
		}
		cin, _ := data.FindTheater(findTheaterQuery)
		if cin == nil {
			apiutil.SendNotFound(c)
			return
		}
		fq["theaterId"] = cin.ID.Hex()
	}

	q := c.MustGet("query_options").(map[string]string)
	if movie := q["movie"]; movie != "" {
		fq["movieId"] = movie
	}

	if include != "" {
		fq["include"] = include
	}
	query := data.BuildSessionQuery(fq)

	// We require at least movie or cinema
	if query.GetCondition("movieId") == nil && query.GetCondition("theaterId") == nil {
		apiutil.SendBadRequest(c)
		return
	}

	period := scheduleutil.GetWeekPeriod(nil)
	period.End = period.End.AddDate(0, 0, 1)
	query.AddCondition("startTime", bson.M{"$gte": period.Start}).
		AddCondition("startTime", bson.M{"$lt": period.End}).
		SetLimit(-1)

	// If we are not sorting let's set the default sort
	query.SetSort("+movieSlug", "+version", "+format", "+startTime")

	sessions, err := data.GetSessions(query)
	apiutil.SendSuccessOrError(c, mapSessionsTo(sessions), err)
}

func mapSessionsTo(sessions []models.Session) []Showtime {
	if sessions == nil {
		return nil
	}

	var result []Showtime

	m := make(map[string][]Showtime, 0)

	var loc *time.Location

	for _, session := range sessions {
		if session.Theater == nil || session.Movie == nil {
			continue
		}
		if session.StartTime == nil {
			continue
		}
		if session.MovieID.IsZero() {
			continue
		}

		if session.Version == "subtitled" {
			session.Version = "subbed"
		}

		if loc == nil || loc.String() != session.TimeZone {
			loc, _ = time.LoadLocation(session.TimeZone)
		}

		offtime := session.StartTime.In(loc)
		ot := fmt.Sprintf("%02d:%02d", offtime.Hour(), offtime.Minute())
		sod := timeutil.StartOfDayForTime(&offtime)
		showtime := Showtime{
			ID:          session.ID,
			MovieID:     session.MovieID,
			CinemaID:    session.Theater.InternalID,
			RoomNumber:  int(session.Room),
			OpeningTime: ot,
			Format:      session.Format,
			Version:     session.Version,
			Weekday:     int(models.TimeWeekdayToWeekday(offtime.Weekday())),
			Period: struct {
				Start time.Time `json:"start,omitempty" bson:"start,omitempty"`
				End   time.Time `json:"end,omitempty" bson:"end,omitempty"`
			}{Start: sod, End: sod},
			Movie: &Movie{
				ID:        session.MovieID,
				Title:     session.Movie.Title,
				PosterURL: session.Movie.PosterURL,
				Rating:    session.Movie.Rating,
			},
		}

		if showtime.CinemaID == "34" {
			showtime.CinemaID = "cinemais-34"
		}

		k := fmt.Sprintf("%s-%d-%s-%s-%s", session.MovieSlug, showtime.RoomNumber, showtime.Format, showtime.Version, showtime.OpeningTime)
		_, ok := m[k]
		if !ok {
			m[k] = make([]Showtime, 0)
		}
		m[k] = append(m[k], showtime)
	}

	// Sort and loop...
	keys := make([]string, len(m))
	index := 0
	for k := range m {
		keys[index] = k
		index++
	}
	sort.Strings(keys)

	weekPeriod := showtimeutil.GetWeekPeriod(nil)
	for _, k := range keys {
		v := m[k]
		// Showtimes for all days
		if len(v) == 7 {
			first := v[0]
			first.Weekday = int(models.ALL)
			first.Period = struct {
				Start time.Time `json:"start,omitempty" bson:"start,omitempty"`
				End   time.Time `json:"end,omitempty" bson:"end,omitempty"`
			}{Start: weekPeriod.Start, End: weekPeriod.End}
			result = append(result, first)
		} else {
			for _, each := range v {
				if each.Weekday == int(models.INVALID) {
					continue
				}
				result = append(result, each)
			}
		}
	}
	return result
}

// BuildSessionQuery builds showtime query from request query string
func BuildSessionQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildSessionQuery(query)
}
