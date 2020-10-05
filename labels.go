package zoo

import (
	"sync"

	expb "github.com/ajzaff/bot_zoo/proto"
)

var policyPool = sync.Pool{
	New: func() interface{} {
		return make([]float32, 232)
	},
}

func resetLabels(ex *expb.Example) {
	ex.Policy = make(map[uint32]float32)
	ex.Value = 0
}

func labelIndex(s Step, pass bool) uint32 {
	if pass {
		return passIndex
	}
	return uint32(s.Index())
}

// PolicyLabels fills in policy labels from the completed search tree.
// The policy labels have shape (232,).
// Labels should be called before the tree is pruned (i.e. before
// calling RetainBestMove).
// lateralLabels is filled with step labels mirrored laterally
// (for dataset augmentation).
// The label value is in logits but the model requires probabilities.
// Calling softmax on the output in Tensorflow should do the trick.
func PolicyLabels(t *Tree, ex *expb.Example) {
	resetLabels(ex)

	for _, n := range t.RootChildren() {
		s, pass := n.Step()
		ex.Policy[labelIndex(s, pass)] = float32(n.Runs())
	}
}

// ValueLabel fills in the value from the tree.
func ValueLabel(t *Tree, ex *expb.Example) {
	v := float32(-1)
	if root := t.Root(); root != nil {
		v = float32(Value(float64(root.Weight()) / float64(root.Runs()))) // TODO(ajzaff): make numerically stable
	}
	ex.Value = v
}
