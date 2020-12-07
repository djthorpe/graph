package graph

import (
	"reflect"

	"github.com/djthorpe/graph"
)

func init() {
	if err := graph.RegisterUnit(reflect.TypeOf(&events{}), reflect.TypeOf((*graph.Events)(nil))); err != nil {
		panic("RegisterUnit(graph.events): " + err.Error())
	}
}
