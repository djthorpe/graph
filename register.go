package graph

import (
	"errors"
	"reflect"
)

/////////////////////////////////////////////////////////////////////
// REGISTER INTERFACES

var (
	iface = make(map[reflect.Type]reflect.Type)
)

// RegisterUnit is called to register a unit as being
// mapped to an interface type, so that you can mock
// units of a particular interface by anonymously including
// their packages. This function should be called in
// the init() method of that package and returns an
// error if there are invalid arguments or two unit
// types map to a single interface.
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

// UnitTypeForInterface returns a concrete type for an
// interface or nil if not found.
func UnitTypeForInterface(i reflect.Type) reflect.Type {
	if t, exists := iface[i]; exists {
		return t
	} else {
		return nil
	}
}
