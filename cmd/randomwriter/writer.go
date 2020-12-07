package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/djthorpe/graph"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Writer struct {
	graph.Unit
	graph.Events
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Run emits a null event at least once every three seconds but
// randomly before this
func (w *Writer) Run(ctx context.Context) error {
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			w.Events.Emit(nil)
			timer.Reset(time.Millisecond * time.Duration(rand.Int31()%3000))
		}
	}
}
