package listener

import (
	"log"

	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/imageservice/cloudinary"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventProcessor ...
type EventProcessor struct {
	EventListener messagequeue.EventListener
	EventEmitter  messagequeue.EventEmitter
	Data          persistence.DataAccessLayer
	Log           *logrus.Entry
}

// ProcessEvents ...
func (p *EventProcessor) ProcessEvents() error {
	log.Println("Listening to events...")

	var eventsList = []string{
		"imageUpload",
		"movieDeleted",
	}

	received, errors, err := p.EventListener.Listen(eventsList...)
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-received:
			p.handle(event)
		case err = <-errors:
			log.Printf("received error while processing message: %s", err)
		}
	}
}

func (p *EventProcessor) handle(event messagequeue.Event) {
	switch event.(type) {
	case *contracts.EventImageUpload:
		p.handleImageUpload(event.(*contracts.EventImageUpload))
	case *contracts.EventMovieDeleted:
		p.handleMovieDeleted(event.(*contracts.EventMovieDeleted))
	default:
		log.Printf("unknown event: %t", event)
	}
}

func (p *EventProcessor) handleImageUpload(e *contracts.EventImageUpload) {
	movieID, err := primitive.ObjectIDFromHex(e.MovieID)
	if err != nil {
		p.Log.Errorf("Aborting image upload because '%s' is not a valid movie id", e.MovieID)
		return
	}

	// NOTE(diego): Defaulting to Cloudinary for now.
	im, err := cloudinary.UploadWebImage(e.URL, e.ImageType)
	if err != nil {
		p.Log.Errorf("Error occurred while uploading image '%s'", e.URL)
		return
	}

	im.MovieID = movieID
	err = p.Data.InsertImage(*im)
	if err != nil {
		p.Log.Errorf("Error occurred while inserting image '%s'", e.URL)
		return
	}
}

func (p *EventProcessor) handleMovieDeleted(e *contracts.EventMovieDeleted) {
	images, err := p.Data.GetMovieImages(e.MovieID, p.Data.DefaultQuery())
	if err != nil {
		// TODO(diego): Persist this event so we can handle it later?
		p.Log.Errorf("Error occurred while getting movie '%s' images", e.MovieID)
		return
	}

	imageIdsToDelete := []string{}

	// NOTE(diego): We could remove many images concurrently but let's keep it simple for now
	for _, im := range images {
		// TODO(diego): Add host in image structure if we add support for amazon s3
		if im.SecureURL == "" {
			continue
		}
		err := cloudinary.DeleteImage(im.SecureURL)
		if err != nil {
			p.Log.Errorf("Error '%s' occurred while deleting image '%s'", err.Error(), im.SecureURL)
		} else {
			imageIdsToDelete = append(imageIdsToDelete, im.ID.Hex())
		}
	}

	size := len(imageIdsToDelete)
	if size > 0 {
		count, err := p.Data.DeleteImagesByIDs(imageIdsToDelete)
		if err != nil {
			p.Log.Errorf("Error '%s' occurred while deleting images", err.Error())
		} else {
			if count == int64(size) {
				p.Log.Infof("Expected images were successfully deleted")
			} else {
				p.Log.Warnf("Expected to delete %d images, but deleted %d", size, count)
			}
		}
	}
}
