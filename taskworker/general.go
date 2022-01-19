package taskworker

import (
	"context"
	"sync"
)

type TaskWorker struct {
	workers        []*worker
	maxConcurrency uint8
}

func NewTaskWorker(maxConcurrency uint8) *TaskWorker {
	return &TaskWorker{
		workers:        make([]*worker, 0),
		maxConcurrency: maxConcurrency,
	}
}

func (t *TaskWorker) Register(work Work) {
	t.workers = append(t.workers, &worker{
		work: work,
	})
}
func (t *TaskWorker) Run(ctx context.Context) []Result {
	wg := sync.WaitGroup{}
	total := len(t.workers)
	temp := holder{
		res: make([]Result, 0),
	}
	wg.Add(total)
	i := 0
	for i < total {
		if temp.GetActiveWorker() < t.maxConcurrency {
			temp.Add()
			go func(index int) {
				res, err := t.workers[index].work(ctx)
				temp.Store(Result{Result: res, Err: err})
				wg.Done()
			}(i)
			i++
		}
	}
	wg.Wait()
	return temp.GetAllResult()
}

type worker struct {
	work Work
}

type Work func(ctx context.Context) (interface{}, error)
