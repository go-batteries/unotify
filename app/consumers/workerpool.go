package consumers

import (
	"context"

	"github.com/sirupsen/logrus"
)

type WorkerPool struct {
	Pool chan chan string
	// Pool chan chan resque.Payload
}

type Worker struct {
	Bench chan string
	Done  chan bool
}

func (w Worker) Start(ctx context.Context, pool *WorkerPool) {
	go func() {
		for {
			// send the job channel
			pool.Pool <- w.Bench

			select {
			// wait for the data to arrive
			case payload := <-w.Bench:
				logrus.Printf("payload from github %+v\n", payload)
			case <-w.Done:
				logrus.Infoln("worker asked to stop")
				return
			}
		}
	}()
}
