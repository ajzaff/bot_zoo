package alpha

import (
	"math/rand"
	"time"
)

type RandomPositionEvaluator struct {
	r *rand.Rand
}

func NewRandomPositionEvaluator() *RandomPositionEvaluator {
	e := &RandomPositionEvaluator{}
	e.SetSeed(time.Now().UnixNano())
	return e
}

func (e *RandomPositionEvaluator) SetSeed(seed int64) {
	e.r = rand.New(rand.NewSource(seed))
}

type ModelBasedPositionEvaluator struct {
	r *rand.Rand
}

func NewModelBasedPositionEvaluator() *ModelBasedPositionEvaluator {
	e := &ModelBasedPositionEvaluator{}
	e.SetSeed(time.Now().UnixNano())
	return e
}

func (e *ModelBasedPositionEvaluator) SetSeed(seed int64) {
	e.r = rand.New(rand.NewSource(seed))
}
