package zoo

import (
	"log"
	"math"
	"math/rand"
	"sort"
)

// Tree represents a game tree for MCTS in memory.
type Tree struct {
	root   *TreeNode           // root node
	tt     *TranspositionTable // tt for looking up transpositions
	p      *Pos                // root position
	sample bool                // sample mode
}

// NewEmptyTree creates a new tree with no root position.
func NewEmptyTree(tt *TranspositionTable) *Tree {
	t := &Tree{
		tt: tt,
	}
	return t
}

// Root returns the root node for this tree.
func (t *Tree) Root() *TreeNode {
	return t.root
}

// SetSample sets the sample mode to sample.
func (t *Tree) SetSample(sample bool) {
	t.sample = sample
}

// UpdateRoot updates the root position to p if p differs from the stored root position.
func (t *Tree) UpdateRoot(p *Pos, model ModelInterface) {
	if t.p == nil || t.p.Hash() != p.Hash() {
		t.p = p
		t.root = t.NewTreeNode(nil, t.p, 0, false, 1, true)
		t.root.rootify(p, model)
	}
}

// Select the next node to expand at position p.
func (t *Tree) Select(p *Pos) (*TreeNode, *Pos) {
	p = p.Clone()
	n := t.root
	for {
		if len(n.children) == 0 {
			return n, p
		}
		for _, c := range n.children {
			c.computePriority(0)
		}
		sort.Stable(byPriority(n.children))
		n = n.children[0]
		s, pass := n.Step()
		if pass {
			p.Pass()
		} else {
			p.Step(s)
		}
	}
}

// Reset clears all nodes from the tree.
func (t *Tree) Reset() {
	t.root = nil
	t.p = nil
}

// RetainOptimalSubtree removes all suboptimal subtrees and resets
// the frontier. After calling this method, the tree is ready to evaluate the next turn.
func (t *Tree) RetainOptimalSubtree(p *Pos, n *TreeNode, model ModelInterface) {
	// Reset the tree root;
	// Clear the step and "rootify".
	t.root = n
	n.rootify(p, model)
}

// BestMove returns the best move from the tree after all runs have been performed.
// This is equivalent to the path from root with the greatest number of playouts.
// If the best move would not be legal (this is possible given a terminal root node)
// nil and false are returned instead.
func (t *Tree) BestMove(r *rand.Rand) (m Move, v Value, n *TreeNode, ok bool) {
	n = t.root
	for n.first && len(n.children) > 0 {
		sort.Stable(byRuns(n.children))
		n = n.children[0]
		step, pass := n.Step()
		if pass {
			t.p.Pass()
			break
		}
		cap := t.p.Step(step)
		m = append(m, step)
		if cap.Capture() {
			m = append(m, cap)
		}
	}
	if len(m) > 0 {
		return m, Value(float64(t.root.children[0].Weight()) / float64(t.root.children[0].Runs())), n, true
	}
	return nil, 0, n, false
}

// RootChildren returns a shallow copy of the children at this root position.
func (t *Tree) RootChildren() []*TreeNode {
	if t.root == nil {
		return nil
	}
	children := make([]*TreeNode, len(t.root.children))
	copy(children, t.root.children)
	return children
}

// Debug the state of the tree.
func (t *Tree) Debug(l *log.Logger) {
	l.Println("tree:")
	n := t.root
	for i := 0; len(n.children) > 0; i++ {
		l.Printf("  depth=%d:", i)
		sort.Stable(byRuns(n.children))
		for _, c := range n.children {
			l.Printf("    step=%s [%f] P=%f runs=%d weight=%f", c.step, float64(c.weight)/float64(c.runs), c.priority, c.runs, c.weight)
		}
		n = n.children[0]
	}
}

// TreeNode represents a game tree node for MCTS in memory.
type TreeNode struct {
	t        *Tree       // parent tree containing the frontier heap
	side     Value       // side-to-move multipier; can be 1 or -1.
	eval     Value       // terminal eval; [-1 0 1]
	weight   Value       // cumulative value of this state; divide by Runs to normalize.
	runs     uint32      // number of runs through this node.
	policy   []float32   // policy from the model, if run.
	priority float64     // computed priority ordering for this node based on value, policy, and runs.
	step     Step        // step played to arrive at this position.
	pass     bool        // pass was played to arrive at this position.
	parent   *TreeNode   // parent node.
	first    bool        // first turn; candidate for bestmove.
	children []*TreeNode // expanded children of this node; used on first turn only to recover bestmove.
}

// NewTreeNode creates a new game tree node for p with initial stats populated from the tt.
func (t *Tree) NewTreeNode(parent *TreeNode, p *Pos, step Step, pass bool, side Value, first bool) *TreeNode {
	e := &TreeNode{
		t:      t,
		side:   side,
		eval:   side * p.Terminal(),
		step:   step,
		pass:   pass,
		parent: parent,
		first:  first,
	}
	return e
}

// Step returns the step for this node or pass.
func (n *TreeNode) Step() (s Step, pass bool) {
	return n.step, n.pass
}

// StepIndex returns the step index for this node.
func (n *TreeNode) StepIndex() uint8 {
	if n.pass {
		return passIndex
	}
	return n.step.Index()
}

// HasParent returns true if n has a non-nil parent.
func (n *TreeNode) HasParent() bool {
	return n.parent != nil
}

// Runs returns the number of MCTS runs propagated through this node.
func (n *TreeNode) Runs() uint32 {
	return n.runs
}

// ParentRuns returns the number of MCTS runs propagated through n's parent.
func (n *TreeNode) ParentRuns() uint32 {
	if p := n.parent; p != nil {
		return p.Runs()
	}
	return 0
}

// Weight returns the total value of node n.
// Divide by Runs to normalize.
func (n *TreeNode) Weight() Value {
	return n.weight
}

// Terminal returns true if this node is a terminal node.
func (n *TreeNode) Terminal() bool {
	return n.eval.Terminal()
}

// Policy returns the step policy for this node.
func (n *TreeNode) Policy() []float32 {
	return n.policy
}

// rootify resets this node to create an expanded root node.
func (n *TreeNode) rootify(p *Pos, model ModelInterface) {
	n.step = 0
	n.side = 1
	n.first = true
	n.parent = nil
	n.children = n.children[:0]
	if e, found := n.t.tt.Probe(p.Hash()); found {
		// TT Hit
		n.weight = e.Weight
		n.runs = e.Runs
		n.policy = e.Policy
	} else {
		n.policy = policyPool.Get().([]float32)
	}
	n.Expand(p, model)
}

func (n *TreeNode) expandOne(p *Pos, model ModelInterface, step Step, pass bool, side, childSide Value) {

	child := n.t.NewTreeNode(n, p, step, pass, childSide, n.first && side == childSide)
	n.children = append(n.children, child)

	var (
		v    Value
		runs = uint32(1)
	)

	if t := p.Terminal(); t != 0 {
		// Terminal value
		v = child.side * t
	} else if e, found := n.t.tt.Probe(p.Hash()); found {
		// TT Hit
		child.policy = e.Policy
		v = e.Weight
		runs = e.Runs
	} else {
		// TT Miss: Evaluate new node.
		model.EvaluatePosition(p)
		v = child.side * Value(model.Value())

		// Save to tt.
		e.Save(p.Hash(), n.t.tt.GlobalAge(), v, 1, child.policy)
	}

	// Backprop
	child.Backprop(v, runs)

	if child.policy == nil {
		child.policy = policyPool.Get().([]float32)
	}
}

// Expand expands the node by generating all legal child nodes from this position.
// All generated children are added to the frontier while n is removed from the frontier.
func (n *TreeNode) Expand(p *Pos, model ModelInterface) {
	// Pos is not at n.
	// Generate pseudo-legal steps.
	var (
		stepList    = NewStepList(64)
		hasChildren bool
	)
	stepList.Generate(p)
	for i := 0; i < stepList.Len(); i++ {

		step := stepList.At(i)
		if !p.Legal(step.Step) {
			continue
		}
		hasChildren = true

		initSide := p.Side()

		p.Step(step.Step)

		childSide := n.side
		if initSide != p.Side() {
			childSide = -childSide
		}

		n.expandOne(p, model, step.Step, false, n.side, childSide)

		p.Unstep()
	}

	if n.t.p.CanPass() {
		hasChildren = true
		p.Pass()
		n.expandOne(p, model, 0, true, n.side, -n.side)
		p.Unpass()
	}

	if !hasChildren {
		n.Backprop(n.side*Loss, 1)
	}
}

const c = 1.41421

func (n *TreeNode) computePriority(deltaN uint32) {
	var x float64
	if n.HasParent() {
		N := n.ParentRuns() + deltaN
		x = c * math.Sqrt(math.Log(float64(N))/float64(n.Runs()))
	}
	n.priority = x + float64(n.Weight()) + float64(n.policy[n.StepIndex()])
}

const largeBackprop = 1000000000

// Backprop propagates the value v representing n runs to parents of this node.
// Fixes the frontier heap.
func (n *TreeNode) Backprop(v Value, runs uint32) {
	// Ensure we play winning moves.
	if n.first && n.Terminal() {
		v *= largeBackprop
		runs *= largeBackprop
	}
	n.weight += v
	n.runs += runs
	for p := n.parent; p != nil; p = p.parent {
		p.weight += v
		p.runs += runs
	}
}

type byRuns []*TreeNode

func (a byRuns) Len() int           { return len(a) }
func (a byRuns) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRuns) Less(i, j int) bool { return a[i].Runs() > a[j].Runs() }

type byPriority []*TreeNode

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool { return a[i].priority > a[j].priority }
