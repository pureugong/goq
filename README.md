# GOQ

[![Go Report Card](https://goreportcard.com/badge/github.com/pureugong/goq)](https://goreportcard.com/report/github.com/pureugong/goq)

`goq` is a simple queue manager library

## Installation
If using go modules.
```sh
go get -u github.com/pureugong/goq
```

## Getting Started
```golang
// 1. init goq manager
manager := goq.NewManager(ctx, 1, nil)

// 2. init goq workers
manager.InitWorkers(10, func() goq.Worker {
    return NewWorkerSample()
})

// 3. enqueue tasks
for i := 0; i < 100; i++ {
    manager.Enqueue(i)
}

// 4. wait
manager.Wait()

```

## License

Released under the [MIT License](https://github.com/pureugong/goq/blob/master/LICENSE)
