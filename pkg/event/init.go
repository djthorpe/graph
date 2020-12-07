package event

import (
	"reflect"

	"github.com/djthorpe/graph"
)

func init() {
	if err := graph.RegisterUnit(reflect.TypeOf(&publisher{}), reflect.TypeOf((*graph.Publisher)(nil))); err != nil {
		panic("RegisterUnit(event.publisher): " + err.Error())
	}
}
