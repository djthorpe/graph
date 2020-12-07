package graph

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Context struct {
	sync.Mutex
	sync.WaitGroup

	policy  RunPolicy
	cancels []context.CancelFunc // Holds cancel functions for any running goroutine
	errs    chan *result         // Channel for return values after Run completes
	objs    *counter             // Reference counter for (root) object completion
	reason  chan struct{}
	done    chan struct{} // Channel which is closed when context completes
	result  error         // Holds the errors to be returned
}

/////////////////////////////////////////////////////////////////////
// CONSTANTS

// RunPolicy determines how the graph.Run method completes. It will
// either complete based on parent context, or when all "objects"
// have completed, or when any "object" completes.
type RunPolicy int

const (
	// RunWait is a policy which terminates when parent context is done
	RunWait RunPolicy = iota
	// RunAny policy terminates when ANY obj Run goroutines end
	RunAny
	// RunAll policy terminates when ALL obj Run goroutines end
	RunAll
)

/////////////////////////////////////////////////////////////////////
// NEW

func NewContext(parent context.Context, policy RunPolicy) *Context {
	c := new(Context)
	c.done, c.reason = make(chan struct{}), make(chan struct{})
	c.errs = make(chan *result)
	c.objs = NewCounter()
	c.policy = policy

	// Goroutine to wait for completion of either parent
	// or objects according to run policy, then cancel all
	// Run goroutines and close channels
	go func() {
		// Wait for either parent context completed or other condition
		select {
		case <-parent.Done():
			// End reason is because parent cancelled
			break
		case <-c.reason:
			// End reason is due to policy
			break
		}
		// Send cancels to children
		for _, cancel := range c.cancels {
			cancel()
		}
		// Wait for children to have terminated, close channels
		fmt.Println("->WAIT")
		c.WaitGroup.Wait()
		fmt.Println("<-WAIT")
		close(c.errs)
		close(c.done)
	}()

	return c
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Context) Run(unit reflect.Value, obj bool) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// Create a context which can be cancelled
	child, cancel := context.WithCancel(context.Background())
	c.cancels = append(c.cancels, cancel)

	// Increment counters which are used to wait for the
	// completion state
	c.WaitGroup.Add(1)
	c.objs.Inc(obj)

	// In goroutine, call Run and pass back the result
	go func() {
		defer c.WaitGroup.Done()
		err := newResult(call("Run", unit, []reflect.Value{reflect.ValueOf(child)}), obj)
		fmt.Println("-> ERR")
		c.errs <- err
		fmt.Println("<- ERR")
	}()
}

/////////////////////////////////////////////////////////////////////
// context.Context INTERFACE IMPLEMENTATION

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (c *Context) Done() <-chan struct{} {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// Collect errors returned by Run calls
	go func() {
		for err := range c.errs {
			fmt.Println("GOT ERR", err)

			// Decrement counter if an object ended
			c.objs.Dec(err.Obj())

			// Check Run terminate policy
			switch c.policy {
			case RunAny:
				if err.Obj() {
					c.reason <- struct{}{}
				}
			case RunAll:
				if err.Obj() && c.objs.IsZero() {
					c.reason <- struct{}{}
				}
			}

			// Append any error
			if err.IsErr() {
				c.result = multierror.Append(c.result, err.Unwrap())
			}
		}
		fmt.Println("END OF ERRS")
		close(c.reason)
	}()

	return c.done
}

func (c *Context) Err() error {
	return c.result
}
