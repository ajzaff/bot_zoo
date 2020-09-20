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

type ModelPositionEvaluator struct {
	r *rand.Rand
}

func NewModelPositionEvaluator() *ModelPositionEvaluator {
	e := &ModelPositionEvaluator{}
	e.SetSeed(time.Now().UnixNano())
	return e
}

func (e *ModelPositionEvaluator) SetSeed(seed int64) {
	e.r = rand.New(rand.NewSource(seed))
}
