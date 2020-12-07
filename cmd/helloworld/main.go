package main

import (
	"context"
	"fmt"
	"os"

	"github.com/djthorpe/graph/pkg/tool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tool.ShellTool(ctx, "helloworld", os.Args, new(App)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		os.Exit(0)
	}
}
