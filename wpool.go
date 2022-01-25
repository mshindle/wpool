package wpool

import (
	"context"
	"fmt"
	"sync"
)

type WorkerPool struct {
	nWorkers int
	jobs     chan Job
	results  chan Result
	Done     chan bool
}

func New(n int) WorkerPool {
	return WorkerPool{
		nWorkers: n,
		jobs:     make(chan Job, n),
		results:  make(chan Result, n),
		Done:     make(chan bool),
	}
}

// Run begins executing the pool of incoming jobs with a known context.
// If the pool is terminated using the Context.Done() channel, the error returned
// by ctx.Done() is wrapped in Result.Err
func (w WorkerPool) Run(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < w.nWorkers; i++ {
		wg.Add(1)
		// fan out worker goroutines, read from jobs channel, and push into results
		go worker(ctx, &wg, w.jobs, w.results)
	}
	wg.Wait()
	close(w.Done)
	close(w.results)
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- Result) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			// fan-in job execution multiplexing results into the results channel
			results <- job.execute(ctx)
		case <-ctx.Done():
			results <- Result{
				Err: fmt.Errorf("cancelled worker: %w", ctx.Err()),
			}
			return
		}
	}
}

// Results provides access to the result channel
func (w WorkerPool) Results() <-chan Result {
	return w.results
}

// AddJob adds a single job to the pool. Call will block until Job can be
// added to the channel.
func (w WorkerPool) AddJob(j Job) {
	w.jobs <- j
}

// AddJobs adds a slice of jobs to the pool. Call will block until each Job
// can be added to the channel.
func (w WorkerPool) AddJobs(j []Job) {
	for i, _ := range j {
		w.jobs <- j[i]
	}
}

// Finish closes the job queue terminating the pool.
func (w WorkerPool) Finish() {
	close(w.jobs)
}