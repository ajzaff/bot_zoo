package zoo

import (
	"math/rand"
	"sort"
)

type ScoredMove struct {
	score Value
	move  []Step
}

type byLen []ScoredMove

func (a byLen) Len() int           { return len(a) }
func (a byLen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLen) Less(i, j int) bool { return len(a[i].move) > len(a[j].move) }

type byScore []ScoredMove

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].score > a[j].score }

// perturbMoves adds noise to the scoredMoves in [-f, +f].
// Intended to be used for parallel Lazy-SMP search.
func (e *Engine) perturbMoves(r *rand.Rand, f float64, scoredMoves []ScoredMove) {
	for i := range scoredMoves {
		scoredMoves[i].score += Value(f * r.NormFloat64())
	}
}

func sortMoves(a []ScoredMove) {
	sort.Stable(byLen(a))
	sort.Stable(byScore(a))
}

// rescorePVMoves updates the scored moves for the PV
// to +inf and sets everything else to -inf.
func (e *Engine) rescorePVMoves(p *Pos, scoredMoves []ScoredMove) {

	// 1: Check the table to get the PV move if any.
	if !e.useTable {
		return
	}
	var best []Step
	v, _, _ := e.table.Best(p)
	if len(v) == 0 {
		return
	}
	best = v

	// 2: Range over all moves.
	for i := range scoredMoves {
		scoredMove := scoredMoves[i]

		// 2a: Update the best move from the table to +inf.
		if MoveEqual(scoredMove.move, best) {
			scoredMoves[i].score = +Inf
			continue
		}

		scoredMoves[i].score = -Inf
	}
}

// stepSelector handles scoring and sorting steps and provides
// Select for getting the next best step that meets the conditions.
type stepSelector struct {
	e      *Engine
	steps  []Step
	scores []Value
}

func (e *Engine) stepSelector(steps []Step) *stepSelector {
	s := &stepSelector{
		e:      e,
		steps:  steps,
		scores: make([]Value, len(steps)),
	}
	s.score()
	return s
}

func (a stepSelector) Len() int           { return len(a.steps) }
func (a stepSelector) Less(i, j int) bool { return a.scores[i] > a.scores[j] }
func (a stepSelector) Swap(i, j int) {
	a.steps[i], a.steps[j] = a.steps[j], a.steps[i]
	a.scores[i], a.scores[j] = a.scores[j], a.scores[i]
}

func (s *stepSelector) score() {
}

// Select selects the next best move.
func (s *stepSelector) SelectScore() (score Value, step Step, ok bool) {
	if len(s.steps) == 0 {
		return -Inf, invalidStep, false
	}
	step = s.steps[0]
	score = s.scores[0]
	s.steps = s.steps[1:]
	s.scores = s.scores[1:]
	return score, step, true
}

// Select selects the next best move.
func (s *stepSelector) Select() (Step, bool) {
	_, step, ok := s.SelectScore()
	return step, ok
}

// SelectCapture selects the next best capture move.
func (s *stepSelector) SelectCapture() (Step, bool) {
	for {
		step, ok := s.Select()
		if !ok {
			break
		}
		if step.Cap.Valid() {
			return step, true
		}
	}
	return invalidStep, false
}
