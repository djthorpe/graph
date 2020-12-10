package tool

import (
	"context"
	"errors"
	"flag"

	pkg "github.com/djthorpe/graph/pkg/graph"
	multierror "github.com/hashicorp/go-multierror"
)

func ShellTool(ctx context.Context, name string, args []string, objs ...interface{}) error {
	var result error

	// Create graph and state
	g, flagset := pkg.NewAny(objs...), NewFlagset(name)
	if g == nil || flagset == nil {
		return errors.New("New() failed")
	}

	// Lifecycle: define->parse->new
	g.Define(flagset)
	if err := flagset.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		} else {
			return err
		}
	}
	if err := g.New(flagset); err != nil {
		if err == flag.ErrHelp {
			flagset.Usage()
			return nil
		} else {
			return err
		}
	}

	// Lifecycle: run->dispose
	if err := g.Run(ctx); err != nil {
		result = multierror.Append(result, err)
	}
	if err := g.Dispose(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}
