package graph

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type RunContext struct {
	sync.Mutex

	parent         context.Context
	done, finished chan struct{}
	cancels        []context.CancelFunc
	all, objs      sync.WaitGroup
	result         *Error
}

type Error struct {
	sync.Mutex
	err *multierror.Error
}

///////////////////////////////////////////////////////////////////////////////
// RUN

// Run is called to initiate goroutines for each unit and waits until
// all "obj" run functions end. The order of
// running unit run functions is not guaranteed. Any errors from
// Run returns are collected and returned.
func (g *Graph) Run(ctx context.Context) error {
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	// Create context which allows units to run
	child := NewContext(ctx)

	// Call run functions for objects and units
	seen := make(map[reflect.Type]bool, len(g.units))
	for _, obj := range g.objs {
		g.do("Run", obj, []reflect.Value{reflect.ValueOf(child)}, seen, true)
	}

	// Wait for end of run condition
	<-child.Done()

	// Return collected errors
	return child.Err()
}

///////////////////////////////////////////////////////////////////////////////
// CONTEXT

func NewContext(parent context.Context) context.Context {
	c := new(RunContext)
	c.parent = parent
	c.done, c.finished = make(chan struct{}), make(chan struct{})
	c.result = new(Error)

	// Wait for either parent to signal done, or all objects
	// to have completed their Run methods
	go func() {
		select {
		case <-c.parent.Done():
			c.result.Append(c.parent.Err())
		case <-c.finished:
			// Finished comes about when all root obj have finished
		}

		// Send cancels to Run methods
		c.Mutex.Lock()
		for _, cancel := range c.cancels {
			cancel()
		}
		c.Mutex.Unlock()

		// Wait for all Run methods to end
		c.all.Wait()

		// Signal done
		c.done <- struct{}{}
	}()

	// Wait for all object Run methods to complete and then
	// send finish signal
	go func() {
		c.objs.Wait()
		c.finished <- struct{}{}
	}()

	// Return context
	return c
}

func (c *RunContext) Done() <-chan struct{} {
	return c.done
}

func (c *RunContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

func (c *RunContext) Run(unit reflect.Value, obj bool) {
	// Create a context which can be cancelled
	child, cancel := context.WithCancel(context.Background())

	// Append cancels, this occurs sequentially so no need to guard
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.cancels = append(c.cancels, cancel)

	// In goroutine, call Run and pass back the result
	if obj {
		c.objs.Add(1)
	}
	c.all.Add(1)
	go func() {
		defer c.all.Done()
		if obj {
			defer c.objs.Done()
		}
		if err := call("Run", unit, []reflect.Value{reflect.ValueOf(child)}); err != nil {
			if errors.Is(err, context.Canceled) == false {
				c.result.Append(err)
			}
		}
	}()
}

func (c *RunContext) Err() error {
	return c.result.Unwrap()
}

func (c *RunContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

///////////////////////////////////////////////////////////////////////////////
// ERROR

func (r *Error) Append(err error) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	if err != nil {
		r.err = multierror.Append(r.err, err)
	}
	return r
}

func (r *Error) Unwrap() error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	return r.err.Unwrap()
}

func (r *Error) Error() string {
	return r.Unwrap().Error()
}
