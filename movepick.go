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

// stepSelector handles scoring and sorting steps and provides
// Select for getting the next best step that meets the conditions.
type stepSelector struct {
	side   Color
	steps  []Step
	scores []Value
}

func newStepSelector(c Color, steps []Step) *stepSelector {
	s := &stepSelector{
		side:   c,
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

var goalRange = [2]Bitboard{
	// Gold
	^NotRank8 | ^NotRank8>>8 | ^NotRank8>>16,
	// Silver
	^NotRank1 | ^NotRank1<<8 | ^NotRank1<<16,
}

func (s *stepSelector) score() {
	for i, step := range s.steps {
		switch {
		case step.Piece1.SameType(GRabbit) && step.Dest.Bitboard()&goalRange[step.Piece1.Color()] != 0:
			// Coarse goal threats.
			s.scores[i] = 5000
		case step.Capture(): // Self captures to be avoided.
			t := step.Cap.Piece
			if t.Color() == s.side {
				s.scores[i] = -pieceValue[t&decolorMask]
			} else {
				s.scores[i] = pieceValue[t&decolorMask]
			}
		default:
		}
	}
	sort.Stable(s)
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
