package fcm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/notificationservice/premiere"
)

// NotificationPostBody is used to post notification to FCM servers
type NotificationPostBody struct {
	To         string               `json:"to"`
	Data       *models.Notification `json:"data,omitempty"`
	TimeToLive int                  `json:"time-to-live,omitempty"`
}

const (
	// MinTTL is the minimum value used for the time to live field of FCM service
	// represented in hours.
	MinTTL = 0

	// MaxTTL is the default maximum time the notification will live to be sent to our users.
	MaxTTL = 86400 * 7 // 1 week

	// Endpoint is the send notification URI for FCM.
	Endpoint = "https://fcm.googleapis.com/fcm/send"
)

var (
	ErrInvalidNotification = errors.New("notification is invalid")
	ErrAuthKeyUndefined    = errors.New("authorization key not defined")
)

// SendNotification ...
func SendNotification(notification *models.Notification) (bool, error) {
	if notification == nil {
		return false, ErrInvalidNotification
	}
	authKey := os.Getenv("FCM_AUTH_KEY")
	if authKey == "" {
		return false, ErrAuthKeyUndefined
	}

	release := os.Getenv("AMENIC_MODE") == "release"

	topic := notification.Type
	if !release {
		topic = "development"
	}

	ttl := MinTTL

	data := *notification
	switch data.Type {
	case premiere.Type:
		// We copy notification data here to remove data payload because
		// this field is being used only to retrieve additional information
		// about the notification that shouldn't be sent over FCM service.
		data.Data = nil
		// Let's use half a week, which will make possible to users receive this notification
		// before next Monday (if this code run in a Thursday)
		ttl = MaxTTL / 2
	default:
		ttl = MaxTTL
	}

	notificationBody := NotificationPostBody{
		To:         "/topics/" + topic,
		Data:       &data,
		TimeToLive: ttl,
	}

	buff, err := json.Marshal(notificationBody)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", Endpoint, bytes.NewBuffer(buff))
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "key="+authKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	return success, nil
}
