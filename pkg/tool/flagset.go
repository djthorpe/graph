package tool

import (
	"flag"
)

type FlagSet struct {
	*flag.FlagSet
}

func (s *FlagSet) Value() interface{} {
	return s.FlagSet
}

func NewFlagset(name string) *FlagSet {
	return &FlagSet{
		flag.NewFlagSet(name, flag.ContinueOnError),
	}
}
