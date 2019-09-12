package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dsbezerra/amenic/src/lib/env"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic/src/lib/util/fileutil"
	"github.com/dsbezerra/amenic/src/lib/util/scheduleutil"
	"go.mongodb.org/mongo-driver/bson"
)

// StaticType enum-like type to represent each static file.
type StaticType uint

const (
	// StaticTypeHome refers to home.json static file.
	StaticTypeHome StaticType = iota
	// StaticTypeNowPlaying refers to now_playing.json static file.
	StaticTypeNowPlaying
	// StaticTypeUpcoming refers to upcoming.json static file.
	StaticTypeUpcoming
	// StaticTypeSize is used to know the size of these enum-like type.
	StaticTypeSize
)

// SessionType represents the type of sessions some theaters may
// be playing for a now playing movie in the current date.
//
// For example: preview, premiere or normal.
//
// normal   - an ordinary movie session.
// premiere - movie's sessions in the same day as release.
// preview  - movie's sessions before its release.
//
type SessionType uint

const (
	// SessionTypeNone indicates the now playing movie has no session type.
	SessionTypeNone SessionType = iota
	// SessionTypeNormal indicates the session is an ordinary one.
	SessionTypeNormal
	// SessionTypePremiere indicates the session are the very first one.
	// This equals to ESTREIA em pt-BR.
	SessionTypePremiere
	// SessionTypePreview indicates the movie will screen before its release date.
	// This equals to PRÉ-ESTREIA in pt-BR.
	SessionTypePreview
)

type (
	// StaticFile Is used to make data reusable and reduce I/O operations.
	StaticFile struct {
		Filepath string
		Data     []byte
	}

	// Data structure used to create static home.json file.
	staticHome struct {
		NowPlayingPeriod scheduleutil.Period `json:"now_playing_week"`
		Movies           []StaticMovie       `json:"movies"`
	}

	// Data structure used to create static now_playing.json file.
	staticNowPlaying struct {
		Period scheduleutil.Period `json:"period"`
		Movies []StaticMovie       `json:"movies"`
	}
	weekPeriod struct {
		Start *time.Time `json:"start"`
		End   *time.Time `json:"end"`
	}

	// Data structure used to create static upcoming.json file.
	staticUpcoming []StaticMovie

	// Data structure used to retrieve now playing movie from database.
	nowPlayingMovie struct {
		// We need to add this `bson:",inline" to make fields from Movie be processed by mongo`
		models.Movie `bson:",inline"`
		Theaters     []models.Theater `bson:"cinemas"`
	}

	// StaticMovie represents a movie data structure, either now playing or upcoming,
	// used in creation of static files.
	StaticMovie struct {
		Title       string      `json:"title"`
		Poster      string      `json:"poster"`
		ReleaseDate *time.Time  `json:"release_date"`
		Theatres    *string     `json:"theatres"`
		MovieURL    string      `json:"movie_url"`
		SessionType SessionType `json:"session_type,omitempty"`
	}
)

// ClearStatic clears the static file correspoding to the given type.
func ClearStatic(t StaticType) (bool, error) {
	filename := ""

	// Add api version once our users are no longer in APP version <= 1.0.22
	switch t {
	case StaticTypeHome:
		filename = "/home.json"
		break

	case StaticTypeNowPlaying:
		filename = "/movies/now_playing.json"
		break

	case StaticTypeUpcoming:
		filename = "/movies/upcoming.json"
		break
	}

	// NOTE: Temporary added only for back compatibility
	// Remove once our users are no longer in APP version <= 1.0.22
	ok, err := removeStaticFile(filename)
	if err != nil {
		return ok, err
	}

	return removeStaticFile("/v1" + filename)
}

// CreateStatic creates the static file correspoding to the given type for the given API.
func CreateStatic(data persistence.DataAccessLayer, t StaticType) (*StaticFile, error) {
	var r *StaticFile
	var err error

	switch t {
	default:
		// Do nothing.

	case StaticTypeHome:
		r, err = createStaticHome(data)
		break

	case StaticTypeNowPlaying:
		r, err = createStaticNowPlaying(data)
		break

	case StaticTypeUpcoming:
		r, err = createStaticUpcoming(data)
		break
	}

	return r, err
}

// ToStaticType converts a given string to its correspondent type
func ToStaticType(s string) StaticType {
	switch s {
	case "now_playing":
		return StaticTypeNowPlaying
	case "upcoming":
		return StaticTypeUpcoming
	}

	// Defaults to Home
	return StaticTypeHome
}

// Creates home.json static file.
// This requires an existent now_playing.json and upcoming.json files.
func createStaticHome(data persistence.DataAccessLayer) (*StaticFile, error) {
	// Make sure we have updated now_playing and upcoming JSON files.
	p1, err := createStaticNowPlaying(data)
	if err != nil {
		return nil, err
	}

	p2, err := createStaticUpcoming(data)
	if err != nil {
		return nil, err
	}

	var s1 staticNowPlaying
	var s2 staticUpcoming

	err = json.Unmarshal(p1.Data, &s1)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(p2.Data, &s2)
	if err != nil {
		return nil, err
	}

	h := staticHome{NowPlayingPeriod: s1.Period}
	h.Movies = append(s1.Movies, s2...)

	checkForStaticFolders()

	// NOTE: This is temporary.
	result, err := produceStaticFile("/home.json", h)
	if err != nil {
		return nil, err
	}
	result, err = produceStaticFile("/v1/home.json", h)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Creates now_playing.json static file for the v1 API.
func createStaticNowPlaying(data persistence.DataAccessLayer) (*StaticFile, error) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(loc)

	opts := &mongolayer.QueryOptions{
		Includes: []mongolayer.QueryInclude{
			mongolayer.QueryInclude{Field: "cinemas"},
		},
	}
	movies, err := data.OldGetNowPlayingMovies(opts)
	if err != nil {
		return nil, err
	}

	period := scheduleutil.GetWeekPeriod(&now)
	result := staticNowPlaying{Period: *period}

	customSessionType := env.IsEnvVariableTrue("SESSION_TYPE_ENABLED")

	sm := make([]StaticMovie, 0)
	for _, movie := range movies {
		if movie.Hidden {
			continue
		}

		static := StaticMovie{
			Title:       movie.Title,
			Poster:      applyAutoFormatForCloudinaryImage(movie.PosterURL),
			ReleaseDate: movie.ReleaseDate,
			SessionType: SessionTypeNormal,
			MovieURL:    "/m/" + movie.ID.Hex(),
		}

		size := len(movie.Theaters)

		// theatres := ""
		// for i, cinema := range movie.Theaters {
		// 	theatres += cinema.ShortName
		// 	if i < size-1 {
		// 		theatres += " - "
		// 	}
		// }
		// static.Theatres = &theatres

		switch size {
		case 1:
			static.Theatres = &movie.Theaters[0].ShortName
			break
		case 2:
			s := fmt.Sprintf("%s - %s", movie.Theaters[0].ShortName, movie.Theaters[1].ShortName)
			static.Theatres = &s
			break
		default:
		}

		// NOTE(diego):
		// This code depends on movie release date to work correctly.
		// So... we need a crawler to keep these release dates updated for upcoming movies.
		//
		// 6 september 2018
		if customSessionType && movie.ReleaseDate != nil {
			maxDate := movie.ReleaseDate.AddDate(0, 0, 7)

			for _, session := range movie.Sessions {
				if session.StartTime.Before(*movie.ReleaseDate) && now.Before(*movie.ReleaseDate) {
					static.SessionType = SessionTypePreview
					break
				} else {
					// If session is exactly in the premiere week let's set this as
					// premiere
					if now.After(*movie.ReleaseDate) && now.Before(maxDate) {
						static.SessionType = SessionTypePremiere
						// NOTE(diego): Not breaking because sessions may be in any order.
					}
				}
			}
		}

		sm = append(sm, static)
	}

	result.Movies = sm

	checkForStaticFolders()

	// NOTE: Temporary added only for back compatibility
	// Remove once our users are no longer in version <= 1.0.22
	r, err := produceStaticFile("/movies/now_playing.json", result)
	if err != nil {
		return nil, err
	}
	r, err = produceStaticFile("/v1/movies/now_playing.json", result)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Creates upcoming.json static file for the v1 API.
func createStaticUpcoming(data persistence.DataAccessLayer) (*StaticFile, error) {
	opts := &mongolayer.QueryOptions{Conditions: bson.M{}}
	movies, err := data.GetUpcomingMovies(opts)
	if err != nil {
		return nil, err
	}

	s := make(staticUpcoming, len(movies))
	for index, movie := range movies {
		if movie.ID.Hex() == "" {
			continue
		}

		static := StaticMovie{
			Title:       movie.Title,
			Poster:      applyAutoFormatForCloudinaryImage(movie.PosterURL),
			ReleaseDate: movie.ReleaseDate,
			MovieURL:    "/m/" + movie.ID.Hex(),
		}

		s[index] = static
	}

	checkForStaticFolders()

	// NOTE: Temporary added only for back compatibility
	// Remove once our users are no longer in version <= 1.0.22
	r, err := produceStaticFile("/movies/upcoming.json", s)
	if err != nil {
		return nil, err
	}
	r, err = produceStaticFile("/v1/movies/upcoming.json", s)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func exists(filePath string) (exists bool) {
	exists = true

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	return
}

func checkForStaticFolders() {
	staticPath, err := filepath.Abs("./static")
	if err != nil {
		log.Fatal(err)
	}

	v1Path := filepath.Join(staticPath, "v1")
	moviesPath := filepath.Join(staticPath, "movies")

	// Make sure we have ./static directories.
	createDirectory(v1Path)
	createDirectory(v1Path + "/movies")
	createDirectory(moviesPath)
}

func createDirectory(directoryPath string) {
	pathErr := os.MkdirAll(directoryPath, 0777)
	if pathErr != nil {
		fmt.Println(pathErr)
	}
}

func produceStaticFile(filename string, data interface{}) (*StaticFile, error) {
	if filename == "" {
		return nil, errors.New("filename is missing")
	}

	if data == nil {
		return nil, errors.New("data is missing")
	}

	staticPath, err := filepath.Abs("./static")
	if err != nil {
		return nil, err
	}

	fp := filepath.Join(staticPath, filename)
	b, err := fileutil.Struct2Json(fp, data)
	if err != nil {
		return nil, err
	}

	result := &StaticFile{
		Filepath: fp,
		Data:     b,
	}

	return result, nil
}

func removeStaticFile(filename string) (bool, error) {
	if filename == "" {
		return false, errors.New("filename is missing")
	}

	staticPath, err := filepath.Abs("./static")
	if err != nil {
		return false, err
	}

	fp := filepath.Join(staticPath, filename)
	err = os.Remove(fp)
	if err != nil {
		return false, err
	}

	return true, nil
}

func readStaticFile(filepath string, data interface{}) error {
	if filepath == "" {
		return errors.New("invalid filepath")
	}

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	contents := new(bytes.Buffer)
	if _, err := io.Copy(contents, f); err != nil {
		return err
	}

	err = json.Unmarshal(contents.Bytes(), &data)
	if err != nil {
		return err
	}

	return nil
}
