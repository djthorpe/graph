
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

## Dependency Injection

A __Unit__ is a singleton `struct` instance which can be injected into dependencies.
Then, the `graph.New` function is called to inject the dependencies. For example,

```go
package main

import (
    graph "github.com/djthorpe/graph"
    pkg "github.com/djthorpe/graph/pkg/graph"
)

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
    g := pkg.New(&B{})    
    // ...
}

```

In this example, both `A` and `B` are defined as __Unit__ through including the anonymous field `graph.Unit`. By calling `pkg.New` an instance of `A` is injected into the instance of `B` _(Note the impossibility of creating circular dependencies by design)_.

If a graph was created by calling `pkg.New(&C{})` instead, instances of `A` and `B` are injected into both `B` and `C`. However in this example, as a _Unit_ is
a singleton pattern, only one `A` and one `B` instance are created, and the
`A` instance is shared with both `B` and `C`

## Lifecycle Management

Unlike other languages, __Go__ does not proscribe lifecycle management for instances other than using `new` and `make` to create zero-valued instances.

Using __Graph__, the lifecycle of instances can be managed through `Define`, `New`, `Run` and `Dispose` functions:

  * `graph.Define(graph.State)` calls instance methods to set any global 
    state within your application. The instance methods will be called in
    order of dependency;
  * `graph.New(graph.State) error` calls instance methods to initialise the
    application. The instance methods will be called in
    order of dependency;
  * `graph.Run(context.Context) error` calls instance methods to run the
    application. The order of calling is not guaranteed compared to
    instance dependencies. Context is passed which indicates when the
    function should terminate and return;
  * `graph.Dispose() error` calls instance methods to dispose of any resources,
    in reverse dependency order.

For example,

```go
func (*A) Define(graph.State) {
    /* Define global state but do not initalize anything */ 
}

func (*A) New(graph.State) error {
    /* Initialise instance from state, return any errors */
    return nil
}

func (*A) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        /* Include other cases that need run here */
        }
    }
}

func (*A) Dispose() error {
    /* Dispose of any resources used here */
    return nil
}

```

Each function definition is optional, not all units will need all the phases of 
the lifecycle.

### What is `graph.State`?

A `graph.State` implementation can define any information which should be passed 
between units. The interface definition is purposefully vague and left to the
implementor. The interface definition is:

```go
type State interface {
    Name() string 
    Value() interface{}
}
```

There are two cases where `graph.State` is used:

  * To pass state around instances in `Define` and `New` phases of the lifecycle;
  * To pass state between instances in the `Run` phase, using a `graph.Events`
    object which provides `Subscribe`, `Unsubscribe` and `Emit` functions.

There is an example in `github.com/djthorpe/graph/pkg/tool`
which defines state as command-line flags and arguments using the `flag` module.
There is more information on passing state as events below.

### Implementing the application lifecycle

You implement the lifecycle within your own application calling the appropriate
methods on the `pkg.Graph` instance. For example,

```go
func RunApp(ctx context.Context, a,b *App, s graph.State) error {
    g := pkg.New(a,b)

	g.Define(s)
	if err := g.New(s); err != nil {
	    return err
	}
	if err := g.Run(ctx); err != nil { // Run
		return err
	}
	if err := g.Dispose(); err != nil {
		return err
	}

	return nil
}
```

### When does `Run` return?

Each __Unit__ invokes the `Run` method independently as a goroutine and 
could return under one of the following conditions:

  * It could return at any time with or without an error condition;
  * It could wait for the passed context to indicate completion.

There are three sensible strategies for when the `pkg.Run` function should
return. If we define the value passed into `pkg.New` as the top-level 
(aka __object__) instance,

  1. When any object instance returns: In the example
     above, the commented __Run__ call will return when either
     `a` or `b` complete or when the parent context indicates completion;
  2. When all object instances return: In the example
     above, the commented __Run__ call will return when both `a` and `b` 
     complete or when the parent context indicates completion;
  3. When the parent context indicates completion: If 
     either `a` or `b` complete beforehand, the commented function will 
     continue to block until the parent context indicates completion.

Typically, the former two policies would be used when be used for developing a 
command-line tool and the latter policy when running a unit test.

## Mapping an interface to a Unit (and integration testing)

Concrete implementation is decoupled in __Graph__ by using interface fields
rather than type fields. Different __Unit__ implementations can then be 
injected based in a lookup. Swapping one implementation for another in this
way aides integration testing with mock instances, for example.

A concrete implementation is mapped to an interface before calling `graph.New`.
Typically an `init.go` in your module implemention is used to map. Alas, you need to
use slightly clunky syntax to do this. If your concrete implementation of the
__Unit__ is `mymodule.myunit` and it satisfies the interface definition of `graph.Events` 
then,

```go
package mymodule

import (
	"reflect"
	"github.com/djthorpe/graph"
)

func init() {
    if err := graph.RegisterUnit(
        reflect.TypeOf(&myunit{}), 
        reflect.TypeOf((*graph.Events)(nil))
    ); err != nil {
		panic("RegisterUnit(graph.Events): " + err.Error())
	}
}
```

Then, inject the dependency as follows:

```go

import (
    _ "github.com/myuser/mymodule"
)

type App struct {
    graph.Unit
    graph.Events
}
```

A call to `graph.New` will then inject a `mymodule.myunit` dependency into 
your instance. Substituting, for example, a mock implementation is then
acheieved through import a different module in your tests.

## Passing state between Unit instances

Instance `Run` functions are loosely coupled. To pass state between instances,
a `graph.Events` dependency can be injected. This is defined by the following
interface:

```go
type Events interface {
	Emit(State)
	Subscribe() <-chan State
	Unsubscribe(<-chan State)
}
```

The `Emit` method is used to pass state to any other subscribed instance. An
instance would typically subscribe to these events like this:

```go
func (app *App) Run(ctx context.Context) error {
	ch := app.Events.Subscribe()
	defer app.Events.Unsubscribe(ch)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt := <-ch:
			app.Process(evt)
		}
	}
}

func (app *App) Process(evt graph.State) {
    // Do something with state...
}
```

It is possible to `Emit` within the `Process` function without causing
deadlock, but care needs to be taken. Emitting a `nil` value will
translate into a `graph.NilEvent` object being emitted.

## Implementing unit tests

You can test any __Unit__ by using the `tool.Test` function. For example,

```go
package mymodule_test

import (
  "testing"
  "mymodule"
  tool "github.com/djthorpe/graph/pkg/tool"
)

func Test_001(t *testing.T) {
  tool.Test(t,nil,new(mymodule.MyUnit),func(unit *mymodule.MyUnit) {
    t.Log("Perform testing here:",unit)
  })
}
```

The `tool.Test` method accepts an array of command-line arguments or nil
if these are not used.

## Other approaches for dependency injection

  * [Dingo](https://pkg.go.dev/flamingo.me/dingo) also maps implementations
    to interfaces. Injecting dependencies is done programmatically using
    the `Bind` function.
  * [Wire](https://pkg.go.dev/github.com/google/wire) is a code generation
    tool which satisfies dependencies before build.
  * [Dig](https://pkg.go.dev/go.uber.org/dig) injects dependencies using 
    reflection, separating providers from result instances.

## Contributions & Usage

Please see the [examples](examples.md) documentation for descriptions on
how __Graph__ can be used practicallly.

__Graph__ is licensed under Apache 2.0. Please read that license about using and forking __Graph__. The main conditions require preservation of copyright and license notices. Contributors provide an express grant of patent rights. Licensed works, modifications, and larger works may be distributed under different terms and without source code.


Questions, comments and pull requests are welcomed.
