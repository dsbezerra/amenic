package premiere

import (
	"fmt"
	"strings"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	Type  = "premiere"
	Title = "Estreias da semana"
)

// PrepareNotification ...
func PrepareNotification(releases []models.Movie) *models.Notification {
	if len(releases) == 0 {
		return nil
	}

	notification := &models.Notification{
		ID:        primitive.NewObjectID(),
		Type:      Type,
		Title:     Title,
		CreatedAt: time.Now().UTC(),
	}

	text := strings.Builder{}
	htmlText := strings.Builder{}

	size := len(releases)

	for mIndex, m := range releases {
		cinemas := strings.Builder{}

		cSize := len(m.Theaters)
		if cSize == 1 {
			cinemas.WriteString(m.Theaters[0].Name)
		} else if cSize > 1 {
			for cIndex, c := range m.Theaters {
				cinemas.WriteString(c.Name)
				if cIndex < cSize-1 {
					cinemas.WriteString(", ")
				}
			}
		}

		t := fmt.Sprintf("%s (%s)", m.Title, cinemas.String())
		h := fmt.Sprintf("<b>%s</b> (<i>%s</i>)", m.Title, cinemas.String())

		text.WriteString(t)
		htmlText.WriteString(h)

		if mIndex < size-1 {
			text.WriteString("\n")
			htmlText.WriteString("<br/>")
		}

		if notification.Data == nil {
			notification.Data = &models.NotificationData{}
		}

		notification.Data.Movies = append(notification.Data.Movies, models.Movie{
			ID:            m.ID,
			Title:         m.Title,
			OriginalTitle: m.OriginalTitle,
			PosterURL:     m.PosterURL,
			Genres:        m.Genres,
			Runtime:       m.Runtime,
			Trailer:       m.Trailer,
		})
	}

	notification.Single = len(notification.Data.Movies) == 1
	if notification.Single {
		notification.ItemID = notification.Data.Movies[0].ID.Hex()
	}

	notification.Text = text.String()
	notification.HTMLText = htmlText.String()

	return notification
}
