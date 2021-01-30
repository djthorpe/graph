package graph

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/djthorpe/graph"
	"github.com/hashicorp/go-multierror"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Graph struct {
	sync.RWMutex

	objs  []reflect.Value
	units map[reflect.Type]reflect.Value
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	errCircularReference = errors.New("Circular Reference")
	errUnassignableField = errors.New("Unassignable (private) Field")
)

/////////////////////////////////////////////////////////////////////
// NEW

// New returns a new graph object with "root" objects
// which are used to create the graph of dependencies. Returns
// nil if any object is not a graph.Unit. Run policy termination
// set to "Wait" (this is the default)
func New(objs ...interface{}) graph.Graph {
	if g := new(Graph).new(objs); g != nil {
		return g
	} else {
		return nil
	}
}

// Private new method which walks the dependencies and creates
// zero-valued fields
func (g *Graph) new(objs []interface{}) *Graph {
	g.objs = make([]reflect.Value, len(objs))
	g.units = make(map[reflect.Type]reflect.Value, len(objs)*4) // Arbitary assumption on number of units per object

	// Assign objects
	for i := range objs {
		v := reflect.ValueOf(objs[i])
		if isUnitType(v.Type()) == false {
			return nil
		} else if err := g.graph(v); err != nil {
			return nil
		} else {
			g.objs[i] = v
		}
	}

	return g
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Define passes state into each zero-valued unit and ensure the calls
// are done with leaf units first. Define is called on any unit only once.
// In general Define is used to set up state only, so there is no error
// return value.
func (g *Graph) Define(state graph.State) {
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	seen := make(map[reflect.Type]bool, len(g.units))
	for _, obj := range g.objs {
		g.do("Define", obj, []reflect.Value{reflect.ValueOf(state)}, seen, true)
	}
}

// New passes state into each unit and ensure the calls
// are done with leaf units first. New is called on any unit only once.
// error. In general state can be used to set up the unit, and co-ordinate
// between units. If any error is returned New immediately fails and returns.
func (g *Graph) New(state graph.State) error {
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	seen := make(map[reflect.Type]bool, len(g.units))
	for _, obj := range g.objs {
		if err := g.do("New", obj, []reflect.Value{reflect.ValueOf(state)}, seen, true); err != nil {
			return err
		}
	}

	return nil
}

// Dispose is called to release any resources. The calling order
// is for leaf units to be last. Errors are accumulated, so it is
// guaranteed that dispose is called on every unit
func (g *Graph) Dispose() error {
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	seen := make(map[reflect.Type]bool, len(g.units))
	for _, obj := range g.objs {
		if err := g.do("Dispose", obj, []reflect.Value{}, seen, true); err != nil {
			return err
		}
	}

	// Release graph resources
	g.objs = nil
	g.units = nil

	return nil
}

/////////////////////////////////////////////////////////////////////
// LOGGING

// Logger returns any registered logger unit or nil
// it not registered
func (g *Graph) Logger() graph.Logger {
	if t := graph.UnitTypeForInterface(logType); t == nil {
		return nil
	} else if v, exists := g.units[t]; exists == false {
		return nil
	} else {
		return v.Interface().(graph.Logger)
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (g *Graph) String() string {
	str := "<graph"
	if len(g.objs) > 0 {
		str += fmt.Sprint(" objs=", g.objs)
	}
	if len(g.units) > 0 {
		str += " units="
		for k := range g.units {
			str += fmt.Sprint(k, ",")
		}
		str = strings.TrimSuffix(str, ",")
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// build walks graph to create zero-values of units
func (g *Graph) graph(unit reflect.Value) error {
	return forEachField(unit, false, func(f reflect.StructField, i int) error {
		t := g.unitTypeForField(f)
		if t == nil {
			// Not a unit type, ignore
			return nil
		}

		// Check for X containing *X as field
		if equalsType(t, unit.Type()) {
			return errCircularReference
		}

		// Field must be public to be assignable
		if isPrivateField(f) {
			return fmt.Errorf("%w: %q of %q", errUnassignableField, f.Name, unit.Type())
		}

		// Create a zero-valued unit
		if _, exists := g.units[t]; exists == false {
			g.units[t] = reflect.New(t.Elem())
			g.graph(g.units[t])
		}

		// Set field to unit
		unit.Elem().Field(i).Set(g.units[t])

		// Return success
		return nil
	})
}

// do calls functions in the right order and ensures no unit is called twice
func (g *Graph) do(fn string, unit reflect.Value, args []reflect.Value, seen map[reflect.Type]bool, obj bool) error {
	var result error

	// Call Dispose leaf-last, continue on error
	switch fn {
	case "Dispose":
		if err := call(fn, unit, args); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Descend into struct
	if err := forEachField(unit, fn == "New", func(f reflect.StructField, i int) error {
		t := g.unitTypeForField(f)
		if t == nil {
			return nil
		} else if equalsType(t, unit.Type()) {
			return errCircularReference
		} else if _, exists := seen[t]; exists {
			return nil
		} else if err := g.do(fn, g.units[t], args, seen, false); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		result = multierror.Append(result, err)
	}

	// Call Define, New and Run leaf-first
	switch fn {
	case "Define", "New":
		if err := call(fn, unit, args); err != nil {
			result = multierror.Append(result, err)
		}
	case "Run":
		args[0].Interface().(*RunContext).Run(unit, obj)
	}

	// Mark this unit as 'seen'
	seen[unit.Type()] = true

	// Return any errors
	return result
}

// Returns type for struct field or nil if not a unit type.
// Will translate any mapped interfaces to concrete types.
func (g *Graph) unitTypeForField(f reflect.StructField) reflect.Type {
	t := f.Type
	if t.Kind() == reflect.Interface {
		t = graph.UnitTypeForInterface(f.Type)
	}
	if t == nil {
		return nil
	}
	if isUnitType(t) {
		return t
	} else {
		return nil
	}
}
