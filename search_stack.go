package zoo

import (
	"sort"
	"sync"
)

// Stack is a structure maintained for each ply of search.
// It contains all the information relevant to that ply.
// The stack is used for reconstructing the searched PV.
type Stack struct {
	PV    []Step
	Depth uint8
	Step  Step
	Eval  Value
	Nodes int
}

// StepList implements an efficient data structure for storing scored steps from search lines.
type StepList struct {
	steps  []Step
	scores []Value
	p      int // start index for sorting a subset of moves
}

func (l *StepList) Len() int           { return len(l.steps) - l.p }
func (a *StepList) Less(i, j int) bool { return a.scores[a.p+i] > a.scores[a.p+j] }
func (a *StepList) Swap(i, j int) {
	a.steps[a.p+i], a.steps[a.p+j] = a.steps[a.p+j], a.steps[a.p+i]
	a.scores[a.p+i], a.scores[a.p+j] = a.scores[a.p+j], a.scores[a.p+i]
}

var scorePool = sync.Pool{
	New: func() interface{} {
		scores := make([]Value, 0, 64)
		return &scores
	},
}

// Generate the steps and scores for position p and append the sorted steps to the move list.
func (l *StepList) Generate(p *Pos) {
	n := l.Len()
	p.generateSteps(&l.steps)
	if v := l.Len(); v < cap(l.scores) { // Reslice
		l.scores = l.scores[:v]
	} else { // Get from pool
		slice := scorePool.Get().(*[]Value)
		if len(*slice) > v-n || v-n >= cap(*slice) {
			*slice = (*slice)[:v-n]
		} else { // Reallocate
			newSlice := make([]Value, v-n)
			*slice = newSlice
		}
		l.scores = append(l.scores, *slice...)
	}
	for i := n; i < l.Len(); i++ {
		l.scores[i] = scoreStep(p, l.steps[i])
	}
	l.p = n
	sort.Stable(l)
	l.p = 0
}

// Truncate truncates the list to the given length n.
func (l *StepList) Truncate(n int) {
	l.steps = l.steps[:n]
}

func (l *StepList) AtScore(i int) (step Step, score Value) {
	return l.steps[i], l.scores[i]
}

func (l *StepList) At(i int) Step {
	return l.steps[i]
}
