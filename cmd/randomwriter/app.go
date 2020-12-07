package main

import (
	"context"
	"fmt"

	"github.com/djthorpe/graph"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type App struct {
	graph.Unit
	*Reader
	*Writer
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (this *App) Run(ctx context.Context) error {
	// Application waits until CTRL+C is pressed
	fmt.Println("Waiting for CTRL+C to end")
	<-ctx.Done()

	return nil
}
