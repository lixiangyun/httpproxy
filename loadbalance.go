package main

import (
	"math/rand"
)

type LoadBalance interface {
	Pick() string
}

type LBRR struct {
	list []string
	max  int
	idx  int
}

func NewLBRR(list []string) LoadBalance {
	cp := make([]string, len(list))
	copy(cp, list)
	return &LBRR{list: cp, max: len(list)}
}

func (l *LBRR) Pick() string {
	l.idx++
	return l.list[l.idx%l.max]
}

type LBRandom struct {
	list []string
	max  int
}

func NewLBRandom(list []string) LoadBalance {
	cp := make([]string, len(list))
	copy(cp, list)
	return &LBRandom{list: cp, max: len(list)}
}

func (r *LBRandom) Pick() string {
	return r.list[rand.Int()%r.max]
}

func NewLB(tp string, list []string) LoadBalance {
	switch tp {
	case "rr":
		{
			return NewLBRR(list)
		}
	case "randmon":
		{
			return NewLBRandom(list)
		}
	}
	return nil
}
