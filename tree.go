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

// Less orders the frontier by UCB1.
// For heap.Heap.
func (t *Tree) Less(i, j int) bool {
	return t.frontier[i].ucb1 > t.frontier[j].ucb1
}

// retainOptimalSubtree removes all suboptimal subtrees and resets
// the frontier to only nodes under the optimal subtree n. After calling
// this method, the tree is ready to evaluate the next turn.
func (t *Tree) retainOptimalSubtree(n *TreeNode) {
	// Clear the frontier heap.
	t.frontier = t.frontier[:0]
	// Reset the tree root;
	// Clear the step and "rootify".
	t.root = n
	n.rootify()
}

// RetainBestMove returns the best move from the tree after all runs have been performed.
// This is equivalent to the path from root with the greatest number of playouts.
// If the best move would not be legal (this is possible given a terminal root node)
// nil and false are returned instead.
// A call to RetainBestMove removes suboptimal child nodes as a side-effect.
// See retainChildren for details.
func (t *Tree) RetainBestMove(r *rand.Rand) (m Move, v Value, ok bool) {
	n := t.root
	for n.first && len(n.children) > 0 {
		sort.Stable(byPlayouts(n.children))
		var i int
		if t.sample {
			// Sample among the moves, use cumulative value as prior probability.
			sum := 0
			for _, child := range n.children {
				sum += int(child.value)
			}
			if sum > 0 {
				x := r.Intn(sum)
				for j, child := range n.children {
					if x -= int(child.value); x <= 0 {
						i = j
						break
					}
				}
			}
		} else {
			// Choose the best move, use Value as a tie breaker.
			bestValue := n.children[0].value
			for j := 0; j < len(n.children) && n.children[i].Playouts() == n.children[j].Playouts(); j++ {
				if v := n.children[j].Value(); v > bestValue {
					i = j
					bestValue = v
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
	t.retainOptimalSubtree(n)
	if len(m) > 0 {
		return m, n.Value(), true
	}
	return nil, 0, false
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
	eval     Value       // theoretical eval; if non-0 we do not do playouts.
	value    Value       // cumulative playout value; divide by playouts to normalize.
	playouts int         // number of playouts through this node.
	ucb1     float64     // computed UCB1
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
		eval:   p.Terminal(),
		step:   step,
		pass:   pass,
		parent: parent,
		first:  first,
	}
	e.eval = p.Terminal()
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

// Playouts returns the number of simulations propagated through this node.
func (n *TreeNode) Playouts() int {
	return n.playouts
}

// ParentPlayouts returns the number of simulations propagated through this node's parent.
func (n *TreeNode) ParentPlayouts() int {
	if p := n.parent; p != nil {
		return p.Playouts()
	}
	return 0
}

// Value computes the estimated win value of the node.
func (n *TreeNode) Value() Value {
	if n.eval != 0 {
		return n.eval
	}
	if n.playouts > 0 {
		return n.value / Value(n.playouts)
	}
	return 0
}

// Eval returns the theoretical value of the node (-1, 0, +1).
func (n *TreeNode) Eval() Value {
	return n.eval
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
		if !child.eval.Terminal() {
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
		n.eval = n.side * Loss
	}
}

// Simulate runs a number of playouts from the position to estimate the value.
// It returns the cumulative value after n playouts.
func (n *TreeNode) Simulate(r *rand.Rand, numPlayouts int) Value {
	p := n.t.p.Clone()
	n.fastForward(p)
	stepList := NewStepList(40 * 64)

	var v Value
	for ; numPlayouts > 0; numPlayouts-- {
		v += n.Playout(p, stepList, r)
	}
	return v
}

// Playout runs a single random playout from the position and returns the value.
// In playout, only a subset of steps are generated and play continues for a fixed
// number of steps only, unless a terminal node is reached. The depth limitation
// is to mitigate the effect of random attacking being much stronger than random
// defense (and thus vastly overestimating the value of rabbit pushes).
// If at any time, no legal steps were generated, or we have reached the depth
// limit before the end of the game we stop and return 0 + a small random bias.
func (n *TreeNode) Playout(p *Pos, stepList *StepList, r *rand.Rand) Value {

	for side := n.side; ; {
		// Is this a terminal node? Return the value immediately.
		if v := p.Terminal(); v != 0 {
			return side * v
		}

		// Generate the steps for the next node.
		stepList.Truncate(0)
		stepList.Generate(p)

		// Test and truncate illegal steps:
		j := 0
		for i := 0; i < stepList.Len(); i++ {
			step := stepList.At(i)
			if p.Legal(step.Step) {
				stepList.Swap(i, j)
				j++
			}
		}
		stepList.Truncate(j)

		// No steps generated? Stop the search.
		if stepList.Len() == 0 {
			break
		}

		// Continue the playout with the chosen step:
		step := stepList.At(r.Intn(stepList.Len()))
		initSide := p.Side()

		p.Step(step.Step)
		if p.Side() != initSide {
			side = -side
		}
		defer p.Unstep()
	}

	return 0
}

const c = 1.41421

func (n *TreeNode) computeUCB1(deltaN int) {
	var x float64
	if n.HasParent() {
		N := n.ParentPlayouts() + deltaN
		x = c * math.Sqrt(math.Log(float64(N))/float64(n.Playouts()))
	}
	if n.Playouts() > 0 {
		x += float64(n.value / Value(n.Playouts()))
	}
	n.ucb1 = x
}

// Backprop propagates the value v representing n playouts to parents of this node.
// Fixes the frontier heap.
func (n *TreeNode) Backprop(v Value, playouts int) {
	n.value += v
	n.playouts += playouts
	n.computeUCB1(playouts)
	if n.idx != -1 {
		heap.Fix(n.t, n.idx)
	}
	for p := n.parent; p != nil; p = p.parent {
		p.value += v
		p.playouts += playouts
		p.computeUCB1(playouts)
		if n.idx != -1 {
			heap.Fix(p.t, p.idx)
		}
	}
}

type byPlayouts []*TreeNode

func (a byPlayouts) Len() int           { return len(a) }
func (a byPlayouts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPlayouts) Less(i, j int) bool { return a[i].Playouts() > a[j].Playouts() }
