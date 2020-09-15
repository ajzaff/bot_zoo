package zoo

import (
	"sort"
	"sync"
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
func (p *Pos) capture(presence Bitboard, src, dest Square) Piece {
	srcB := src.Bitboard()
	destB := dest.Bitboard()
	newPresence := presence ^ (srcB | destB)
	if b := unguardedB(destB&Traps, newPresence); b != 0 {
		// Capture of piece pushed from srcB to destB.
		return p.At(src)
	}
	if b := trappedB(newPresence&srcB.Neighbors(), newPresence); b != 0 {
		// Capture of piece left next to src.
		// TODO(ajzaff): It would be more clear to have a adjacentTrap helper for this.
		return p.At(b.Square())
	}
	return 0xf
}

// generateSteps appends all legal steps to a.
// a legal step is any sliding step, push or
// pull step in which the Src, Dest, or Alt is occupied
// and the Src piece is not frozen, and all captures
// are completed.
func (p *Pos) generateSteps(a *[]Step) {
	if p.stepsLeft < 4 {
		*a = append(*a, Pass)
	}
	if p.stepsLeft == 0 {
		return
	}
	c := p.Side()
	presence := p.Presence(c)
	enemyPresence := p.Presence(c.Opposite())
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

		// Generate default step from src to dest with possible capture.
		for b2 := emptyDB; b2 > 0; b2 &= b2 - 1 {
			dest := b2.Square()
			*a = append(*a, MakeDefaultCapture(src, dest, t, p.capture(presence, src, dest)))
		}
		// Pushing and pulling is not possible.
		if p.stepsLeft < 2 || t.SameType(GRabbit) {
			continue
		}
		// Generate pushes from src to dest (to alt) with possible capture.
		for b2 := db & enemyPresence & p.Weaker(t); b2 > 0; b2 &= b2 - 1 {
			dest := b2.Square()
			for ab := stepsB[dest] & ^sb & empty; ab > 0; ab &= ab - 1 {
				alt := ab.Square()
				if cap := p.capture(presence, src, dest); cap.Valid() {
					*a = append(*a, MakeAlternateCapture(src, dest, alt, t, p.At(dest), cap))
				} else if cap := p.capture(enemyPresence, dest, alt); cap.Valid() {
					*a = append(*a, MakeAlternateCapture(src, dest, alt, t, p.At(dest), cap))
				} else {
					*a = append(*a, MakeAlternate(src, dest, alt, t, p.At(dest)))
				}
			}
		}
		// Generate pulls from alt to src (to dest) with possible capture.
		for ab := stepsB[src] & enemyPresence & p.Weaker(t); ab > 0; ab &= ab - 1 {
			for b2 := emptyDB; b2 > 0; b2 &= b2 - 1 {
				dest := b2.Square()
				alt := ab.Square()
				if cap := p.capture(presence, src, dest); cap.Valid() {
					*a = append(*a, MakeAlternateCapture(src, dest, alt, t, p.At(alt), cap))
				} else if cap := p.capture(enemyPresence, alt, src); cap.Valid() {
					*a = append(*a, MakeAlternateCapture(src, dest, alt, t, p.At(alt), cap))
				} else {
					*a = append(*a, MakeAlternate(src, dest, alt, t, p.At(alt)))
				}
			}
		}
	}
}

func recurring2(ka, kb StepKind, a, b Step) bool {
	// Three cases:
	// 	a and b are default steps that cancel eachother out
	//	a and b are push & pull
	//	a and b are pull & push
	return (ka == KindDefault && kb == KindDefault ||
		ka == KindPush && kb == KindPull ||
		ka == KindPull && kb == KindPush) &&
		a.Src() == b.Dest() && a.Dest() == b.Src()
}

func recurring4(a, b, c, d Step) bool {
	ka, kb, kc, kd := a.Kind(), b.Kind(), c.Kind(), d.Kind()
	return recurring2(ka, kb, a, b) || recurring2(kb, kc, b, c) ||
		recurring2(ka, kc, a, c) || recurring2(kb, kd, b, d) ||
		recurring2(ka, kd, a, d) || recurring2(kb, kc, b, c)
}

// Recurring evaluates statically whether the move leads to a recurring position
// Or whether a move contains intermediate redundancies like a unnecessary pivot
// or push pull pairs. The provided move should be terminated with a (pass) move.
func Recurring(move []Step) bool {
	switch len(move) {
	case 1: // (pass)
		return true
	case 3: // default (pass)
		return recurring2(move[0].Kind(), move[1].Kind(), move[0], move[1])
	case 5: // step step (pass)
		return recurring4(move[0], move[1], move[2], move[3])
	default: // 1, 3, other lengths
		return false
	}
}
