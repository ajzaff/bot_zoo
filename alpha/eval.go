package alpha

import (
	"math/rand"
	"time"
)

type RandomEvaluator struct {
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

type ModelBasedEvaluator struct {
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
