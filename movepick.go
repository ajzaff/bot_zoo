package zoo

import (
	"fmt"
	"math/rand"
	"sort"
)

type ScoredMove struct {
	score int
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
		scoredMoves[i].score += int(f * r.NormFloat64())
	}
}

func sortMoves(a []ScoredMove) {
	sort.Stable(byLen(a))
	sort.Stable(byScore(a))
}

// sortMoves computes move scores based on naive eval.
// call rescorePVMoves to update the score after finding a PV.
func (e *Engine) scoreMoves(p *Pos, moves [][]Step) []ScoredMove {
	a := make([]ScoredMove, 0, len(moves))

	// 2: Range over all moves.
	for _, move := range moves {

		// 2a. Try the move.
		// In the unlikely case it's illegal give a score of -inf.
		// We want to prune this move from the results.
		if err := p.Move(move); err != nil {
			if err != errRecurringPosition {
				panic(fmt.Sprintf("moveorder_move: %v", err))
			}
		} else {
			// Step 2b: Get the evaluation and set the score.
			// Negate the score since sides have changed.
			a = append(a, ScoredMove{
				score: -p.Score(),
				move:  move,
			})
		}

		// Step 2c. Undo the move.
		if err := p.Unmove(); err != nil {
			panic(fmt.Sprintf("moveorder_unmove: %v", err))
		}
	}
	return a
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
			scoredMoves[i].score = +inf
			continue
		}

		scoredMoves[i].score = -inf
	}
}

// stepSelector handles scoring and sorting steps and provides
// Select for getting the next best step that meets the conditions.
type stepSelector struct {
	e      *Engine
	steps  []Step
	scores []int
}

func (e *Engine) stepSelector(steps []Step) *stepSelector {
	s := &stepSelector{
		e:      e,
		steps:  steps,
		scores: make([]int, len(steps)),
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
	// Check the table to get the PV step if any.
	var pv Step
	if s.e.useTable {
		entry, ok := s.e.table.ProbeDepth(s.e.Pos().zhash, 0)
		if ok && entry.Bound == ExactBound && entry.Step != nil {
			pv = *entry.Step
		}
	}
	// TODO(ajzaff): Estimate value of captures.
	// Make this usable in
	for i, step := range s.steps {
		if step == pv {
			s.scores[i] = inf
		}
		s.scores[i] = -inf
	}
}

// Select selects the next best move.
func (s *stepSelector) SelectScore() (score int, step Step, ok bool) {
	if len(s.steps) == 0 {
		return -inf, invalidStep, false
	}
	step = s.steps[0]
	score = s.scores[0]
	s.steps = s.steps[1:]
	s.scores = s.scores[1:]
	return score, step, true
}

// Select selects the next best move.
func (s *stepSelector) Select() (Step, bool) {
	if len(s.steps) == 0 {
		return invalidStep, false
	}
	step := s.steps[0]
	s.steps = s.steps[1:]
	return step, true
}

// SelectCapture selects the next best capture move.
func (s *stepSelector) SelectCapture() (Step, bool) {
	for len(s.steps) > 0 {
		step := s.steps[0]
		s.steps = s.steps[1:]
		if step.Cap.Valid() {
			return step, true
		}
	}
	return invalidStep, false
}
