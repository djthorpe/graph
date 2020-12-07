package main

import (
	"context"
	"fmt"

	"github.com/djthorpe/graph"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Reader struct {
	graph.Unit
	graph.Events
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (r *Reader) Run(ctx context.Context) error {
	ch := r.Events.Subscribe()
	defer r.Events.Unsubscribe(ch)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt := <-ch:
			r.Process(evt)
		}
	}
}

func (r *Reader) Process(evt graph.State) {
	fmt.Println("Event: ", evt)
}
