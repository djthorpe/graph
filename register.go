package graph

import (
	"errors"
	"fmt"
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
		return errors.New("RegisterUnit: Nil Parameter")
	}
	for i.Kind() == reflect.Ptr {
		i = i.Elem()
	}
	if i.Kind() != reflect.Interface {
		return errors.New("RegisterUnit: Not an interface: " + fmt.Sprint(i))
	}
	if t.Implements(i) == false {
		return errors.New("RegisterUnit: Does not implement interface: " + fmt.Sprint(i))
	}
	if _, exists := iface[i]; exists {
		return errors.New("RegisterUnit: Duplicate call to RegisterUnit")
	}
	iface[i] = t
	return nil
}

// MustRegisterUnit calls RegisterUnit and panics if any errors occur
func MustRegisterUnit(t, i reflect.Type) {
	if err := RegisterUnit(t, i); err != nil {
		panic(fmt.Sprint(t, ": ", err))
	}
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

// Requires will panic if a defined field is nil
func Requires(u Graph, fields ...string) {
	v := reflect.ValueOf(u).Elem()
	for _, field := range fields {
		if f := v.FieldByName(field); f.IsValid() == false {
			panic(fmt.Sprintf("%v: Invalid dependency %q", v.Type(), field))
		} else if f.Kind() != reflect.Interface && f.Kind() != reflect.Ptr {
			panic(fmt.Sprintf("%v: Invalid field %q", v.Type(), field))
		} else {
			t := f.Type()
			if f.Kind() == reflect.Interface {
				t = UnitTypeForInterface(t)
			}
			if t == nil {
				panic(fmt.Sprintf("%v: Missing import for interface %q", v.Type(), f.Type()))
			} else if f.IsZero() {
				panic(fmt.Sprintf("%v: Missing dependency %q", v.Type(), t))
			}
		}
	}
}
