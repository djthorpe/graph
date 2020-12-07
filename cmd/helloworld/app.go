package main

import (
	"context"
	"fmt"

	"github.com/djthorpe/graph"
	"github.com/djthorpe/graph/pkg/tool"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type App struct {
	graph.Unit

	// -name flag on command line
	name *string
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (this *App) Define(flags *tool.FlagSet) {
	this.name = flags.String("name", "", "Your name")
}

func (this *App) Run(context.Context) error {
	if *this.name != "" {
		fmt.Println("Hello," + *this.name)
	} else {
		fmt.Println("Hello, world")
	}
	return nil
}
