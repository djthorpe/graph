package graph

import (
	"github.com/djthorpe/graph"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type nullstate struct{}

/////////////////////////////////////////////////////////////////////
// METHODS

// NullState returns an empty value which can be used for Define()
// or emitted as an event
func NullState() graph.State {
	return &nullstate{}
}

func (*nullstate) Name() string {
	return "<nil>"
}

func (*nullstate) Value() interface{} {
	return nil
}

func (*nullstate) String() string {
	return "<nil>"
}
