package graph

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/djthorpe/graph"
	"github.com/hashicorp/go-multierror"
)

type Context struct {
	sync.WaitGroup

	cancels []context.CancelFunc
	errs    chan error
	stop    chan struct{}
	result  error
}

func NewContext(parent context.Context, runType graph.RunType) *Context {
	c := new(Context)
	c.stop = make(chan struct{})
	c.errs = make(chan error)

	// Collect errors returned by Run calls
	go func() {
		for err := range c.errs {
			if err != nil {
				c.result = multierror.Append(c.result, err)
				// TODO: Determine if cancel needed
				// if type=RunWait then no cancels
				// if type=RunAny then cancel if any root objs ended
				// if type=RunAll then cancel when all root objs ended
			}
		}
		fmt.Println("Ended A")
	}()

	// Goroutine to wait for completion of either parent
	// or objects according to run policy, then cancel all
	// Run goroutines and close channels
	go func() {
		// Wait for either parent context completed or other condition
		select {
		case <-parent.Done():
			fmt.Println("End reason = parent")
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
		fmt.Println("Ended B")
	}()

	return c
}

func (c *Context) Run(unit reflect.Value) {
	// Create a context which can be cancelled
	child, cancel := context.WithCancel(context.Background())
	c.cancels = append(c.cancels, cancel)
	c.WaitGroup.Add(1)
	go func() {
		defer c.WaitGroup.Done()
		c.errs <- call("Run", unit, []reflect.Value{reflect.ValueOf(child)})
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
