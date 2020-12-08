
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

To use __Graph__ just import the definitions and either the package which creates graphs
or the tools which provide lifecycle management:

```go
package main

import (
  graph "github.com/djthorpe/graph"
  pkg "github.com/djthorpe/graph/pkg/graph"
  tool "github.com/djthorpe/graph/pkg/tool"
)

func init() {
  // Register myUnit as implementation of exported MyInterface
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
```

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

