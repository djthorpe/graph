package tool

import (
	"context"
	"reflect"
	"sync"
	"testing"

	pkg "github.com/djthorpe/graph/pkg/graph"
	multierror "github.com/hashicorp/go-multierror"
)

type TestFunc func(interface{})

func Test(t *testing.T, args []string, obj, fn interface{}) {
	// Create graph and state
	g, flagset := pkg.New(obj), NewFlagset(t.Name())
	if g == nil || flagset == nil {
		t.Fatal("New() failed")
	}

	// Lifecycle: define->parse
	g.Define(flagset)
	if err := flagset.Parse(args); err != nil {
		t.Fatal(err)
	}

	// Set debug mode
	if logger := g.(*pkg.Graph).Logger(); logger != nil {
		logger.SetTest(t)
	}

	// Lifecycle: new
	if err := g.New(flagset); err != nil {
		t.Fatal(err)
	}

	// Create context with a cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call run and dispose in goroutine
	var result error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Lifecycle: run->dispose
		if err := g.Run(ctx); err != nil {
			result = multierror.Append(result, err)
		}
		if err := g.Dispose(); err != nil {
			result = multierror.Append(result, err)
		}
		wg.Done()
	}()

	// Call unit test
	if fn_ := reflect.ValueOf(fn); fn_.Kind() != reflect.Func {
		t.Fatal("Invalid test function")
	} else {
		fn_.Call([]reflect.Value{reflect.ValueOf(obj)})
	}

	// Wait for run and dispose to end
	wg.Wait()

	// Check any errors
	if result != nil {
		t.Error(result)
	}
}
