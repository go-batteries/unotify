package consumers

import (
	"context"

	"github.com/sirupsen/logrus"
)

type WorkerPool[E any] struct {
	Pool chan chan E
	// Pool chan chan resque.Payload
}

type WorkerProcessor func(context.Context, string) (any, error)

type Worker[E any] struct {
	Bench         chan E
	Done          chan bool
	ProcessorChan chan E
}

func (w Worker[E]) Start(ctx context.Context, pool *WorkerPool[E]) {
	go func() {
		for {
			// send the job channel
			pool.Pool <- w.Bench

			select {
			// wait for the data to arrive
			case payload := <-w.Bench:
				logrus.WithContext(ctx).Debugf("payload received from github %+v. sending to process\n", payload)
				w.ProcessorChan <- payload
			case <-w.Done:
				logrus.WithContext(ctx).Infoln("worker asked to stop")
				return
			}
		}
	}()
}
