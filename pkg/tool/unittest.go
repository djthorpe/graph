package tool

import (
	"context"
	"sync"
	"testing"

	pkg "github.com/djthorpe/graph/pkg/graph"
	multierror "github.com/hashicorp/go-multierror"
)

type TestFunc func(interface{})

func Test(t *testing.T, args []string, obj interface{}, fn TestFunc) {
	var result error

	// Create graph and state
	g, flagset := pkg.New(obj), NewFlagset(t.Name())
	if g == nil || flagset == nil {
		t.Fatal("New() failed")
	}

	// Lifecycle: define->parse->new
	g.Define(flagset)
	if err := flagset.Parse(args[1:]); err != nil {
		t.Fatal(err)
	}
	if err := g.New(flagset); err != nil {
		t.Fatal(err)
	}

	// Create context with a cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call run and dispose in goroutine
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
	fn(obj)

	// Wait for run and dispose to end
	wg.Wait()

	// Check any errors
	if result != nil {
		t.Error(result)
	}
}
