// Package runner manages the running and lifetime of a process
package runner

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"
)

type Runner struct {
	// interrupt reports on a signal from the OS
	interrupt chan os.Signal

	// complete reports that processing has finished
	complete chan error

	// timeout duration to allow task to run
	timeout time.Duration

	// tasks holds a set of functions that are executed in index order
	tasks []func(int)
}

var (
	ErrTimeout   = errors.New("received timeout")
	ErrInterrupt = errors.New("received interrupt")
)

func New(d time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   d,
	}
}

func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

func (r *Runner) Start(ctx context.Context) error {
	signal.Notify(r.interrupt, os.Interrupt)
	to, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	go func() {
		r.complete <- r.run()
	}()

	select {
	case err := <-r.complete:
		return err
	case <-to.Done():
		return ErrTimeout
	}
}

func (r *Runner) run() error {
	for id, task := range r.tasks {
		select {
		case <-r.interrupt:
			signal.Stop(r.interrupt)
			return ErrInterrupt
		default:
			task(id)
		}
	}
	return nil
}
