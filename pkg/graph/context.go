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

type runContext struct {
	sync.Mutex
	sync.WaitGroup

	policy  runPolicy
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
type runPolicy int

const (
	// runWait is a policy which terminates when parent context is done
	runWait runPolicy = iota
	// runAny policy terminates when ANY obj Run goroutines end
	runAny
	// runAll policy terminates when ALL obj Run goroutines end
	runAll
)

/////////////////////////////////////////////////////////////////////
// NEW

func newContext(parent context.Context, policy runPolicy) *runContext {
	c := new(runContext)
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
		fmt.Println("GOT END REASON")
		// Send cancels to children
		for _, cancel := range c.cancels {
			cancel()
		}
		// Wait for children to have terminated, close channels
		c.WaitGroup.Wait()
		fmt.Println("FINISHED RUN")
		close(c.errs)
		close(c.done)
	}()

	return c
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *runContext) Run(unit reflect.Value, obj bool) {
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
		c.errs <- err
	}()
}

/////////////////////////////////////////////////////////////////////
// context.Context INTERFACE IMPLEMENTATION

func (c *runContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (c *runContext) Done() <-chan struct{} {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// Collect errors returned by Run calls
	go func() {
		var wg sync.WaitGroup
		for err := range c.errs {
			// Decrement counter if an object ended
			c.objs.Dec(err.Obj())

			// Evaluate terminate policy (in separate goroutine)
			wg.Add(1)
			go func(err *result) {
				defer wg.Done()
				switch c.policy {
				case runAny:
					if err.Obj() {
						c.reason <- struct{}{}
					}
				case runAll:
					if err.Obj() && c.objs.IsZero() {
						c.reason <- struct{}{}
					}
				}
			}(err)

			// Append any error
			if err.IsErr() {
				c.result = multierror.Append(c.result, err.Unwrap())
			}
		}
		// Wait until policy evaluation has completed before
		// closing channels
		wg.Wait()
		close(c.reason)
	}()

	return c.done
}

func (c *runContext) Err() error {
	// Return no-error
	if c.result == nil {
		return nil
	}
	// Unwrap error
	if err, ok := c.result.(*multierror.Error); ok {
		if len(err.Errors) == 0 {
			return nil
		} else if len(err.Errors) == 1 {
			return err.Errors[0]
		}
	}
	// Return standard error
	return c.result
}
