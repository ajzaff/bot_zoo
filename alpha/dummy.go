package alpha

import (
	"math"
	"math/rand"
)

// Dummy implements a dummy model giving random evaluations.
type Dummy struct {
	r      *rand.Rand
	value  float32
	policy []float32
}

// NewDummy creates a new dummy evaluator.
func NewDummy() *Dummy {
	m := &Dummy{}
	m.SetSeed(1337)
	return m
}

// SetSeed reseeds the random state for this dummy model.
func (m *Dummy) SetSeed(seed int64) {
	m.r = rand.New(rand.NewSource(seed))
}

// EvaluatePosition regenerates random outputs for the position.
func (m *Dummy) EvaluatePosition() {
	if m.policy == nil {
		m.policy = make([]float32, 256)
	}
	m.value = float32(math.Tanh(m.r.NormFloat64()))
	for i := range m.policy {
		m.policy[i] = float32(-m.r.ExpFloat64())
	}
}

// Value returns a randomly generated value estimate for the last position.
func (m *Dummy) Value() float32 {
	return m.value
}

// Policy populates the policy with randomly generated logits for the last position.
func (m *Dummy) Policy(policy []float32) {
	copy(policy, m.policy)
}
