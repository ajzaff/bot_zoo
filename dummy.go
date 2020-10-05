package zoo

import (
	"math"
	"math/rand"
)

// DummyModel implements a dummy model giving random evaluations.
type DummyModel struct {
	r      *rand.Rand
	value  float32
	policy []float32
}

// NewDummyModel creates a new dummy evaluator.
func NewDummyModel() *DummyModel {
	m := &DummyModel{}
	m.SetSeed(1337)
	return m
}

// SetSeed reseeds the random state for this dummy model.
func (m *DummyModel) SetSeed(seed int64) {
	m.r = rand.New(rand.NewSource(seed))
}

// EvaluatePosition regenerates random outputs for the position.
func (m *DummyModel) EvaluatePosition(p *Pos) {
	if m.policy == nil {
		m.policy = make([]float32, 232)
	}
	m.value = float32(math.Tanh(0.5 * m.r.NormFloat64()))
	for i := range m.policy {
		m.policy[i] = float32(-m.r.ExpFloat64())
	}
}

// Value returns a randomly generated value estimate for the last position.
func (m *DummyModel) Value() float32 {
	return m.value
}

// Policy populates the policy with randomly generated logits for the last position.
func (m *DummyModel) Policy(policy []float32) {
	copy(policy, m.policy)
}
