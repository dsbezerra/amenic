package queue

import (
	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/scraperservice/task"
)

func NewWorker(data persistence.DataAccessLayer, eventEmitter messagequeue.EventEmitter, id int, workerQueue chan chan WorkRequest) Worker {
	worker := Worker{
		Data:         data,
		EventEmitter: eventEmitter,
		ID:           id,
		Work:         make(chan WorkRequest),
		WorkerQueue:  workerQueue,
		QuitChan:     make(chan bool)}

	return worker
}

type Worker struct {
	Data         persistence.DataAccessLayer
	EventEmitter messagequeue.EventEmitter
	ID           int
	Work         chan WorkRequest
	WorkerQueue  chan chan WorkRequest
	QuitChan     chan bool
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work
			select {
			case work := <-w.Work:
				opts := task.ScraperOptions{
					ScraperID:     work.ScraperID,
					IgnoreLastRun: work.IgnoreLastRun,
				}
				run, err := task.StartScraper(w.Data, opts)
				if err != nil {
					// p.Log.Errorln(err.Error())
					return
				}
				w.EventEmitter.Emit(&contracts.EventScraperFinished{
					Type:      run.Scraper.Type,
					ScraperID: opts.ScraperID,
				})
			case <-w.QuitChan:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
