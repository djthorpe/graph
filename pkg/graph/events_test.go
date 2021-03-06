package graph_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	graph "github.com/djthorpe/graph"
	pkg "github.com/djthorpe/graph/pkg/graph"
)

/////////////////////////////////////////////////////////////////////
// UNITS

type E struct {
	graph.Unit
	graph.Events
}

type W struct {
	graph.Unit
	graph.Events
}

type R struct {
	graph.Unit
	graph.Events
}

func (this *E) Run(ctx context.Context) error {
	ch := this.Events.Subscribe()
	defer this.Events.Unsubscribe(ch)

	n := 100
	go func() {
		for i := 0; i < n; i++ {
			this.Events.Emit(nil)
		}
	}()

	i := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
			i++
			if i == n {
				fmt.Println("got all", n, "events")
				return nil
			}
		}
	}
}

func (this *R) Run(ctx context.Context) error {
	ch := this.Events.Subscribe()
	defer this.Events.Unsubscribe(ch)

	i := 0
	n := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
			i++
			if i == n {
				fmt.Println("got all", n, "events")
				return nil
			}
		}
	}
}

func (this *W) Run(ctx context.Context) error {
	n := 100
	for i := 0 i < n; i++ {
		this.Events.Emit(nil)
	}
}

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Events_001(t *testing.T) {
	e := new(E)
	if g := pkg.New(e); g == nil {
		t.Error("Expected non-nil return")
	}
	if e.Events == nil {
		t.Error("Expected non-nil Events object")
	}
}

func Test_Events_002(t *testing.T) {
	g, s := pkg.New(&E{}), NewState(t)
	if g == nil {
		t.Error("Expected non-nil return")
	}

	// Define -> New
	g.Define(s)
	if err := g.New(s); err != nil {
		t.Error(err)
	}

	// Run -> Dispose
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := g.Run(ctx); err != nil {
		t.Error(err)
	} else if err := g.Dispose(); err != nil {
		t.Error(err)
	}
}
