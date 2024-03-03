package consumers

import (
	"context"

	"github.com/sirupsen/logrus"
)

const DefaultGithubWorkerCount int = 4

func GithubDispatcher(ctx context.Context, EventChannel chan string, ProcessorChan chan string) {
	pool := &WorkerPool[string]{
		Pool: make(chan chan string, DefaultGithubWorkerCount*2),
	}

	for i := 0; i < DefaultGithubWorkerCount*2; i++ {
		worker := &Worker[string]{
			Bench:         make(chan string),
			Done:          make(chan bool),
			ProcessorChan: ProcessorChan,
		}

		worker.Start(ctx, pool)

	}

	go func(cx context.Context) {
		for {
			select {
			// wait to receive an event
			case event := <-EventChannel:
				jobChn := <-pool.Pool
				jobChn <- event
			case <-cx.Done():
				logrus.WithContext(ctx).Error("dispatcher quiting, no event consumers")
				return
			}
		}
	}(ctx)
}
