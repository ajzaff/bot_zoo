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
func (a byLen) Less(i, j int) bool { return len(a[i].move) < len(a[j].move) }

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

func scoreStep(p *Pos, step Step) Value {
	side := step.Piece1.Color()
	var value Value
	// Add +Inf - |step| for goal moves.
	if step.Piece1.SameType(GRabbit) && step.Dest.Bitboard()&goalRanks[step.Piece1.Color()] != 0 {
		value += +Inf - Value(step.Len()) // find shortest mate
	}
	// Add O(large) for rabbit moves close to goal.
	if step.Piece1.SameType(GRabbit) && step.Dest.Bitboard()&goalRange[step.Piece1.Color()] != 0 { // Coarse goal threat:
		value += 2000
	}
	// Add static value of capture:
	if step.Capture() {
		t := step.Cap.Piece
		if t.Color() == side {
			value -= pieceValue[t&decolorMask]
		} else {
			value += pieceValue[t&decolorMask]
		}
	}
	return value
}

// scoreMoves is generally called at the search root and provides a better initial
// order before the PV order takes over. These relative orders will be maintained
// later due to stable sort so it's very important to have a goo order heuristic.
func (e *Engine) scoreMoves(p *Pos, moves [][]Step) []ScoredMove {
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
			scoredMoves[i].score += scoreStep(p, step)
		}
	}
	sortMoves(scoredMoves)
	return scoredMoves
}
