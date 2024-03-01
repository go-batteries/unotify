package consumers

import (
	"context"
	"sync"
	"time"
	"unotify/app/resque"

	"github.com/sirupsen/logrus"
)

type GithubEventConsumer struct {
	rsqcl resque.Queuer

	EventChannel chan string
	DoneCh       chan bool
}

const DefaultGithubWorkerCount int = 4

func NewGithubEventConsumer(rsqcl resque.Queuer) *GithubEventConsumer {
	consumer := &GithubEventConsumer{
		rsqcl:        rsqcl,
		EventChannel: make(chan string, 1000),
		DoneCh:       make(chan bool),
	}

	return consumer
}

// BLPOP is blocking, So, use multiple workers
// Each worker runs a BLPOP and sends the event on EventChannel
// The Dispatcher receives the event
// Waits for an Worker to register a channel
// Sends the data to that channel
// The can be multiple of those Workers

// The main idea is not block the consuming channel with processing
// Once way to do this would be using a central channel
// which accepts a channel, that way

// var EventChannel chan events.GithubEvents
func (gevcon *GithubEventConsumer) Start(ctx context.Context, key string) {
	var wg sync.WaitGroup

	for i := 0; i < DefaultGithubWorkerCount; i++ {
		wg.Add(1)

		go func(key string, done chan bool) {
			// debounce, every 5 seconds read data
			ticker := time.NewTicker(10 * time.Second)
			defer wg.Done()

			for {
				logrus.WithContext(ctx).Debugln("starting to read from github event list")
				select {
				case <-done:
					return
				case <-ctx.Done():
					return
				case <-ticker.C:
					results, err := gevcon.rsqcl.ReadMsg(ctx, key)
					if err != nil {
						// logrus.WithContext(ctx).WithError(err).Debug("failed to read")
					}

					for _, result := range results {
						logrus.WithContext(ctx).Debugln("event ", result)
						gevcon.EventChannel <- result
					}
				}
			}
		}(key, gevcon.DoneCh)
	}

	wg.Wait()
	close(gevcon.DoneCh)
}

func (gevcon *GithubEventConsumer) Stop() {
	gevcon.DoneCh <- true
}

func GithubDispatcher(ctx context.Context, EventChannel chan string) {
	// pool := &WorkerPool{
	// Pool: make(chan chan resque.Payload, DefaultGithubWorkerCount)
	// }
	pool := &WorkerPool{
		Pool: make(chan chan string, DefaultGithubWorkerCount*2),
	}

	for i := 0; i < DefaultGithubWorkerCount*2; i++ {
		worker := &Worker{Bench: make(chan string), Done: make(chan bool)}
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
