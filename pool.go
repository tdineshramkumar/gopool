package gopool

import (
	"sync"
	"time"
)

var (
	// RequestTimeout is the time Execute waits to for a existing goroutine in
	// the pool to become free to execute the task, else it launches a new goroutine
	// into the thread pool
	RequestTimeout time.Duration = 1 * time.Millisecond
	// IdleTimeout is the time a goroutine in the pool waits in the pool for task to execute
	// before exiting
	IdleTimeout time.Duration = 1 * time.Second
)

type Task func()

type Pool interface {
	Execute(Task)
	Wait()
}

type pool struct {
	taskQ chan Task
	wg    sync.WaitGroup
}

func New() Pool {
	return &pool{
		taskQ: make(chan Task),
	}
}

func (p *pool) worker() {
	p.wg.Add(1)
	defer p.wg.Done()
	for {
		select {
		case taskfn, ok := <-p.taskQ:
			if !ok {
				// If taskQ is closed then
				return
			}
			// Execute the task
			taskfn()
		case <-time.After(IdleTimeout):
			// If worker idle for timeout
			// then exit
			return
		}
	}
}

func (p *pool) Execute(fn Task) {
	select {
	case p.taskQ <- fn:
		// A go-routine from the pool will execute the task
		return
	case <-time.After(RequestTimeout):
		// If all go-routine are busy
		// Launch a new worker into the pool
		go p.worker()
		// Now wait till write completes
		p.taskQ <- fn
	}
}

func (p *pool) Wait() {
	// Close the task queue to indicate no more tasks
	close(p.taskQ)
	// Wait for all workers to exit
	p.wg.Wait()
}
