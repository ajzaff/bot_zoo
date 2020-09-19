package alpha

import (
	"math/rand"
	"time"
)

type RandomMovePolicy struct {
	r *rand.Rand
}

func NewRandomMovePolicy() *RandomMovePolicy {
	p := &RandomMovePolicy{}
	p.SetSeed(time.Now().UnixNano())
	return p
}

func (p *RandomMovePolicy) SetSeed(seed int64) {
	p.r = rand.New(rand.NewSource(seed))
}

func (*RandomMovePolicy) Select() {
}

type ModelBasedMovePolicy struct {
	r *rand.Rand
}

func NewModelBasedMovePolicy() *ModelBasedMovePolicy {
	e := &ModelBasedMovePolicy{}
	e.SetSeed(time.Now().UnixNano())
	return e
}

func (e *ModelBasedMovePolicy) SetSeed(seed int64) {
	e.r = rand.New(rand.NewSource(seed))
}

func (*ModelBasedMovePolicy) Select() {
}
