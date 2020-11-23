package goq_test

import (
	"context"
	"fmt"
	"goq"
)

type WorkerSample struct {
}

func NewWorkerSample() goq.Worker {
	return &WorkerSample{}
}

func (w *WorkerSample) Process(ctx context.Context, data interface{}) {
	fmt.Println(data)
}

func (w *WorkerSample) SetName(name string) {

}
