package zoo

import (
	"math/rand"
	"sort"
	"time"
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

var goalRanks = [2]Bitboard{
	^NotRank8, // Gold
	^NotRank1, // Silver
}

var goalRange = [2]Bitboard{
	^NotRank8 | ^NotRank8>>8 | ^NotRank8>>16, // Gold
	^NotRank1 | ^NotRank1<<8 | ^NotRank1<<16, // Silver
}

func scoreStep(step Step) Value {
	switch side := step.Piece1.Color(); {
	case step.Piece1.SameType(GRabbit) && step.Dest.Bitboard()&goalRanks[step.Piece1.Color()] != 0: // Goal:
		return +Inf
	case step.Piece1.SameType(GRabbit) && step.Dest.Bitboard()&goalRange[step.Piece1.Color()] != 0: // Coarse goal threat:
		return 1000
	case step.Capture(): // Self captures to be avoided, captures good:
		t := step.Cap.Piece
		if t.Color() == side {
			return -pieceValue[t&decolorMask]
		}
		return pieceValue[t&decolorMask]
	default:
		return 0
	}
}

// stepSelector handles scoring and sorting steps and provides
// Select for getting the next best step that meets the conditions.
type stepSelector struct {
	steps  []Step
	scores []Value
}

func newStepSelector(steps []Step) *stepSelector {
	s := &stepSelector{
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
	for i, step := range s.steps {
		s.scores[i] = scoreStep(step)
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

// scoreMoves is generally called at the search root and provides a better initial
// order before the PV order takes over. These relative orders will be maintained
// later due to stable sort so it's very important to have a goo order heuristic.
func (e *Engine) scoreMoves(moves [][]Step) []ScoredMove {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(moves), func(i, j int) {
		moves[i], moves[j] = moves[j], moves[i]
	})
	scoredMoves := make([]ScoredMove, len(moves))
	for i, move := range moves {
		// Add root order noise, if configured.
		if e.rootOrderNoise != 0 {
			scoredMoves[i] = ScoredMove{score: Value(float64(e.rootOrderNoise) * r.Float64()), move: move}
		}
		// Add individual step values.
		for _, step := range move {
			scoredMoves[i].score += scoreStep(step)
		}
	}
	sortMoves(scoredMoves)
	return scoredMoves
}
