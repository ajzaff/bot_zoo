package zoo

import (
	"fmt"
	"time"
)

const inf = 1000000

type SearchResult struct {
	Depth int
	Score int
	Nodes int
	Move  []Step
	PV    []Step
	Time  time.Duration
}

func (e *Engine) SearchFixedDepth(depth int) SearchResult {
	p := e.Pos()
	if p.MoveNum == 1 {
		// TODO(ajzaff): Find best setup moves using a specialized search.
		// For now, choose a random setup.
		res := SearchResult{
			Move: e.RandomSetup(),
		}
		return res
	}
	start := time.Now()
	var best SearchResult
	for d := 1; d <= depth; d++ {
		best = e.searchRoot(p, d)
	}
	e.table.Clear()

	best.Time = time.Now().Sub(start)
	return best
}

func (e *Engine) searchRoot(p *Pos, depth int) SearchResult {
	best := SearchResult{
		Depth: depth,
		Score: -inf,
	}
	n := depth
	if n > 4 {
		n = 4
	}
	for _, move := range e.GetMovesLen(p, n) {
		t, mseq, err := p.Move(move, false)
		if err != nil {
			continue
		}
		score := -e.search(t, -1, inf, -inf, depth-CountSteps(move))
		if score > best.Score {
			best.Score = score
			best.Move = mseq
			fmt.Printf("log depth %d\n", depth)
			fmt.Printf("log score %d\n", score)
			fmt.Printf("log bestmove %s\n", MoveString(mseq))
		}
	}
	return best
}

func (e *Engine) search(p *Pos, c, alpha, beta, depth int) int {
	if depth <= 0 || p.Terminal() {
		// TODO(ajzaff): Add quiescence search.
		// Among other things, check statically whether any pieces
		// can be flipped into a trap on the next turn.
		return -c * p.Score()
	}
	n := depth
	if n > 4 {
		n = 4
	}
	var best int
	for _, move := range e.GetMovesLen(p, n) {
		t, _, err := p.Move(move, false)
		if err != nil {
			continue
		}
		score := -e.search(t, -c, -beta, -alpha, depth-CountSteps(move))
		if score > best {
			best = score
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break // fail-hard cutoff
		}
	}
	return best
}
