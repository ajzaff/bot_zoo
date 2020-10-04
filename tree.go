package zoo

import (
	"container/heap"
	"math"
	"math/rand"
	"sort"
)

// Tree represents a game tree for MCTS in memory.
type Tree struct {
	root     *TreeNode           // root node
	frontier []*TreeNode         // frontier heap
	tt       *TranspositionTable // tt for looking up transpositions
	p        *Pos                // root position
	sample   bool                // sample mode
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
func (t *Tree) UpdateRoot(p *Pos) {
	if t.p == nil || t.p.Hash() != p.Hash() {
		t.p = p
		t.root = t.NewTreeNode(nil, t.p, 0, false, 1, true)
		t.root.Expand()
	}
}

// Push the tree node x onto the frontier.
// For heap.Heap.
func (t *Tree) Push(x interface{}) {
	n := t.Len()
	t.frontier = append(t.frontier, x.(*TreeNode))
	x.(*TreeNode).idx = n
}

// Select pops and returns the most promising node from the frontier.
func (t *Tree) Select() *TreeNode {
	return t.Pop().(*TreeNode)
}

// Pop returns the top node from the frontier.
// For heap.Heap.
func (t *Tree) Pop() interface{} {
	n := t.Len() - 1
	x := t.frontier[n]
	t.frontier = t.frontier[:n]
	x.idx = -1
	return x
}

// Swap the frontier nodes i and j.
// For heap.Heap.
func (t *Tree) Swap(i, j int) {
	t.frontier[i], t.frontier[j] = t.frontier[j], t.frontier[i]
	t.frontier[i].idx = i
	t.frontier[j].idx = j
}

// Len returns the length of the frontier.
// For heap.Heap.
func (t *Tree) Len() int {
	return len(t.frontier)
}

// Less orders the frontier by priority.
// For heap.Heap.
func (t *Tree) Less(i, j int) bool {
	return t.frontier[i].priority > t.frontier[j].priority
}

// RetainOptimalSubtree removes all suboptimal subtrees and resets
// the frontier. After calling this method, the tree is ready to evaluate the next turn.
func (t *Tree) RetainOptimalSubtree(n *TreeNode) {
	// Clear the frontier heap.
	t.frontier = t.frontier[:0]
	// Reset the tree root;
	// Clear the step and "rootify".
	t.root = n
	n.rootify()
}

// BestMove returns the best move from the tree after all runs have been performed.
// This is equivalent to the path from root with the greatest number of playouts.
// If the best move would not be legal (this is possible given a terminal root node)
// nil and false are returned instead.
func (t *Tree) BestMove(r *rand.Rand) (m Move, v Value, n *TreeNode, ok bool) {
	n = t.root
	for n.first && len(n.children) > 0 {
		sort.Stable(byRuns(n.children))
		var i int
		if t.p.MoveNum() > 1 && t.sample { // TODO(ajzaff): Allow exploration in setup.
			// Sample among the moves, use cumulative runs as prior probability.
			sum := 0
			for _, child := range n.children {
				sum += int(child.Runs())
			}
			if sum > 0 {
				x := r.Intn(sum)
				for j, child := range n.children {
					if x -= int(child.Runs()); x <= 0 {
						i = j
						break
					}
				}
			}
		}
		n = n.children[i]
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
		return m, Value(float64(n.Weight()) / float64(n.Runs())), n, true
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

// TreeNode represents a game tree node for MCTS in memory.
type TreeNode struct {
	t        *Tree       // parent tree containing the frontier heap
	idx      int         // frontier heap index or -1
	side     Value       // side-to-move multipier; can be 1 or -1.
	weight   Value       // cumulative value of this state; divide by Runs to normalize.
	runs     int         // number of runs through this node.
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
		idx:    -1,
		side:   side,
		step:   step,
		pass:   pass,
		parent: parent,
		first:  first,
		runs:   1,
	}
	if v := p.Terminal(); v != 0 {
		e.weight = side * v
	}
	return e
}

// Step returns the step for this node or pass.
func (n *TreeNode) Step() (s Step, pass bool) {
	return n.step, n.pass
}

// HasParent returns true if n has a non-nil parent.
func (n *TreeNode) HasParent() bool {
	return n.parent != nil
}

// Runs returns the number of MCTS runs propagated through this node.
func (n *TreeNode) Runs() int {
	return n.runs
}

// ParentRuns returns the number of MCTS runs propagated through n's parent.
func (n *TreeNode) ParentRuns() int {
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
	return n.Weight() != 0 && int(math.Abs(float64(n.Weight()))) == n.Runs()
}

// rootify resets this node to create an expanded root node.
func (n *TreeNode) rootify() {
	n.step = 0
	n.idx = -1
	n.side = 1
	n.first = true
	n.parent = nil
	n.children = n.children[:0]
	n.Expand()
}

// fastForward plays out the root position to the position at n.
func (n *TreeNode) fastForward(root *Pos) {
	if n.step != 0 {
		defer root.Step(n.step)
	}
	for p := n.parent; p != nil; p = p.parent {
		defer func(s Step) {
			if s != 0 {
				root.Step(s)
			}
		}(p.step)
	}
}

// Expand expands the node by generating all legal child nodes from this position.
// All generated children are added to the frontier while n is removed from the frontier.
func (n *TreeNode) Expand() {
	// Fast-forward to this position.
	p := n.t.p.Clone()
	n.fastForward(p)
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

		child := n.t.NewTreeNode(n, p, step.Step, false, childSide, n.first && initSide == p.Side())
		if n.first {
			n.children = append(n.children, child)
		}
		if !n.Terminal() {
			heap.Push(n.t, child)
		}

		p.Unstep()
	}

	if n.first && n.t.p.CanPass() {
		p.Pass()
		child := n.t.NewTreeNode(n, p, 0, true, -n.side, false)
		n.children = append(n.children, child)
		p.Unpass()
	}

	if !hasChildren {
		n.weight = n.side * Loss
	}
}

// Evaluate uses the model to compute the value of the position.
func (n *TreeNode) Evaluate(model ModelInterface) Value {
	p := n.t.p.Clone()
	n.fastForward(p)
	model.EvaluatePosition(p)
	return Value(model.Value())
}

const c = 1.41421

func (n *TreeNode) computePriority(deltaN int) {
	var x float64
	if n.HasParent() {
		N := n.ParentRuns() + deltaN
		x = c * math.Sqrt(math.Log(float64(N))/float64(n.Runs()))
	}
	n.priority = x + float64(n.Weight())
}

// Backprop propagates the value v representing n runs to parents of this node.
// Fixes the frontier heap.
func (n *TreeNode) Backprop(v Value, runs int) {
	n.weight += v
	n.runs += runs
	n.computePriority(runs)
	if n.idx != -1 {
		heap.Fix(n.t, n.idx)
	}
	for p := n.parent; p != nil; p = p.parent {
		p.weight += v
		p.runs += runs
		p.computePriority(runs)
		if n.idx != -1 {
			heap.Fix(p.t, p.idx)
		}
	}
}

type byRuns []*TreeNode

func (a byRuns) Len() int           { return len(a) }
func (a byRuns) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRuns) Less(i, j int) bool { return a[i].Runs() > a[j].Runs() }
