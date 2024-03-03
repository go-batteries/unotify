package workerpool

import (
	"context"
	"fmt"
	"regexp"
	"unotify/app/core/events"
	"unotify/app/pkg/ds"

	"github.com/sirupsen/logrus"
)

type Result struct {
	Err  error
	Data any
}
type WorkerProcessor[E any] func(context.Context, E) Result

type WorkerPool[E any] struct {
	Pool    chan chan E
	DoneCh  chan error
	Workers []Worker[E]
}

func NewWorkerPool[E any](
	ctx context.Context,
	poolSize int,
	processor WorkerProcessor[E],
) *WorkerPool[E] {
	workerPool := &WorkerPool[E]{
		Pool: make(chan chan E, poolSize),
		// DoneCh: make(chan error),
	}

	workers := []Worker[E]{}

	for i := 0; i < poolSize; i++ {
		worker := Worker[E]{
			Bench: make(chan E, poolSize),
			// Done:      make(chan bool),
			Processor: processor,
		}

		workers = append(workers, worker)
	}

	workerPool.Workers = workers

	for i := 0; i < poolSize; i++ {
		worker := workers[i]
		worker.Pool = workerPool

		// This context should close the workers
		worker.Start(ctx)
	}

	return workerPool
}

func (wp *WorkerPool[E]) Start(ctx context.Context, ReceiverChan chan E) {
	go func(cx context.Context) {
		for {
			select {
			// wait to receive an event
			case event := <-ReceiverChan:
				jobChn := <-wp.Pool
				jobChn <- event
			// case <-wp.DoneCh:
			// 	logrus.WithContext(ctx).Infoln("dispatcher quiting, closed by user")
			// 	close(wp.DoneCh)
			// 	return
			case <-cx.Done():
				logrus.WithContext(ctx).Infoln("dispatcher quiting, no event consumers")
				return
			}
		}
	}(ctx)
}

func (wp WorkerPool[E]) Stop(ctx context.Context) error {
	logrus.Printf("stopping worker pool")

	// for i := range wp.Workers {
	// 	worker := wp.Workers[i]
	// 	worker.Done <- true
	// }

	// logrus.Printf("stopped workers")

	// wp.DoneCh <- nil
	// logrus.Printf("stopped worker pool")

	return nil
}

type Worker[E any] struct {
	Bench     chan E
	Done      chan bool
	Pool      *WorkerPool[E]
	Processor WorkerProcessor[E]
}

func (w Worker[E]) Start(ctx context.Context) {
	go func() {
		for {
			// send the job channel
			w.Pool.Pool <- w.Bench

			select {
			// wait for the data to arrive
			case payload := <-w.Bench:
				logrus.WithContext(ctx).Printf("payload received by worker processor %+v\n", payload)
				w.ProcessGithubPayload(ctx, fmt.Sprintf("%v", payload))
			case <-ctx.Done():
				logrus.WithContext(ctx).Infoln("stopped by context")
				return
				// case <-w.Done:
				// 	logrus.WithContext(ctx).Infoln("worker asked by worker pool")
				// 	return
			}
		}
	}()
}

func (w Worker[E]) ProcessGithubPayload(ctx context.Context, payload string) error {
	// Here payload is github events
	event, err := events.ParseGithubEvenFromCache([]byte(payload))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to marshal github event")
		return err
	}

	logrus.WithContext(ctx).Infof("github event %+v\n", event.Release.Body)

	// Read the event Body, to Find Jira Tickets
	re := regexp.MustCompile(`[A-Z]+-\d+`)
	matches := re.FindAllString(event.Release.Body, -1)

	issues := ds.ToSet(matches...)
	iter := issues.Iter()

	logrus.WithContext(ctx).Infoln("issues found", matches)
	for val, ok := iter.Next(); ok; {
		var iv interface{} = val
		w.Processor(ctx, iv.(E)) // boy

		val, ok = iter.Next()
	}

	return nil
}
