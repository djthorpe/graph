package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/djthorpe/graph/pkg/log"
	tool "github.com/djthorpe/graph/pkg/tool"
)

///////////////////////////////////////////////////////////////////////////////
// Main

func main() {
	// Set context for running
	ctx, cancel := context.WithCancel(context.Background())

	// Set up signals, send cancel on SIGINT or SIGTERM
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	// Start lifecycle, exit on error
	if err := tool.ShellTool(ctx, "t", nil, NewObj("0", 0), NewObj("A", time.Second), NewObj("B", 2*time.Second), NewObj("C", 3*time.Second)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
