package graph

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/djthorpe/graph"
	"github.com/hashicorp/go-multierror"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Context struct {
	sync.WaitGroup

	cancels []context.CancelFunc
	errs    chan *Error
	objs    *counter
	stop    chan struct{}
	result  error
}

/////////////////////////////////////////////////////////////////////
// NEW

func NewContext(parent context.Context, runType graph.RunType) *Context {
	c := new(Context)
	c.stop = make(chan struct{})
	c.errs = make(chan *Error)
	c.objs = NewCounter()
	policy := make(chan graph.RunType)

	// Collect errors returned by Run calls
	go func() {
		for err := range c.errs {
			// Decrement counter if an object ended
			c.objs.Dec(err.Obj())

			// Check Run terminate policy
			switch runType {
			case graph.RunAll:
				if err.Obj() && c.objs.IsZero() {
					policy <- runType
				}
			case graph.RunAny:
				if err.Obj() {
					policy <- runType
				}
			}

			// Append any error
			if err.IsErr() {
				c.result = multierror.Append(c.result, err.Unwrap())
			}
		}
	}()

	// Goroutine to wait for completion of either parent
	// or objects according to run policy, then cancel all
	// Run goroutines and close channels
	go func() {
		// Wait for either parent context completed or other condition
		select {
		case <-parent.Done():
			// End reason is because parent cancelled
			break
		case <-policy:
			// End reason is due to runType policy
			break
		}
		// Send cancels to children
		for _, cancel := range c.cancels {
			cancel()
		}
		// Wait for children to have terminated, close channels
		c.WaitGroup.Wait()
		close(c.stop)
		close(c.errs)
		close(policy)
	}()

	return c
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Context) Run(unit reflect.Value, obj bool) {
	// Create a context which can be cancelled
	child, cancel := context.WithCancel(context.Background())
	c.cancels = append(c.cancels, cancel)
	c.WaitGroup.Add(1)
	c.objs.Inc(obj)
	go func() {
		defer c.WaitGroup.Done()
		c.errs <- NewError(call("Run", unit, []reflect.Value{reflect.ValueOf(child)}), obj)
	}()
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (c *Context) Done() <-chan struct{} {
	return c.stop
}

func (c *Context) Err() error {
	return c.result
}
