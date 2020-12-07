package graph

import (
	"context"
	"errors"
)

/////////////////////////////////////////////////////////////////////
// CONSTANTS

// RunType determines how the graph.Run method completes. It will
// either complete based on parent context, or when all "objects"
// have completed, or when any "object" completes.
type RunType int

const (
	RunWait RunType = iota // Terminates when parent context is done
	RunAny                 // Terminates when any obj Run goroutines end
	RunAll                 // Terminates when all obj Run goroutines end
)

var (
	ErrParentDeadlineExceeded = context.DeadlineExceeded
	ErrParentCanceled         = context.Canceled
	ErrAnyObjectEnded         = errors.New("Any object Run() ended")
	ErrAllObjectsEnded        = errors.New("All object Run() ended")
)

/////////////////////////////////////////////////////////////////////
// CREATE A GRAPH

// NewGraph creates a new lifecycle for objects and dependent
// units and will return the lifecycle structure (Graph).
//func NewGraph(objs ...interface{}) Graph {
// TODO return pkg.New(objs...)
//	return nil
//}

/////////////////////////////////////////////////////////////////////
// INTERFACES

// Graph encapulates the lifecycle of objects and units
type Graph interface {
	Define(State)
	New(State) error
	Run(context.Context, RunType) error
	Dispose() error
}

type State interface {
	Name() string       // Name
	Value() interface{} // Arbitary value
}

// Publisher is used to pass state between units
type Publisher interface {
	// Emit state
	Emit(State)

	// Subscribe to receive state
	Subscribe() <-chan State

	// Unsubscribe from receiving state
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
