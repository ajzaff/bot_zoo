package zoo

import (
	"fmt"
	"log"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// Model wraps a Tensorflow SavedModel.
type Model struct {
	m      *tf.SavedModel
	input  []float32
	policy []float32
	value  []float32
}

// NewModel loads the saved_model.pb from the saved_models directory or returns an error.
func NewModel() (*Model, error) {
	m, err := tf.LoadSavedModel("data/saved_models/bot_alpha_zoo-16", []string{"serve"}, nil)
	if err != nil {
		return nil, err
	}
	for _, op := range m.Graph.Operations() {
		dtype, _ := op.Attr("dtype")
		log.Println(op.Type(), op.Name(), op.NumInputs(), op.NumOutputs(), dtype)
	}
	model := &Model{
		m:      m,
		input:  make([]float32, modelInputSize),
		policy: policyPool.Get().([]float32),
		value:  make([]float32, 1),
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

const (
	modelInputName   = "alpha_convolution_layer/input_layer_conv/kernel/Read/ReadVariableOp"
	valueOutputName  = "alpha_value_head/value/kernel/Read/ReadVariableOp"
	policyOutputName = "alpha_policy_head/policy/kernel/Read/ReadVariableOp"
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

	opInput := m.m.Graph.Operation(modelInputName)
	if opInput == nil {
		panic("opInput == nil")
	}
	opPolicy := m.m.Graph.Operation(policyOutputName)
	if opInput == nil {
		panic("opPolicy == nil")
	}
	opValue := m.m.Graph.Operation(valueOutputName)
	if opInput == nil {
		panic("opValue == nil")
	}

	ts, err := m.m.Session.Run(map[tf.Output]*tf.Tensor{
		opInput.Output(0): tIn,
	}, []tf.Output{opPolicy.Output(0), opValue.Output(0)}, nil)
	if err != nil {
		panic(err)
	}

	_ = ts
	for _, t := range ts {
		fmt.Println(t.DataType(), t.Shape())
	}
	fmt.Println(len(ts))
}

// SetSeed is a noop in the real model.
// Provided for ModelInterface.
func (m *Model) SetSeed(seed int64) {}

// Value returns the value estimate from the last model run.
func (m *Model) Value() float32 {
	return m.value[0]
}

// Policy populates the policy values with logits from the last model run.
func (m *Model) Policy(policy []float32) {
	copy(policy, m.policy)
}

// Close closes the model session.
func (m *Model) Close() error {
	return m.m.Session.Close()
}
