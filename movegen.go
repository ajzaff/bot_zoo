package zoo

import (
	"sort"
)

var (
	stepsB       [64]Bitboard
	rabbitStepsB [2][64]Bitboard
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

func init() {
	for i := Square(0); i < 64; i++ {
		b := i.Bitboard()
		steps := b.Neighbors()
		stepsB[i] = steps
		grSteps := steps
		if b&NotRank1 != 0 { // rabbits can't move backwards.
			grSteps ^= b >> 8
		}
		srSteps := steps
		if b&NotRank8 != 0 {
			srSteps ^= b << 8
		}
		rabbitStepsB[Gold][i] = grSteps
		rabbitStepsB[Silver][i] = srSteps
	}
}

func unguardedB(b, presence Bitboard) Bitboard {
	return b & ^presence.Neighbors()
}

func trappedB(b, presence Bitboard) Bitboard {
	return b & unguardedB(Traps, presence)
}

// capture computes statically the capture resulting from a move from src to dest if any.
// The possible types of captures are abandoning a trapped piece or
// capturing ourselves by stepping onto an unguarded trap square.
func (p *Pos) capture(presence Bitboard, src, dest Square) Square {
	srcB := src.Bitboard()
	destB := dest.Bitboard()
	newPresence := presence ^ (srcB | destB)
	if b := unguardedB(destB&Traps, newPresence); b != 0 {
		// Capture of piece pushed from srcB to destB.
		return src
	}
	if b := trappedB(newPresence&srcB.Neighbors(), newPresence); b != 0 {
		// Capture of piece left next to src.
		// TODO(ajzaff): It would be more clear to have a adjacentTrap helper for this.
		return b.Square()
	}
	return 64
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
			db = rabbitStepsB[c][src]
		} else {
			db = stepsB[src]
		}
		emptyDB := db & empty

		// Generate step from src to dest.
		for b2 := emptyDB; b2 > 0; b2 &= b2 - 1 {
			dest := b2.Square()
			*a = append(*a, makeExtStep(MakeCapture(t, p.capture(presence, src, dest))))
		}
	}
}
