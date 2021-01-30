package main

import (
	"context"
	"fmt"
	"time"

	"github.com/djthorpe/graph"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Obj2 struct {
	graph.Unit
}

type Obj struct {
	graph.Unit
	graph.Logger
	key string
	*Obj2
	d time.Duration
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewObj(key string, d time.Duration) *Obj {
	this := new(Obj)
	this.key = key
	this.d = d
	return this
}

func (o *Obj) New(graph.State) error {
	graph.Requires(o, "Logger", "Obj2")
	return nil
}

func (o *Obj) Run(ctx context.Context) error {
	o.Printf("->Run %q\n", o.key)
	defer o.Printf("<-Run %q\n", o.key)

	select {
	case <-time.After(o.d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o *Obj) String() string {
	return fmt.Sprintf("<obj key=%q>", o.key)
}
