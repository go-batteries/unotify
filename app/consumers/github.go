package consumers

import (
	"context"
	"sync"
	"time"
	"unotify/app/resque"

	"github.com/sirupsen/logrus"
)

// Consumes Events from Redis Q
// returning an EventChannel, to which
// it pases the data
// Worker Pool reads from this channel
// and assigns work

type GithubEventConsumer struct {
	rsqcl resque.Queuer

	EventChannel chan string
	// DoneCh       chan chan bool
}

const DefaultGithubEventChanSize = 100

func NewGithubEventConsumer(rsqcl resque.Queuer) *GithubEventConsumer {
	consumer := &GithubEventConsumer{
		rsqcl:        rsqcl,
		EventChannel: make(chan string, DefaultGithubEventChanSize),
		// DoneCh:       make(chan chan bool),
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

		go func(key string) {
			// debounce, every 10 seconds read data
			ticker := time.NewTicker(10 * time.Second)
			defer wg.Done()
			defer ticker.Stop()

			for {
				logrus.WithContext(ctx).Debugln("starting to read from github event list")

				select {
				case <-ctx.Done():
					logrus.WithContext(ctx).Infoln("closing github event consumer")
					return

				case <-ticker.C:
					results, err := gevcon.rsqcl.ReadMsg(ctx, key)
					if err != nil {
						logrus.WithContext(ctx).WithError(err).Debugf("%+v", err)
					}

					for _, result := range results {
						logrus.WithContext(ctx).Debugln("event ", result)
						gevcon.EventChannel <- result
					}
				}
			}
		}(key)
	}

	wg.Wait()
}

// TODO: Check if data Handle fix worked
func (gevcon *GithubEventConsumer) Stop() {
	// gevcon.DoneCh <- true
	// close(gevcon.DoneCh)
}
