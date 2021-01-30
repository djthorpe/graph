package graph

import (
	"context"
	"testing"
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

// Logger provides a simple interface for logging to stderr
type Logger interface {
	Print(...interface{})          // Print output logging to stderr
	Debug(...interface{})          // Print output logging to stderr when debugging
	Printf(string, ...interface{}) // Print formatted logging to stderr
	Debugf(string, ...interface{}) // Print formatted logging to stderr when debugging

	IsDebug() bool      // IsDebug returns true if debug flag is set
	Test() *testing.T   // Test returns testing context when in a unit test
	SetTest(*testing.T) // SetTest will set debug to true and if provided the test context
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
func (*Unit) Define(State)              { /* NOOP */ }
func (*Unit) New(State) error           { /* NOOP */ return nil }
func (*Unit) Run(context.Context) error { /* NOOP */ return nil }
func (*Unit) Dispose() error            { /* NOOP */ return nil }
