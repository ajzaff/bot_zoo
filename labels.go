package zoo

func resetLabels(labels, lateralLabels []float32) {
	for i := range labels {
		labels[i] = 0
		lateralLabels[i] = 0
	}
}

func labelIndex(s Step, pass bool, lateral bool) int {
	if pass {
		return 231
	}
	if lateral {
		s = s.MirrorLateral()
	}
	return int(s.Index())
}

// PolicyLabels fills in policy labels from the completed search tree.
// The policy labels have shape (231,).
// Labels should be called before the tree is pruned (i.e. before
// calling RetainBestMove).
// lateralLabels is filled with step labels mirrored laterally
// (for dataset augmentation).
// The label value is in logits but the model requires probabilities.
// Calling softmax on the output in Tensorflow should do the trick.
func PolicyLabels(t *Tree, labels, lateralLabels []float32) {
	resetLabels(labels, lateralLabels)

	for _, n := range t.RootChildren() {
		s, pass := n.Step()
		labels[labelIndex(s, pass, false)] = float32(n.Runs())
		lateralLabels[labelIndex(s, pass, true)] = float32(n.Runs())
	}
}

// ValueLabel fills in the value from the tree.
func ValueLabel(t *Tree, value []float32) {
	v := float32(-1)
	if root := t.Root(); root != nil {
		v = float32(Value(float64(root.Weight()) / float64(root.Runs()))) // TODO(ajzaff): make numerically stable
	}
	value[0] = v
}
