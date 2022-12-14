package pool

import (
	"github.com/go-logr/logr"
	"io"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/stdr"
)

const (
	maxRoutines   = 25
	poolResources = 2
)

// idCounter provides a method for us giving each connection a unique id
var idCounter int32

type fakeConnection struct {
	ID     int32
	logger logr.Logger
}

func (f *fakeConnection) Close() error {
	f.logger.Info("Close", "fake connection", f.ID)
	return nil
}

// createConnection is a simple factory method we can use for testing
func createConnectionFactory(logger logr.Logger) FactoryFunc {
	return func() (io.Closer, error) {
		id := atomic.AddInt32(&idCounter, 1)
		logger.Info("createConnection", "new id", id)
		return &fakeConnection{ID: id, logger: logger}, nil
	}
}

func executeQuery(query int, p *Pool) error {
	// get conn from pool
	conn, err := p.Acquire()
	if err != nil {
		return err
	}
	defer p.Release(conn)

	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	p.logger.Info("finished query", "QID", query, "CID", conn.(*fakeConnection).ID)
	return nil
}

func TestPool(t *testing.T) {
	var wg sync.WaitGroup

	logger := stdr.New(nil)
	p, err := New(createConnectionFactory(logger), poolResources, WithLogger(logger))
	if err != nil {
		t.Errorf("unable to create pool with size = %d", poolResources)
		return
	}

	wg.Add(maxRoutines)
	for query := 0; query < maxRoutines; query++ {
		go func(q int) {
			err = executeQuery(q, p)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			wg.Done()
		}(query)
	}
	wg.Wait()

	logger.Info("closing test")
	p.Close()
}
