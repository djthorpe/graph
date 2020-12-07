package graph

import (
	"context"
)

/////////////////////////////////////////////////////////////////////
// INTERFACES

// Graph encapulates the lifecycle of objects and units
type Graph interface {
	Define(State)
	New(State) error
	Run(context.Context) error
	Dispose() error
}

type State interface {
	Name() string       // Arbitary name
	Value() interface{} // Arbitary value
}

// Events is used to pass state between units
type Events interface {
	// Emit state
	Emit(State)

	// Subscribe to receive all events
	Subscribe() <-chan State

	// Unsubscribe from receiving any events
	Unsubscribe(<-chan State)
}

/////////////////////////////////////////////////////////////////////
// UNITS

// Unit marks a singleton instance. You should include a unit
// as an anonymous field in your structure, for example:
//
// type MyUnit struct {
//    graph.Unit
//    /* ...other fields... */
// }
// Which marks your type so that dependencies can be injected
// when the graph is created.
type Unit struct{}

// No-op default functions for lifecycle
func (this *Unit) Define(State)              { /* NOOP */ }
func (this *Unit) New(State) error           { /* NOOP */ return nil }
func (this *Unit) Run(context.Context) error { /* NOOP */ return nil }
func (this *Unit) Dispose() error            { /* NOOP */ return nil }
