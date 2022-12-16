package wpool

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

const nJobs = 10
const nWorkers = 2

func TestWorkerPool(t *testing.T) {
	w := New(nWorkers)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// add jobs to the pool
	go func() {
		w.AddJobs(createJobs(nJobs))
		w.Finish()
	}()

	// execute on the pool
	go w.Run(ctx)

	// look at the results...
	for {
		select {
		case r, ok := <-w.Results():
			if !ok {
				continue
			}

			i, convertOk := r.Descriptor.Metadata["expected"].(int)
			if !convertOk {
				t.Fatalf("descriptor does not contain correct expected value: %v", r.Descriptor)
			}

			val := r.Value.(int)
			if val != i {
				t.Fatalf("wrong value: got = %d; want = %d", val, i*i)
			}
		case <-w.done:
			return
		default:

		}
	}
}

func TestWorkerPool_TimeOut(t *testing.T) {
	w := New(nWorkers)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Nanosecond*10)
	defer cancel()

	go w.Run(ctx)

	for {
		select {
		case r := <-w.Results():
			if r.Err != nil && !errors.Is(r.Err, context.DeadlineExceeded) {
				t.Fatalf("expected error: %v; got: %v", context.DeadlineExceeded, r.Err)
			}
		case <-w.done:
			return
		default:
		}
	}
}

func TestWorkerPool_Cancel(t *testing.T) {
	w := New(nWorkers)

	ctx, cancel := context.WithCancel(context.TODO())
	go w.Run(ctx)
	cancel()

	for {
		select {
		case r := <-w.Results():
			if r.Err != nil && !errors.Is(r.Err, context.Canceled) {
				t.Fatalf("expected error: %v; got: %v", context.Canceled, r.Err)
			}
		case <-w.done:
			return
		default:
		}
	}
}

func createJobs(n int) []Job {
	jobs := make([]Job, n)
	t := TaskFunc(intSquare)

	for i := 0; i < n; i++ {
		jobs[i] = Job{
			Descriptor: Descriptor{
				ID:   fmt.Sprintf("%02d", i),
				Type: "calc",
				Metadata: map[string]interface{}{
					"expected": i * i,
				},
			},
			Task: t,
			Args: i,
		}
	}
	return jobs
}
