package tool

import (
	"context"

	graph "github.com/djthorpe/graph"
	pkg "github.com/djthorpe/graph/pkg/graph"
	multierror "github.com/hashicorp/go-multierror"
)

func ShellTool(ctx context.Context, name string, args []string, objs ...interface{}) error {
	var results error

	// Create graph and state
	g, flagset := pkg.New(objs...), NewFlagset(name)

	// Lifecycle: define->parse flags->new
	g.Define(flagset)
	if err := flagset.Parse(args); err != nil {
		return err
	}
	if err := g.New(flagset); err != nil {
		return err
	}

	// Lifecycle: run->dispose
	if err := g.Run(ctx, graph.RunAny); err != nil {
		results = multierror.Append(results, err)
	}
	if err := g.Dispose(); err != nil {
		results = multierror.Append(results, err)
	}

	return results
}
