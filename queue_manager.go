package goq

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager is queue manager's interface
type Manager interface {
	SetName(name string) Manager
	SetQueueChunkSize(chunkSize int) Manager
	SetSleep(duration time.Duration) Manager
	InitWorkers(numOfWorkers int, worker func() Worker)
	Enqueue(data interface{}) error
	Wait()
	Cancel(err error)
}

// ManagerImpl is implementation of Manager
type ManagerImpl struct {
	name   string
	ctx    context.Context
	cancel context.CancelFunc
	start  time.Time
	queue  chan interface{}
	// queue in chunk
	queueChunkSize int
	queueInChunk   []interface{}
	// queue progress log
	queueCount           int
	loggingPerQueueCount int
	// sleep
	hasSleep      bool
	sleepDuration time.Duration
	// wg is to wait for each worker
	wg     sync.WaitGroup
	logger *logrus.Entry
}

// NewManager is a constructor of Manager
func NewManager(ctx context.Context, queueBufferSize int, logger *logrus.Entry) Manager {
	ctx, cancel := context.WithCancel(ctx)
	qm := &ManagerImpl{
		ctx:    ctx,
		start:  time.Now(),
		cancel: cancel,
		queue:  make(chan interface{}, queueBufferSize),
		logger: logger,
		// queue in chunk
		queueChunkSize: 1,
		queueInChunk:   make([]interface{}, 0),
		// queue progress log
		queueCount:           0,
		loggingPerQueueCount: 10000,
	}
	return qm
}

// SetName is to set queue manager's name
func (m *ManagerImpl) SetName(name string) Manager {
	m.name = name
	m.logger = m.logger.WithField("name", name)
	return m
}

// SetQueueChunkSize is to queue tasks in chunk
func (m *ManagerImpl) SetQueueChunkSize(chunkSize int) Manager {
	m.queueChunkSize = chunkSize
	m.logger = m.logger.WithField("queue_chunk_size", chunkSize)
	return m
}

// SetSleep is to make worker to sleep after each process
func (m *ManagerImpl) SetSleep(duration time.Duration) Manager {
	m.hasSleep = true
	m.sleepDuration = duration
	return m
}

// InitWorkers is to init worker to watch task queue
func (m *ManagerImpl) InitWorkers(numOfWorkers int, worker func() Worker) {
	for i := 0; i < numOfWorkers; i++ {
		m.wg.Add(1)
		go func(id int) {
			var w *WorkerWrapper
			if !m.hasSleep {
				// noromal processor case
				w = newWorker(id, &m.wg, worker())
			} else {
				// sleeping processor case
				w = newSleepingWorker(id, &m.wg, worker(), m.sleepDuration)
			}
			m.logger.Infof("init %s", w.name)
			w.process(m.ctx, m.queue)
			return
		}(i)
	}
}

// Enqueue is to queue task
func (m *ManagerImpl) Enqueue(data interface{}) error {
	if err := m.ctx.Err(); err != nil {
		m.logger.Errorf("enqueue failed: %v", err)
		return err
	}

	// count queue number
	m.queueCount = m.queueCount + 1
	if m.queueCount%m.loggingPerQueueCount == 0 {
		m.logger.WithFields(logrus.Fields{
			"count": m.queueCount,
		}).Info("enqueued")
	}

	// enqueue
	if m.queueChunkSize == 1 {
		m.queue <- data
		return nil
	}

	// enqueue in chunk
	m.queueInChunk = append(m.queueInChunk, data)
	if len(m.queueInChunk) == m.queueChunkSize {
		m.queue <- m.queueInChunk
		m.queueInChunk = nil
		return nil
	}

	return nil
}

// Wait is to block until the end of queue process
func (m *ManagerImpl) Wait() {
	// empty queue in chunk if not empty
	if m.queueChunkSize > 1 && len(m.queueInChunk) > 0 {
		m.queue <- m.queueInChunk
		m.queueInChunk = nil
	}

	// close queue channel
	close(m.queue)
	m.logger.Debug("closed queue")

	// wait for workers
	m.wg.Wait()
	m.logger.Debug("all workers done")

	// update ctx
	m.ctx.Done()

	// log
	m.logger.WithFields(logrus.Fields{
		"elapsed": time.Since(m.start).String(),
		"count":   m.queueCount,
	}).Info("complete all the queued tasks")
}

// Cancel is to cancel all workers
func (m *ManagerImpl) Cancel(err error) {
	m.logger.Errorf("cancelling queue due to: %v", err.Error())
	// cancel
	m.cancel()
}
