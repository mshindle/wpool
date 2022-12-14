// Package pool manages a user defined set of resources
package pool

import (
	"errors"
	"io"
	"sync"

	"github.com/go-logr/logr"
)

type FactoryFunc func() (io.Closer, error)

type Pool struct {
	m         sync.Mutex
	resources chan io.Closer
	factory   FactoryFunc
	closed    bool
	created   int
	logger    logr.Logger
}

type Option func(*Pool)

var (
	// ErrPoolClosed is used when Acquire is called on a closed pool.
	ErrPoolClosed  = errors.New("pool has been closed")
	ErrInvalidSize = errors.New("pool size must be 1 or greater")
	ErrMaxCreated  = errors.New("maximum number of resources have been created")
)

func New(fn FactoryFunc, size uint, opts ...Option) (*Pool, error) {
	if size == 0 {
		return nil, ErrInvalidSize
	}

	p := &Pool{
		factory:   fn,
		resources: make(chan io.Closer, size),
		logger:    logr.Discard(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

// Acquire grabs a resource from the pool.
func (p *Pool) Acquire() (io.Closer, error) {

	if p.closed {
		return nil, ErrPoolClosed
	}

	if r, err := p.createNew(); err != ErrMaxCreated {
		return r, err
	}

	r, ok := <-p.resources
	p.logger.Info("Acquire: shared resource")
	if !ok {
		return nil, ErrPoolClosed
	}
	return r, nil
}

func (p *Pool) createNew() (io.Closer, error) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.created >= cap(p.resources) {
		return nil, ErrMaxCreated
	}

	p.logger.Info("Acquire: new resource", "created", p.created, "cap()", cap(p.resources))
	c, err := p.factory()
	if err == nil {
		p.created = p.created + 1
	}
	return c, err
}

func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	// close the channel before we drain the channel of all resources.
	// failure to do this causes a deadlock
	close(p.resources)
	for r := range p.resources {
		_ = r.Close()
	}
}

// Release places a new resource onto the pool
func (p *Pool) Release(r io.Closer) {
	p.m.Lock()
	defer p.m.Unlock()

	// if the pool is closed, discard the resource
	if p.closed {
		_ = r.Close()
		return
	}

	select {
	case p.resources <- r:
		// attempt to place resource into queue
		p.logger.Info("Release: in queue")

	default:
		// if queue is at capacity, close it
		p.logger.Info("Release: closing")
		_ = r.Close()
	}
}

// WithLogger assigns a logger to pool
func WithLogger(logger logr.Logger) Option {
	return func(p *Pool) {
		p.logger = logger
	}
}
