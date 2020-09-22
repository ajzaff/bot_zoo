package zoo

import (
	"sort"
)

// ExtStep contains a step and associated value.
type ExtStep struct {
	Step
	Value
}

func makeExtStep(step Step) ExtStep {
	return ExtStep{
		Step:  step,
		Value: -Inf,
	}
}

// StepList implements an efficient data structure for storing scored steps from search lines.
type StepList struct {
	steps []ExtStep
	begin int // begin index for Sort
}

func (l *StepList) Len() int           { return len(l.steps) - l.begin }
func (a *StepList) Less(i, j int) bool { return a.steps[a.begin+i].Value > a.steps[a.begin+j].Value }
func (a *StepList) Swap(i, j int) {
	a.steps[a.begin+i], a.steps[a.begin+j] = a.steps[a.begin+j], a.steps[a.begin+i]
}

// Generate the steps and scores for position p and append the sorted steps to the move list.
func (l *StepList) Generate(p *Pos) {
	begin := l.Len()
	p.generateSteps(&l.steps)
	l.Sort(begin)
}

// Sorts all steps by value at begin to l.Len().
func (l *StepList) Sort(begin int) {
	l.begin = begin
	sort.Stable(l)
	l.begin = 0
}

// Truncate truncates the list to the given length n.
func (l *StepList) Truncate(n int) {
	l.steps = l.steps[:n]
}

func (l *StepList) At(i int) ExtStep {
	return l.steps[i]
}

func (l *StepList) SetValue(i int, v Value) {
	l.steps[i].Value = v
}

// generateSteps appends all legal steps to a.
// a legal step is any sliding step, push or
// pull step in which the Src, Dest, or Alt is occupied
// and the Src piece is not frozen, and all captures
// are completed.
func (p *Pos) generateSteps(a *[]ExtStep) {
	if p.stepsLeft == 0 {
		return
	}
	if p.moveNum == 1 {
		p.generateSetupSteps(a)
		return
	}
	ourSide := p.Side()
	ourRabbit := GRabbit.WithColor(ourSide)
	empty := p.Empty()
	occupied := ^empty
	for b := occupied; b > 0; b &= b - 1 {
		src := b.Square()
		t := p.At(src)

		var db Bitboard
		if t == ourRabbit {
			db = src.ForwardNeighbors(ourSide)
		} else {
			db = src.Neighbors()
		}

		// Generate step from src to dest.
		emptyDB := db & empty
		for b2 := emptyDB; b2 > 0; b2 &= b2 - 1 {
			dest := b2.Square()
			*a = append(*a, makeExtStep(MakeStep(t, src, dest)))
		}
	}
}

func (p *Pos) generateSetupSteps(a *[]ExtStep) {
	c := p.Side()
	i := A1
	for ; i <= H2 && (c == Gold && p.At(i) != Empty || c == Silver && p.At(i.Flip()) != Empty); i++ {
	}
	if i > H2 {
		return
	}
	if c == Silver {
		i = i.Flip()
	}
	for t := GRabbit.WithColor(c); t <= GElephant.WithColor(c); t++ {
		*a = append(*a, makeExtStep(MakeSetup(Piece(t).WithColor(c), i)))
	}
}
