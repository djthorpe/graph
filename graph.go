package graph

import (
	"context"
	"errors"
	"reflect"
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

var (
	iface = make(map[reflect.Type]reflect.Type)
)

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

// Events is used to pass state between units
type Events interface {
	// Emit state
	Emit(State)

	// Subscribe to receive events
	Subscribe() <-chan State

	// Unsubscribe from receiving events
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

/////////////////////////////////////////////////////////////////////
// REGISTER INTERFACES

func RegisterUnit(t, i reflect.Type) error {
	if t == nil || i == nil {
		return errors.New("Nil Parameter")
	}
	for i.Kind() == reflect.Ptr {
		i = i.Elem()
	}
	if i.Kind() != reflect.Interface {
		return errors.New("Not an interface")
	}
	if t.Implements(i) == false {
		return errors.New("Does not implement interface")
	}
	if _, exists := iface[i]; exists {
		return errors.New("Duplicate call to RegisterUnit")
	}

	iface[i] = t
	return nil
}

func UnitTypeForInterface(i reflect.Type) reflect.Type {
	if t, exists := iface[i]; exists {
		return t
	} else {
		return nil
	}
}

// NewGraph creates a new lifecycle for objects and dependent
// units and will return the lifecycle structure (Graph).
//func NewGraph(objs ...interface{}) Graph {
// TODO return pkg.New(objs...)
//	return nil
//}
