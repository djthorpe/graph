package graph_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	graph "github.com/djthorpe/graph"
	pkg "github.com/djthorpe/graph/pkg/graph"
)

/////////////////////////////////////////////////////////////////////
// MOCK STATE OBJECT

type state struct {
	*testing.T
	sync.RWMutex

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
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()
	return s.value
}
func (s *state) Add(value string) {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	s.value += value
}
func (s *state) Equals(value string) bool {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()
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
	this.state.Add("a")
	this.state.Log("A Run done")
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
	this.state.Log("B Run done")
	this.state.Add("b")
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
	this.state.Add("c")
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
	this.state.Add("d")
	select {
	case <-ctx.Done():
		this.Log("D ctx done")
	case <-time.After(time.Second):
		this.Log("D ticker done")
	}

	// Return any errors
	return ctx.Err()
}

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Graph_001(t *testing.T) {
	type TestNotUnit struct{}
	type TestUnit struct{ graph.Unit }
	type TestNamedUnit struct{ named graph.Unit }

	if g := pkg.New(); g == nil {
		t.Error("Expected non-nil return")
	}
	if g := pkg.New(&TestNotUnit{}); g != nil {
		t.Error("Expected nil return")
	}
	if g := pkg.New(&TestUnit{}); g == nil {
		t.Error("Expected non-nil return")
	}
	if g := pkg.New(&TestNamedUnit{}); g != nil {
		t.Error("Expected nil return")
	}
}

func Test_Graph_002(t *testing.T) {
	type A struct{ graph.Unit }
	type B struct {
		graph.Unit
		*A
	}

	if g := pkg.New(new(A)); g == nil {
		t.Error("Expected non-nil return")
	}
	b := new(B)
	if g := pkg.New(b); g == nil {
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
		*A
		*B
	}
	type X struct {
		graph.Unit
		*X // Circular reference
	}

	if g := pkg.New(new(A)); g == nil {
		t.Error("Expected non-nil return")
	}

	b := new(B)
	if g := pkg.New(b); g == nil {
		t.Error("Expected non-nil return")
	} else if b.A == nil {
		t.Error("Expected non-nil A")
	}

	c := new(C)
	if g := pkg.New(c); g == nil {
		t.Error("Expected non-nil return")
	} else if c.A == nil {
		t.Error("Expected non-nil A")
	} else if c.B == nil {
		t.Error("Expected non-nil B")
	} else if c.B.A == nil {
		t.Error("Expected non-nil A in B")
	} else if c.A != c.B.A {
		t.Error("Expected A to be identical in C and B")
	}

	if g := pkg.New(&X{}); g != nil {
		t.Error("Expected nil X due to circular references", g, g == nil)
	}
}

func Test_Graph_004(t *testing.T) {
	g := pkg.New(new(B))
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
	g := pkg.New(new(B))
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
	g := pkg.New(new(B))
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
	g := pkg.New(new(B), new(C))
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
	g := pkg.New(new(B), new(C), new(A))
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
	g := pkg.New(new(B))
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
	if state.Equals("ABab") == false {
		t.Error("Unexpected New call order:", state.Value(), "...expected:", "ABab")
	}
}

func Test_Graph_010(t *testing.T) {
	g, state := pkg.NewAll(new(A)), NewState(t)
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
	g, state := pkg.NewAny(new(A)), NewState(t)
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
	g, state := pkg.NewAll(new(D), new(D)), NewState(t)
	if err := g.New(state); err != nil {
		t.Error(err)
	}

	// Start running, returns after either D ends, which both end after
	// one second
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Wait for 3 secs (should end anyway after one second)
	now := time.Now()
	if err := g.Run(ctx); err != nil && err != context.DeadlineExceeded {
		t.Error(err)
	}
	if time.Since(now) < time.Second || time.Since(now) > 2*time.Second {
		t.Error("Run did not return after one second")
	}
	// Run will complete randomly for A and D
	ok := state.Equals("ABDDaddb") || state.Equals("ABDDdadb") || state.Equals("ABDDddab")
	if ok == false {
		t.Error("Unexpected New call order:", state.Value(), "...expected:", "ABDDxxxB")
	}
}

func Test_Graph_013(t *testing.T) {
	g, state := pkg.NewAny(new(D), new(D)), NewState(t)
	if err := g.New(state); err != nil {
		t.Error(err)
	}

	// Any should be the same as All in this example since both D's end together
	// but one D (and B) will get a cancel at some unspecified time so the result
	// could be either ABDDADDB or ABDDADBD

	// Start running, returns after either D ends, which both end after
	// one second
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Wait for 3 secs (should end anyway after one second)
	now := time.Now()
	if err := g.Run(ctx); err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		t.Error(err)
	}
	if time.Since(now) < time.Second || time.Since(now) > 2*time.Second {
		t.Error("Run did not return after one second")
	}
	// TODO
}
