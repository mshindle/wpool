package runner

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

const smallTimeout = 20 * time.Millisecond
const mediumTimeout = 3 * time.Second

func proc(id int) {
	log.Printf("Processor - Task #%d", id)
	time.Sleep(time.Duration(id) * time.Second)
}

func TestNew(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 2*smallTimeout)
	defer cancel()

	r := New(smallTimeout)
	select {
	case <-r.timeout:
		log.Printf("received timeout")
	case <-ctx.Done():
		t.Errorf("timeout did not fire")
	}
}

func TestRunner_Add(t *testing.T) {
	type args struct {
		d     time.Duration
		tasks []func(int)
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "two_tasks",
			args: args{
				d:     smallTimeout,
				tasks: []func(int){proc, proc},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.args.d)
			r.Add(tt.args.tasks...)
			if got := len(r.tasks); got != tt.want {
				t.Errorf("Add() error -  got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestRunner_Start(t *testing.T) {
	tests := []struct {
		name  string
		d     time.Duration
		tasks []func(int)
		err   error
	}{
		{
			name:  "no_errors",
			d:     mediumTimeout,
			tasks: []func(int){proc, proc},
			err:   nil,
		},
		{
			name:  "timeout",
			d:     smallTimeout,
			tasks: []func(int){proc, proc, proc},
			err:   ErrTimeout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.d)
			r.Add(tt.tasks...)
			if err := r.Start(); err != tt.err {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.err)
			}
		})
	}
}

func TestRunner_gotInterrupt(t *testing.T) {
	var wg sync.WaitGroup
	var err error

	r := New(mediumTimeout)
	r.Add(proc, proc)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = r.Start()
	}()

	// send the interrupt signal
	r.interrupt <- os.Interrupt
	wg.Wait()

	if err != ErrInterrupt {
		t.Errorf("expected ErrInterrupt, got %v", err)
	}
}
