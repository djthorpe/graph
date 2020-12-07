package event

import (
	"context"
	"sync"

	"github.com/djthorpe/graph"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type publisher struct {
	graph.Unit
	sync.RWMutex

	q  chan graph.State
	ch []chan graph.State
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (p *publisher) New(graph.State) error {
	p.q = make(chan graph.State)
	return nil
}

func (p *publisher) Dispose() error {
	p.RWMutex.Lock()
	defer p.RWMutex.Unlock()

	close(p.q)
	for _, ch := range p.ch {
		if ch != nil {
			close(ch)
		}
	}
	p.q = nil
	p.ch = nil

	return nil
}

func (p *publisher) Run(ctx context.Context) error {
	for {
		select {
		case evt := <-p.q:
			p.RWMutex.RLock()
			for _, ch := range p.ch {
				if ch != nil {
					ch <- evt
				}
			}
			p.RWMutex.RUnlock()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *publisher) Subscribe() <-chan graph.State {
	p.RWMutex.Lock()
	defer p.RWMutex.Unlock()

	ch := make(chan graph.State)
	p.ch = append(p.ch, ch)
	return ch
}

func (p *publisher) Unsubscribe(ch <-chan graph.State) {
	p.RWMutex.Lock()
	defer p.RWMutex.Unlock()

	for i, other := range p.ch {
		if other == ch {
			close(other)
			p.ch[i] = nil
		}
	}
}

func (p *publisher) Emit(s graph.State) error {
	// Use NullState when evt is nil
	if s == nil {
		p.q <- NullState()
	} else {
		p.q <- s
	}
}
