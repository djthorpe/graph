package main

import (
	"testing"
	"time"

	tool "github.com/djthorpe/graph/pkg/tool"
)

func Assert(t *testing.T, cond bool) {
	if cond == false {
		t.Fatal("Assertation failed")
	}
}

func Test_Obj001(t *testing.T) {
	tool.Test(t, nil, NewObj("A", 100*time.Millisecond), func(obj *Obj) {
		t.Log("Obj=", obj)
		Assert(t, obj.key == "A")
		Assert(t, obj.d == 100*time.Millisecond)
	})
}
