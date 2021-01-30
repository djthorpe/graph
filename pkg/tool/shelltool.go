package tool

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"

	pkg "github.com/djthorpe/graph/pkg/graph"
	multierror "github.com/hashicorp/go-multierror"
)

func ShellTool(ctx context.Context, name string, args []string, objs ...interface{}) error {
	var result error

	// Check parameters
	if ctx == nil || len(objs) == 0 {
		return context.Canceled
	}
	if args == nil {
		args = os.Args
	}
	if name == "" {
		name = filepath.Base(args[0])
	}

	// Create graph and state
	g, flagset := pkg.New(objs...), NewFlagset(name)
	if g == nil || flagset == nil {
		return errors.New("New() failed")
	}

	// Add debugging flag
	debug := flagset.Bool("debug", false, "Verbose logging")

	// Lifecycle: define->parse
	g.Define(flagset)
	if err := flagset.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		} else {
			return err
		}
	}

	// Set debug mode if -debug flag
	if logger := g.(*pkg.Graph).Logger(); logger != nil && *debug {
		logger.SetTest(nil)
	}

	// Lifecycle: new
	if err := g.New(flagset); err != nil {
		if err == flag.ErrHelp {
			flagset.Usage()
			return nil
		} else {
			return err
		}
	}

	// Lifecycle: run->dispose
	if err := g.Run(ctx); err == flag.ErrHelp {
		flagset.Usage()
		return nil
	} else if err != nil {
		result = multierror.Append(result, err)
	}
	if err := g.Dispose(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}
