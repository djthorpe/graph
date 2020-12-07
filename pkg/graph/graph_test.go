package graph_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	graph "github.com/djthorpe/graph"
	pkg "github.com/djthorpe/graph/pkg/graph"
)

/////////////////////////////////////////////////////////////////////
// MOCK STATE OBJECT

type state struct {
	*testing.T

	value string
}

func NewState(t *testing.T) *state {
	s := new(state)
	s.T = t
	return s
}
func (s *state) Name() string {
	return s.Name()
}
func (s *state) Value() interface{} {
	return s.value
}
func (s *state) Add(value string) {
	s.value += value
}
func (s *state) Equals(value string) bool {
	return s.value == value
}

/////////////////////////////////////////////////////////////////////
// UNITS

type A struct {
	graph.Unit
	*state
}

type B struct {
	graph.Unit
	*A
}

type C struct {
	graph.Unit
	*B
}

type D struct {
	graph.Unit
	*B
}

func (*A) Define(s *state) {
	s.Log("Called Define on A")
	s.Add("A")
}

func (this *A) New(s *state) error {
	s.Log("Called New on A")
	s.Add("A")
	this.state = s
	return nil
}

func (this *A) Dispose() error {
	this.state.Log("Called Dispose on A")
	this.state.Add("A")
	return nil
}

func (this *A) Run(context.Context) error {
	this.state.Log("Called Run on A (immediately returns)")
	this.state.Add("A")
	return nil
}

func (*B) Define(s *state) {
	s.Log("Called Define on B")
	s.Add("B")
}

func (*B) New(s *state) error {
	s.Log("Called New on B")
	s.Add("B")
	return nil
}

func (this *B) Run(ctx context.Context) error {
	this.state.Log("Called Run on B (waits for done)")
	<-ctx.Done()
	fmt.Println("B is done")
	this.state.Log("B Run done")
	this.state.Add("B")
	return nil
}

func (this *B) Dispose() error {
	this.state.Log("Called Dispose on B")
	this.state.Add("B")
	return nil
}

func (this *C) New(s *state) error {
	s.Log("Called New on C")
	s.Add("C")
	return nil
}

func (this *C) Dispose() error {
	this.state.Log("Called Dispose on C")
	this.state.Add("C")
	return nil
}

func (this *C) Run(context.Context) error {
	this.state.Log("Called Run on C (returns an error)")
	this.state.Add("C")
	return errors.New("Error from C")
}

func (this *D) New(s *state) error {
	s.Log("Called New on D")
	s.Add("D")
	return nil
}

func (this *D) Dispose() error {
	this.state.Log("Called Dispose on D")
	this.state.Add("D")
	return nil
}

func (this *D) Run(ctx context.Context) error {
	this.state.Log("Called Run on D (returns after one second)")
	this.state.Add("D")
	select {
	case <-ctx.Done():
		fmt.Println("ctxdone")
	case <-time.After(time.Second):
		fmt.Println("ticker done")
	}
	return ctx.Err()
}

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Graph_001(t *testing.T) {
	type TestNotUnit struct{}
	type TestUnit struct{ graph.Unit }
	type TestNamedUnit struct{ named graph.Unit }

	if g := pkg.New(pkg.RunWait); g == nil {
		t.Error("Expected non-nil return")
	}
	if g := pkg.New(pkg.RunWait, &TestNotUnit{}); g != nil {
		t.Error("Expected nil return")
	}
	if g := pkg.New(pkg.RunWait, &TestUnit{}); g == nil {
		t.Error("Expected non-nil return")
	}
	if g := pkg.New(pkg.RunWait, &TestNamedUnit{}); g != nil {
		t.Error("Expected nil return")
	}
}

func Test_Graph_002(t *testing.T) {
	type A struct{ graph.Unit }
	type B struct {
		graph.Unit
		*A
	}

	if g := pkg.New(pkg.RunWait, new(A)); g == nil {
		t.Error("Expected non-nil return")
	}
	b := new(B)
	if g := pkg.New(pkg.RunWait, b); g == nil {
		t.Error("Expected non-nil return")
	} else if b.A == nil {
		t.Error("Expected non-nil A")
	}
}

func Test_Graph_003(t *testing.T) {
	type A struct{ graph.Unit }
	type B struct {
		graph.Unit
		*A
	}
	type C struct {
		graph.Unit
		a *A
		b *B
	}

	if g := pkg.New(pkg.RunWait, new(A)); g == nil {
		t.Error("Expected non-nil return")
	}
	b := new(B)
	if g := pkg.New(pkg.RunWait, b); g == nil {
		t.Error("Expected non-nil return")
	} else if b.A == nil {
		t.Error("Expected non-nil A")
	}
}

func Test_Graph_004(t *testing.T) {
	g := pkg.New(pkg.RunWait, new(B))
	if g == nil {
		t.Error("Expected non-nil return")
	}
	state := NewState(t)
	g.Define(state)
	if state.Equals("AB") == false {
		t.Error("Unexpected define call order:", state.Value())
	}
}

func Test_Graph_005(t *testing.T) {
	g := pkg.New(pkg.RunWait, new(B))
	if g == nil {
		t.Error("Expected non-nil return")
	}
	state := NewState(t)
	g.Define(state)
	if err := g.New(state); err != nil {
		t.Error(err)
	}
	if state.Equals("ABAB") == false {
		t.Error("Unexpected define/new call order:", state.Value())
	}
}

func Test_Graph_006(t *testing.T) {
	g := pkg.New(pkg.RunWait, new(B))
	if g == nil {
		t.Error("Expected non-nil return")
	}
	state := NewState(t)
	g.Define(state)
	if err := g.New(state); err != nil {
		t.Error(err)
	}
	if err := g.Dispose(); err != nil {
		t.Error(err)
	}
	if state.Equals("ABABBA") == false {
		t.Error("Unexpected define/new/dispose call order:", state.Value())
	}
}

func Test_Graph_007(t *testing.T) {
	// A <- B <- C
	g := pkg.New(pkg.RunWait, new(B), new(C))
	state := NewState(t)

	if err := g.New(state); err != nil {
		t.Error(err)
	}
	if state.Equals("ABC") == false {
		t.Error("Unexpected New call order:", state.Value(), "...expected:", "ABC")
	}
}

func Test_Graph_008(t *testing.T) {
	// A <- B <- C for first two objects and then A
	g := pkg.New(pkg.RunWait, new(B), new(C), new(A))
	state := NewState(t)

	if err := g.New(state); err != nil {
		t.Error(err)
	}
	if state.Equals("ABCA") == false {
		t.Error("Unexpected New call order:", state.Value(), "...expected:", "ABCA")
	}
}

func Test_Graph_009(t *testing.T) {
	// A <- B
	g := pkg.New(pkg.RunWait, new(B))
	state := NewState(t)

	if err := g.New(state); err != nil {
		t.Error(err)
	}
	if state.Equals("AB") == false {
		t.Error("Unexpected New call order:", state.Value(), "...expected:", "AB")
	}

	// Cancel after one second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start running, returns only on deadline exceeded
	if err := g.Run(ctx); err != nil && err != context.DeadlineExceeded {
		t.Error(err)
	}

	// B waits to end so A should end running first, order should be AB
	if state.Equals("ABAB") == false {
		t.Error("Unexpected New call order:", state.Value(), "...expected:", "ABAB")
	}
}

func Test_Graph_010(t *testing.T) {
	// A
	g := pkg.New(pkg.RunAll, new(A))
	state := NewState(t)

	if err := g.New(state); err != nil {
		t.Error(err)
	}

	// Wait for 2 secs (should end anyway immediately)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start running, returns immediately as A ends immediately
	now := time.Now()
	if err := g.Run(ctx); err != nil && err != context.DeadlineExceeded {
		t.Error(err)
	}
	if time.Since(now) >= time.Second {
		t.Error("Run did not return immediately")
	}
}

func Test_Graph_011(t *testing.T) {
	// A
	g := pkg.New(pkg.RunAny, new(A))
	state := NewState(t)

	if err := g.New(state); err != nil {
		t.Error(err)
	}

	// Wait for 2 secs (should end anyway immediately)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start running, returns immediately as A ends immediately
	now := time.Now()
	if err := g.Run(ctx); err != nil && err != context.DeadlineExceeded {
		t.Error(err)
	}
	if time.Since(now) >= time.Second {
		t.Error("Run did not return immediately")
	}
}

func Test_Graph_012(t *testing.T) {
	// A <- B <- D
	g := pkg.New(pkg.RunAny, new(D), new(D))
	state := NewState(t)

	if err := g.New(state); err != nil {
		t.Error(err)
	}

	// Start running, returns after either D ends
	if err := g.Run(context.Background()); err != nil && err != context.DeadlineExceeded {
		t.Error(err)
	}
}
