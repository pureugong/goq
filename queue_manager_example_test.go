package goq_test

import (
	"context"
	"goq"
)

func ExampleQueueManager() {
	ctx := context.Background()

	// 1. init goq manager
	manager := goq.NewManager(ctx, 1, nil)

	// 2. init qoq workers
	manager.InitWorkers(10, func() goq.Worker {
		return NewWorkerSample()
	})

	// 3. enqueue tasks
	for i := 0; i < 100; i++ {
		manager.Enqueue(i)
	}

	// 4. wait
	manager.Wait()
}
