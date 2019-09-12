package shared

import (
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/sirupsen/logrus"
)

// UpdateTaskStatus ...
func UpdateTaskStatus(id string, data persistence.DataAccessLayer, log *logrus.Entry, runErr error) {
	task, err := data.GetTask(id, data.DefaultQuery())
	if err != nil {
		log.Errorf("couldn't retrieve task with ID: %s", id)
		return
	}

	task.LastRun = time.Now().UTC()

	if runErr != nil {
		task.LastError = runErr.Error()
	} else {
		task.LastError = ""
	}

	_, err = data.UpdateTask(task.ID, *task)
	if err != nil {
		log.Errorf("couldn't update task %s", task.ID)
	}
}
