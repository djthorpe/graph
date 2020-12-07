package event

import (
	"github.com/djthorpe/graph"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type nullstate struct{}

/////////////////////////////////////////////////////////////////////
// METHODS

func NullState() graph.State {
	return &nullstate{}
}

func (*nullstate) Name() string {
	return "<nil>"
}

func (*nullstate) Value() interface{} {
	return nil
}
