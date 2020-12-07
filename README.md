# Graph: Dependency and Lifecycle Management

[Dependency injection](https://en.wikipedia.org/wiki/Dependency_injection) is a technique
to "magically" satisfy _dependencies_ within an application, reducing the mental overhead of the programmer and ensuring the dependencies can remain _loosely coupled_.

This module for the _golang_  provides one implementation of such magic, with the following aims:

  * Provide dependency injection for `struct` fields;
  * Explicit _lifecycle management_ for an application;
  * Mapping between an `interface` and its' concrete implementation through
    module imports;
  * Passing of state between dependencies in different phases of the lifecycle;
  * A framework for developing tools and unit tests.

The __Graph__ module provides a programming pattern which aims to target the
best features of _golang_ (channels, goroutines, composition) and enhance 
complex application development.

## Dependency Injection

A __Unit__ in __Graph__ is a singleton `struct` value which can be injected into dependencies. For example,

```go
package main

import "github.com/djthorpe/graph"

type A struct {
    graph.Unit
}

type B struct {
    graph.Unit
    *A
}

type C struct {
    graph.Unit
    *A
    *B
}

func main() {
    g := graph.New(&B{})    
    // ...
}

```

In this example, both `A` and `B` are defined as __Unit__ through including the anonymous field `graph.Unit`. By calling `graph.New` an instance of `A` is injected into the instance of `B`. _(Note: The impossibility of creating circular dependencies by design)_

If a graph was created by calling `graph.New(&C{})` instead, instances of `A` and `B` are injected into both `B` and `C`, however as a _Unit_ is a singleton pattern, only one `A` and one `B` are created first.

## Lifecycle Management

Unlike some other languages, Go does not proscribe any lifecycle management for instances, you can use `make` and `new` to create instances with zero-values.
Using __Graph__ lifecycle of instances can be managed through `Define`, `New`, `Run` and `Dispose` methods:

  * `graph.Define(graph.State)` calls instance methods to set any global 
    state within your application. The instance methods will be called in
    order of dependency;
  * `graph.New(graph.State) error` calls instance methods to initialise the
    application. The instance methods will be called in
    order of dependency;
  * `graph.Run(context.Context) error` calls instance methods to run the
    application. The order of calling is not guaranteed compared to
    instance dependencies;
  * `graph.Dispose() error` calls instance methods to dispose of any resources,
    in reverse dependency order.

To implement the lifecycle within a __Unit__ for example,

```go
func (a *A) Define(graph.State) {
    /* Define global state but do not initalize anything */ 
}

func (a *A) New(graph.State) error {
    /* Initialise instance from state, return any errors */
    return nil
}

func (a *A) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        /* Include other cases that need run here */
        }
    }
}

func (a *A) Dispose() error {
    /* Dispose of any resources used here */
    return nil
}

```

Each function definition is optional, not all __Unit__ instances will need
all the phases of the lifecycle.

## What is `graph.State`

The state instance defines any information which should be passed between
units. The interface definition is somewhat vague,
like `context.WithValue`:

```go
type State interface {
    Name() string 
    Value() interface{}
}
```

You can implement any other state information required within your own
implementation. There is an example in `github.com/djthorpe/graph/pkg/tool`
which defines state as command-line flags and arguments passed in.

## Implementing Application Lifecycle

You could implement the lifecycle within your own application like this. The `Graph` will ensure the lifecycle is   object 

```go
func Run(ctx context.Context,a *App,s graph.State) error {
    g := graph.New(a)

	g.Define(s)
	if err := g.New(s); err != nil {
	    return err
	}
	if err := g.Run(ctx); err != nil {
		return err
	}
	if err := g.Dispose(); err != nil {
		return err
	}

	return nil
}
```

## When does `Run` end


# Mapping An `interface` to a Unit

# Passing state between Unit instances

# Unit tests and Mocking

# Example: Shelltool lifecycle

# Example: Hello, world

# Example: Reader and Writer
