package event

import (
	"reflect"

	"github.com/djthorpe/graph"
)

func init() {
	if err := graph.RegisterUnit(reflect.TypeOf(&events{}), reflect.TypeOf((*graph.Events)(nil))); err != nil {
		panic("RegisterUnit(event.events): " + err.Error())
	}
}
