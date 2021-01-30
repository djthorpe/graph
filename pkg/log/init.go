package log

import (
	"reflect"

	graph "github.com/djthorpe/graph"
)

func init() {
	graph.MustRegisterUnit(reflect.TypeOf(&Log{}), reflect.TypeOf((*graph.Logger)(nil)))
}
