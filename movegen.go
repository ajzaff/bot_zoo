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
	c := p.Side()
	presence := p.Presence(c)
	empty := p.Empty()
	for b := presence; b > 0; b &= b - 1 {
		sb := b & -b
		src := b.Square()
		t := p.board[src]

		if p.frozenB(t, sb) {
			continue
		}
		var db Bitboard
		if t.SameType(GRabbit) {
			db = src.ForwardNeighbors(c)
		} else {
			db = src.Neighbors()
		}
		emptyDB := db & empty

		// Generate step from src to dest.
		for b2 := emptyDB; b2 > 0; b2 &= b2 - 1 {
			dest := b2.Square()
			*a = append(*a, makeExtStep(MakeStep(t, src, dest)))
		}
	}
}

func (p *Pos) generateSetupSteps(a *[]ExtStep) {
	setupPieces := map[Piece]int{
		GRabbit:   8,
		GCat:      2,
		GDog:      2,
		GHorse:    2,
		GCamel:    1,
		GElephant: 1,
	}
	i, end := A1, H2
	c := p.Side()
	if c == Silver {
		i = A8
		end = H7
	}
	for i != end {
		t := p.At(i)
		if t == Empty {
			break
		}
		setupPieces[t.RemoveColor()]--
		if i.File() == 7 && c == Silver {
			i -= 15
		} else {
			i++
		}
	}
	var piecesInt []int
	for t := range setupPieces {
		if t != Empty && setupPieces[t] > 0 {
			piecesInt = append(piecesInt, int(t))
		}
	}
	sort.Ints(piecesInt)
	for _, t := range piecesInt {
		*a = append(*a, makeExtStep(MakeSetup(Piece(t).WithColor(c), i)))
	}
}
