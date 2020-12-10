package graph_test

import (
	"sync"
	"testing"

	"github.com/djthorpe/graph/pkg/graph"
)

func Test_Counter_001(t *testing.T) {
	c := graph.NewCounter()
	if c.IsZero() == false {
		t.Error("Expected counter to be zero")
	}
}

func Test_Counter_002(t *testing.T) {
	c := graph.NewCounter()
	for i := 0; i < 1000; i++ {
		c.Dec(true)
	}
	for i := 0; i < 1000; i++ {
		c.Inc(true)
	}
	if c.IsZero() == false {
		t.Error("Expected counter to be zero")
	}
}

func Test_Counter_003(t *testing.T) {
	c := graph.NewCounter()
	for i := 0; i < 1000; i++ {
		c.Dec(true)
		c.Inc(true)
	}
	if c.IsZero() == false {
		t.Error("Expected counter to be zero")
	}
}

func Test_Counter_004(t *testing.T) {
	c := graph.NewCounter()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Dec(true)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc(true)
		}()
	}
	wg.Wait()
	if c.IsZero() == false {
		t.Error("Expected counter to be zero")
	}
}
