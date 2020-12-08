
# Graph: Dependency and Lifecycle Management

[Dependency injection](https://en.wikipedia.org/wiki/Dependency_injection) is a technique
to "magically" satisfy _dependencies_ within an application, reducing the mental overhead for programmer and ensuring the dependencies remain _loosely coupled_.

This module for __Go__  provides one implementation of such magic, with the following aims:

  * Provide dependency injection for `struct` fields;
  * Explicit _lifecycle management_ for an application;
  * Mapping between an `interface` and its' concrete implementation through
    module imports;
  * Passing of state between dependencies in different phases of the lifecycle;
  * A framework for developing tools and unit tests.

The __Graph__ module provides a programming pattern which aims to target the
best features of __Go__ (channels, goroutines and composition for example)
to simplify complex application development.

## Installing

To use __Graph__ just import the definitions to create a __Unit__, an instance which can
be used in dependency injection. Mark any __Unit__ with an anonymous `graph.Unit` field.
For example, `mymodule.MyInterface` can be injected into an application shell tool:

```go
package main

import (
  graph "github.com/djthorpe/graph"
  tool "github.com/djthorpe/graph/pkg/tool"
  mymodule
)

type App struct {
  graph.Unit
  mymodule.MyInterface
}

func main() {
  tool.Shelltool(context.WithCancel(/* ... */),"myapp",os.Args[1:],new(App))
}
```

You can map an interface to an implementation and define the __Unit__ lifecycle:

```go
package mymodule

import (
  graph "github.com/djthorpe/graph"
  "reflect"
)

func init() {
  // Register mymodule.myUnit as implementation of exported mymodule.MyInterface
  graph.RegisterUnit(
      reflect.TypeOf(&myUnit{}), 
      reflect.TypeOf((*MyInterface)(nil))
  )
}

type MyInterface interface {
  // ... interface definition
}

type myUnit struct {
  graph.Unit
  graph.Events // Inject event pubsub dependency
  // ... other dependencies injected here
}

// New, Run and Dispose define the lifecycle
func (*myUnit) New(graph.State) error {
  // ...Initialize myUnit
}

func (*myUnit) Run(context.Context) error {
  // ...Run myUnit until cancel or deadline exceeded
}

func (*myUnit) Dispose() error {
  // ...Dispose of any resources used by myUnit
}
```

Then you can inject myUnit into your application, for example:

## Documentation

More information on usage of __Graph__ is provided in the following documentation:

  * [Guide](blob/main/doc/README.md)
  * [Examples of Graph usage](blob/main/doc/examples.md)
  * [pkg.go.dev](https://pkg.go.dev/github.com/djthorpe/graph)

## Project Status

This module is currently __in development__ but is mostly feature-complete.

## Community

  * [File an issue or question](http://github.com/djthorpe/graph/issues) on github.
  * Licensed under Apache 2.0, please read that license about using and forking __Graph__.
    Essentially, you are free to use in all circumstances as long as credit is provided.

