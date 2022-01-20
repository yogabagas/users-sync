package taskworker

import (
	"context"
	"sync"
)

type SingleWork func(ctx context.Context, input interface{}) (interface{}, error)

type SingleTaskWorker struct {
	work   SingleWork
	wg     *sync.WaitGroup
	jobs   chan interface{}
	holder *holder
}

func NewSingleTaskWorker(ctx context.Context, maxConcurrency uint8, work SingleWork, total int) *SingleTaskWorker {
	wg := sync.WaitGroup{}
	wg.Add(total)
	jobs := make(chan interface{})
	t := &SingleTaskWorker{
		work:   work,
		wg:     &wg,
		jobs:   jobs,
		holder: &holder{},
	}
	for i := 0; i < int(maxConcurrency); i++ {
		go t.worker(ctx)
	}
	return t
}
func New(ctx context.Context, wg *sync.WaitGroup, maxConcurrency uint8, work SingleWork, total int) *SingleTaskWorker {
	wg.Add(total)
	jobs := make(chan interface{})
	t := &SingleTaskWorker{
		work:   work,
		wg:     wg,
		jobs:   jobs,
		holder: &holder{},
	}
	for i := 0; i < int(maxConcurrency); i++ {
		go t.worker(ctx)
	}
	return t
}
func (t *SingleTaskWorker) Do(input interface{}) {
	t.jobs <- input
}
func (t *SingleTaskWorker) worker(ctx context.Context) {
	for job := range t.jobs {
		result, err := t.work(ctx, job)
		t.holder.Store(Result{Result: result, Err: err})
		t.wg.Done()
	}
}
func (t *SingleTaskWorker) Results() []Result {
	t.wg.Wait()
	close(t.jobs)
	return t.holder.res
}
