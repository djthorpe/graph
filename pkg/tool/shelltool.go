package tool

import (
	"context"
	"errors"

	graph "github.com/djthorpe/graph"
	pkg "github.com/djthorpe/graph/pkg/graph"
	multierror "github.com/hashicorp/go-multierror"
)

func ShellTool(ctx context.Context, name string, args []string, objs ...interface{}) error {
	var result error

	// Create graph and state
	g, flagset := pkg.New(objs...), NewFlagset(name)
	if g == nil || flagset == nil {
		return errors.New("New() failed")
	}

	// Lifecycle: define->parse->new
	g.Define(flagset)
	if err := flagset.Parse(args); err != nil {
		return err
	}
	if err := g.New(flagset); err != nil {
		return err
	}

	// Lifecycle: run->dispose
	if err := g.Run(ctx, graph.RunAny); err != nil {
		result = multierror.Append(result, err)
	}
	if err := g.Dispose(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}
