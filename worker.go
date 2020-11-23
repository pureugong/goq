package goq

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Worker is the interface for processing queue tasks,
// it's required to implement business logic how to process queue tasks
type Worker interface {
	Process(ctx context.Context, data interface{})
	SetName(name string)
}

// WorkerWrapper is a wrapper of worker to be managed by an queue manager
type WorkerWrapper struct {
	name   string
	wg     *sync.WaitGroup
	worker Worker
	// sleep
	hasSleep      bool
	sleepDuration time.Duration
}

// newWorker is to init worker
func newWorker(id int, wg *sync.WaitGroup, worker Worker) *WorkerWrapper {
	name := fmt.Sprintf("worker # %d", id)
	worker.SetName(name)
	return &WorkerWrapper{
		name:     name,
		wg:       wg,
		worker:   worker,
		hasSleep: false,
	}
}

// NewWorker is to init worker
func newSleepingWorker(id int, wg *sync.WaitGroup, worker Worker, sleepDuration time.Duration) *WorkerWrapper {
	name := fmt.Sprintf("sleeping (%s) worker # %d", sleepDuration.String(), id)
	worker.SetName(name)
	return &WorkerWrapper{
		name:          name,
		wg:            wg,
		worker:        worker,
		hasSleep:      true,
		sleepDuration: sleepDuration,
	}
}

func (w *WorkerWrapper) process(ctx context.Context, datas <-chan interface{}) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// cancel, or complete by ctx
			return

		case data, ok := <-datas:
			if ok {
				w.worker.Process(ctx, data)
				if w.hasSleep {
					time.Sleep(w.sleepDuration)
				}
			} else {
				// channel closed
				return
			}

		default:
			// waiting for queue
		}
	}
}
