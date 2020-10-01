package alpha

import (
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// Model wraps a Keras SavedModel.
type Model struct {
	m *tf.SavedModel
}

// NewModel loads the saved_model.pb from the saved_models directory or returns an error.
func NewModel() (*Model, error) {
	m, err := tf.LoadSavedModel("data/saved_models/bot_alpha_zoo-16", []string{"serve"}, nil)
	if err != nil {
		return nil, err
	}
	model := &Model{
		m: m,
	}
	return model, nil
}

// EvaluatePosition initiates a model run against the new positon.
func (m *Model) EvaluatePosition() {

}

// Value returns the value estimate from the last model run.
func (m *Model) Value() float32 {
	return 0
}

// Policy populates the policy values with logits from the last model run.
func (m *Model) Policy(values []float64) {
}
