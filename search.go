package zoo

import "math"

func (e *Engine) Search() (move []Step, score int) {
	p := e.Pos()
	if p.MoveNum == 1 {
		// TODO(ajzaff): Find best setup moves using a specialized search.
		// For now, choose a random setup.
		return e.RandomSetup(), 0
	}
	move, score = e.searchRoot(p, 2)
	return move, score
}

var inf = 1000000

func (e *Engine) searchRoot(p *Pos, depth int) ([]Step, int) {
	bestScore := math.MinInt64
	c := 1
	if p.Side != Gold {
		c = -1
	}
	var bestMove []Step
	for _, move := range p.GetMoves() {
		t, mseq, err := p.Move(move, false)
		if err != nil {
			continue
		}
		score := -e.search(t, -c, -inf, inf, depth-1)
		if len(bestMove) == 0 || score > bestScore {
			bestScore = score
			bestMove = mseq
		}
	}
	return bestMove, bestScore
}

func (e *Engine) search(p *Pos, c, alpha, beta, depth int) int {
	if depth == 0 || p.Terminal() {
		// TODO(ajzaff): Add quiescence search.
		// Among other things, check statically whether any pieces
		// can be flipped into a trap on the next turn.
		return c * p.Score()
	}
	for _, move := range p.GetMoves() {
		t, _, err := p.Move(move, false)
		if err != nil {
			continue
		}
		v := -e.search(t, -c, -beta, -alpha, depth-1)
		if v >= beta {
			return beta // fail-hard cutoff
		}
		if v > alpha {
			alpha = v
		}
	}
	return alpha
}
