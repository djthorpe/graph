
# Graph: Dependency and Lifecycle Management

This page outlines some example usage of __Graph__. If you use the module
in your own code and wish to provide it as an example, please edit this
page and send a pull request.

## Example: ShellTool

>[Code: github.com/djthorpe/graph/pkg/tool](https://github.com/djthorpe/graph/tree/main/pkg/tool)

An implementation of __Graph__ so that your application and lifecycle
can easily be encapsulated in a command-line tool. Your `main` func
could look like this:

```go
package main

import (
    // ... other imports
	tool "github.com/djthorpe/graph/pkg/tool"
)

type App struct {
    graph.Unit
    // ... other dependencies
}

func (this *App) Define(flags *tool.FlagSet) {
    // ... define command-line flags
}

func (this *App) Run(context.Context) error {
    // ... run your application
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()

	if err := tool.ShellTool(ctx, "helloworld", os.Args, new(App)); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
```

## Example: Hello, World

>[Code: github.com/djthorpe/graph/cmd/helloworld](https://github.com/djthorpe/graph/tree/main/cmd/helloworld)

The canonical hello world application. Defines a `--name` flag on the command line and prints it out if set.

## Example: RandomWriter

>[Code: github.com/djthorpe/graph/cmd/randomwriter](https://github.com/djthorpe/graph/tree/main/cmd/randomwriter)

A `Writer` unit emits an event at least once every three seconds, and a `Reader` unit consumes the event and prints out its' contents. The implementation of the `Run` methods are as follows:

`writer.go`:

```go
func (w *Writer) Run(ctx context.Context) error {
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			w.Events.Emit(nil)
			timer.Reset(time.Millisecond * time.Duration(rand.Int31()%3000))
		}
	}
}
```

`reader.go`:
```go

func (r *Reader) Run(ctx context.Context) error {
	ch := r.Events.Subscribe()
	defer r.Events.Unsubscribe(ch)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt := <-ch:
			fmt.Println("Event: ", evt)
		}
	}
}
```

The command-line tool quits when CTRL+C or interrupt signal is caught.
