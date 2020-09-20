package alpha

import (
	"math/rand"
	"time"
)

type RandomStepPolicy struct {
	r *rand.Rand
}

func NewRandomStepPolicy() *RandomStepPolicy {
	p := &RandomStepPolicy{}
	p.SetSeed(time.Now().UnixNano())
	return p
}

func (p *RandomStepPolicy) SetSeed(seed int64) {
	p.r = rand.New(rand.NewSource(seed))
}

func (p *RandomStepPolicy) Policy(values []float64) {
	for i := range values {
		values[i] = 2*p.r.Float64() - 1
	}
}

type ModelStepPolicy struct {
	r *rand.Rand
}

func NewModelStepPolicy() *ModelStepPolicy {
	p := &ModelStepPolicy{}
	p.SetSeed(time.Now().UnixNano())
	return p
}

func (p *ModelStepPolicy) SetSeed(seed int64) {
	p.r = rand.New(rand.NewSource(seed))
}

func (p *ModelStepPolicy) Policy(values []float64) {
}
