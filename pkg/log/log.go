package log

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"

	graph "github.com/djthorpe/graph"
)

type Log struct {
	graph.Unit
	sync.Mutex

	D bool       // D comtains debug flag
	T *testing.T // T contains testing context
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (*Log) New(graph.State) error {
	log.SetFlags(log.Ltime)
	return nil
}

func (this *Log) Run(ctx context.Context) error {
	this.Print("->Logger Run")
	defer this.Print("<-Logger Run")
	<-ctx.Done()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Log) Print(args ...interface{}) {
	this.Lock()
	defer this.Unlock()
	if this.T != nil {
		this.T.Log(args...)
	} else {
		log.Print(args...)
	}
}

func (this *Log) Debug(args ...interface{}) {
	if this.IsDebug() {
		this.Print(args...)
	}
}

func (this *Log) Printf(fmt string, args ...interface{}) {
	this.Lock()
	defer this.Unlock()
	if this.T != nil {
		this.T.Logf(fmt, args...)
	} else {
		log.Printf(fmt, args...)
	}
}

func (this *Log) Debugf(fmt string, args ...interface{}) {
	if this.IsDebug() {
		this.Printf(fmt, args...)
	}
}

func (this *Log) IsDebug() bool {
	return this.D
}

func (this *Log) Test() *testing.T {
	return this.T
}

func (this *Log) SetTest(t *testing.T) {
	this.D = true
	this.T = t
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Log) String() string {
	str := "<log"
	if this == nil {
		str += " nil"
	} else if debug := this.IsDebug(); debug {
		str += fmt.Sprint(" debug")
		if t := this.Test(); t != nil {
			str += fmt.Sprintf(" test=%q", t.Name())
		}
	}
	return str + ">"
}
