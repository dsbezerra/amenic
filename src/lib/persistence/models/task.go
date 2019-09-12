package models

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dsbezerra/amenic/src/lib/util/timeutil"
)

const (
	TaskCreateStatic       = "create_static"
	TaskSyncScores         = "sync_scores"
	TaskCheckOpeningMovies = "check_opening_movies"
	TaskStartScraper       = "start_scraper"
)

// Task is a single unit of work for service to perform.
type Task struct {
	ID          string    `json:"id" bson:"_id"`                    // Task identifier.
	Service     string    `json:"service" bson:"service"`           // Service that will run the task.
	Type        string    `json:"type" bson:"type"`                 // Task type.
	Name        string    `json:"name" bson:"name"`                 // Name is the name of the Task.
	Description string    `json:"description" bson:"description"`   // Task friendly description.
	Args        []string  `json:"args" bson:"args"`                 // Task alternative arguments.
	Cron        []string  `json:"cron" bson:"cron"`                 // Cron pattern of the Task.
	LastRun     time.Time `json:"last_run" bson:"last_run"`         // LastRun is the last time that the Task was executed.
	LastError   string    `json:"last_error" bson:"last_error"`     // LastError is the error message or stack trace from the last time the task failed.
	Enabled     bool      `json:"enabled" bson:"enabled"`           // Whether this task is enabled or should be executed.
	RunAtStart  bool      `json:"run_at_start" bson:"run_at_start"` // Whether this task should run at start or not.
}

// IsValid checks if an instance of Task is valid.
// It's valid whenever we have cron, service and type defined.
func (m *Task) IsValid() bool {
	// NOTE: We assume cron is a valid expression.
	return len(m.Cron) != 0 && m.Service != "" && m.Type != ""
}

// LastRunFormatted retrieves LastRun formatted in Mon Jan _2 15:04:05 2006 layout.
func (m *Task) LastRunFormatted() string {
	if m.LastRun.IsZero() {
		return ""
	}
	return timeutil.FriendlyFormat(&m.LastRun, true)
}

// GenerateID builds an string from fields:
// Type, Args and generates a md5 hash
//
// Like: Type Args joined by space
func (m *Task) GenerateID() (bool, string) {
	ok, ID := GenerateTaskID(m.Type, m.Args)
	if ok {
		m.ID = ID
		return true, ID
	}

	return false, ""
}

// GenerateTaskID ...
func GenerateTaskID(_type string, args []string) (bool, string) {
	if _type == "" {
		return false, ""
	}

	b := strings.Builder{}
	b.WriteString(_type)
	if len(args) != 0 {
		b.WriteString(" ")
		b.WriteString(strings.Join(args, " "))
	}

	h := md5.New()
	io.WriteString(h, b.String())
	return true, fmt.Sprintf("%x", h.Sum(nil))
}
