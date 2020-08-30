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
	var res SearchResult
	for d := 1; d <= depth; d++ {
		res = e.searchRoot(p, d)
	}
	return res
}

func (e *Engine) searchRoot(p *Pos, depth int) SearchResult {
	start := time.Now()
	c := 1
	if depth <= 1 {
		c = -1
	}
	best := SearchResult{
		Depth: depth,
		Score: -inf,
	}
	for _, move := range e.GetMoves(p) {
		t, mseq, err := p.Move(move, false)
		if err != nil {
			continue
		}
		var r SearchResult
		e.search(t, &r, -c, -inf, inf, depth-1)
		r.Score = -r.Score
		best.Nodes += r.Nodes
		if r.Score > best.Score {
			fmt.Printf("log new best move [depth=%d] [%d] %s\n", depth, r.Score, MoveString(mseq))
			best.Score = r.Score
			pv := make([]Step, len(mseq)+len(r.PV))
			copy(pv, mseq)
			copy(pv[len(mseq):], r.PV)
			best.PV = pv
			best.Move = mseq
		}
	}
	best.Time = time.Now().Sub(start)
	return best
}

func (e *Engine) search(p *Pos, res *SearchResult, c, alpha, beta, depth int) {
	if depth <= 0 || p.Terminal() {
		// TODO(ajzaff): Add quiescence search.
		// Among other things, check statically whether any pieces
		// can be flipped into a trap on the next turn.
		res.Score = c * p.Score()
		res.Nodes++
		return
	}
	var bestPV []Step
	for _, move := range e.GetMoves(p) {
		t, _, err := p.Move(move, false)
		if err != nil {
			continue
		}
		var r SearchResult
		pv := make([]Step, len(move)+len(res.PV))
		copy(pv, res.PV)
		copy(pv[len(res.PV):], move)
		r.PV = pv
		e.search(t, &r, -c, -beta, -alpha, depth-1)
		r.Score = -r.Score
		res.Nodes += r.Nodes
		if r.Score >= beta {
			res.Score = beta // fail-hard cutoff
			return
		}
		if r.Score > alpha {
			alpha = r.Score
			bestPV = r.PV
		}
	}
	res.Score = alpha
	res.PV = bestPV
}
