package queue

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
)

// WorkerQueue ...
var WorkerQueue chan chan WorkRequest

// SetupWorkerQueue ...
func SetupWorkerQueue(data persistence.DataAccessLayer, eventEmitter messagequeue.EventEmitter, nworkers int) {
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	for i := 0; i < nworkers; i++ {
		worker := NewWorker(data, eventEmitter, i+1, WorkerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				go func() {
					worker := <-WorkerQueue
					worker <- work
				}()
			}
		}
	}()
}
