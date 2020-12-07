package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/djthorpe/graph/pkg/tool"
)

/////////////////////////////////////////////////////////////////////
// MAIN FUNC

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals - call cancel when interrupt received
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()

	// Run shell tool and print any errors
	if err := tool.ShellTool(ctx, "randomwriter", os.Args, new(App)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		os.Exit(0)
	}
}
