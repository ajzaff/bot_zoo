package zoo

import (
	"io/ioutil"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// Model wraps a Tensorflow SavedModel.
type Model struct {
	g      *tf.Graph
	sess   *tf.Session
	input  []float32
	policy [][]float32
	value  [][]float32
}

// NewModel loads the saved_model.pb from the saved_models directory or returns an error.
func NewModel() (*Model, error) {
	bs, err := ioutil.ReadFile("data/saved_models/bot_alpha_zoo-16.pb")
	if err != nil {
		return nil, err
	}
	g := tf.NewGraph()
	if err := g.Import(bs, ""); err != nil {
		return nil, err
	}
	sess, err := tf.NewSession(g, nil)
	if err != nil {
		return nil, err
	}
	model := &Model{
		g:      g,
		sess:   sess,
		input:  make([]float32, modelInputSize),
		policy: [][]float32{policyPool.Get().([]float32)},
		value:  make([][]float32, 1, 232),
	}
	return model, nil
}

const (
	modelInputSize        = 8 * 8 * 21
	modelOutputPolicySize = 232
)

var (
	modelInputShape   = []int64{1, 8, 8, 21}
	valueOutputShape  = []int64{1, 1}
	policyOutputShape = []int64{1, 232}
)

// Horribly named nodes, not for lack of trying.
const (
	modelInputName   = "x"
	valueOutputName  = "bot_alpha_zoo-16/value_/dense_1/Tanh"
	policyOutputName = "bot_alpha_zoo-16/policy_/dense_3/BiasAdd"
)

// EvaluatePosition initiates a model run against the new positon.
func (m *Model) EvaluatePosition(p *Pos) {
	tIn, err := tf.NewTensor(m.input)
	if err != nil {
		panic(err)
	}
	tIn.Reshape(modelInputShape)

	tPolicy, err := tf.NewTensor(m.policy)
	if err != nil {
		panic(err)
	}
	tPolicy.Reshape(policyOutputShape)

	tValue, err := tf.NewTensor(m.value)
	if err != nil {
		panic(err)
	}
	tValue.Reshape(valueOutputShape)

	opInput := m.g.Operation(modelInputName)
	if opInput == nil {
		panic("opInput == nil")
	}
	opPolicy := m.g.Operation(policyOutputName)
	if opInput == nil {
		panic("opPolicy == nil")
	}
	opValue := m.g.Operation(valueOutputName)
	if opInput == nil {
		panic("opValue == nil")
	}

	ts, err := m.sess.Run(map[tf.Output]*tf.Tensor{
		opInput.Output(0): tIn,
	}, []tf.Output{opPolicy.Output(0), opValue.Output(0)}, nil)
	if err != nil {
		panic(err)
	}
	m.policy = ts[0].Value().([][]float32)
	m.value = ts[1].Value().([][]float32)
}

// SetSeed is a noop in the real model.
// Provided for ModelInterface.
func (m *Model) SetSeed(seed int64) {}

// Value returns the value estimate from the last model run.
func (m *Model) Value() float32 {
	return m.value[0][0]
}

// Policy populates the policy values with logits from the last model run.
func (m *Model) Policy(policy []float32) {
	copy(policy, m.policy[0])
}

// Close closes the model session.
func (m *Model) Close() error {
	return m.sess.Close()
}
