package alpha

import (
	"math/rand"
	"time"
)

type RandomSelectionPolicy struct {
	r *rand.Rand
}

func NewRandomSelectionPolicy() *RandomSelectionPolicy {
	p := &RandomSelectionPolicy{}
	p.SetSeed(time.Now().UnixNano())
	return p
}

func (p *RandomSelectionPolicy) SetSeed(seed int64) {
	p.r = rand.New(rand.NewSource(seed))
}

func (*RandomSelectionPolicy) Select() Move {

}

type ModelbasedPolicy struct {
	r *rand.Rand
}

func NewRandomEvaluator() *RandomEvaluator {
	e := &RandomEvaluator{}
	e.SetSeed(time.Now().UnixNano())
	return e
}

func (e *RandomEvaluator) SetSeed(seed int64) {
	e.r = rand.New(rand.NewSource(seed))
}
