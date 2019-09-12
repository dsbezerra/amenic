package task

import (
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/notificationservice/fcm"
	"github.com/dsbezerra/amenic/src/notificationservice/premiere"
	"github.com/pkg/errors"
)

var (
	// ErrNoMoviesPlaying ...
	ErrNoMoviesPlaying = errors.New("there are no movies currently playing")
)

// CheckOpeningMovies simply checks for release movies in the
// week of its execution and sends a notification.
func CheckOpeningMovies(data persistence.DataAccessLayer) error {
	releases, err := GetWeekReleases(data)
	if err != nil {
		return err
	}

	notification := premiere.PrepareNotification(releases)
	if notification != nil {
		_, err := fcm.SendNotification(notification)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetWeekReleases ...
func GetWeekReleases(data persistence.DataAccessLayer) ([]models.Movie, error) {
	result := make([]models.Movie, 0)

	playing, err := data.GetNowPlayingMovies(data.DefaultQuery().AddInclude("cinemas"))
	if err != nil {
		return nil, err
	}

	if len(playing) == 0 {
		return nil, ErrNoMoviesPlaying
	}

	// Now we check which movies releases in the current now playing week with a limit of 2 days
	// because we don't want to generate a notification after the weekend.
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	for _, m := range playing {
		if m.Hidden || m.ReleaseDate == nil {
			continue
		}

		// Checks if movie release date is between the current now playing week.
		// Maximmum time to notify is two days later
		if now.Sub(*m.ReleaseDate) <= time.Hour*24*2 {
			result = append(result, m)
		}
	}

	return result, nil
}
