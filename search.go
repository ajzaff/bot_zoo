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
	var best SearchResult
	for d := 1; d <= depth; d++ {
		best = e.searchRoot(p, depth)
	}
	return best
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
	alpha, beta := -inf, inf
	for _, move := range e.GetMoves(p) {
		t, mseq, err := p.Move(move, false)
		if err != nil {
			continue
		}
		var r SearchResult
		if v, ok := e.table.GetDepth(t.ZHash, depth); ok {
			r = *v
		} else {
			e.search(t, &r, -c, -beta, -alpha, depth-1)
			r.Score = -r.Score
		}
		best.Nodes += r.Nodes
		if r.Score > best.Score {
			pv := make([]Step, len(mseq)+len(r.PV))
			copy(pv, mseq)
			copy(pv[len(mseq):], r.PV)
			best.Score = r.Score
			best.PV = pv
			best.Move = mseq
			fmt.Printf("log depth %d\n", depth)
			fmt.Printf("log score %d\n", r.Score)
			fmt.Printf("log pv %s\n", MoveString(pv))
		}
		if r.Score > alpha {
			alpha = r.Score
		}
		if alpha >= beta {
			break
		}
	}
	best.Time = time.Now().Sub(start)
	return best
}

func (e *Engine) search(p *Pos, best *SearchResult, c, alpha, beta, depth int) {
	if depth <= 0 || p.Terminal() {
		// TODO(ajzaff): Add quiescence search.
		// Among other things, check statically whether any pieces
		// can be flipped into a trap on the next turn.
		best.Score = c * p.Score()
		best.Nodes++
		return
	}
	for _, move := range e.GetMoves(p) {
		t, _, err := p.Move(move, false)
		if err != nil {
			continue
		}
		var r SearchResult
		if v, ok := e.table.GetDepth(t.ZHash, depth); ok {
			r = *v
		} else {
			pv := make([]Step, len(move)+len(best.PV))
			copy(pv, best.PV)
			copy(pv[len(best.PV):], move)
			r.PV = pv
			e.search(t, &r, -c, -beta, -alpha, depth-1)
			r.Score = -r.Score
		}
		best.Nodes += r.Nodes
		if r.Score > best.Score {
			best.Score = r.Score
			best.PV = r.PV
			e.table.Update(t.ZHash, &r)
		}
		if r.Score > alpha {
			alpha = r.Score
		}
		if alpha >= beta {
			break // fail-hard cutoff
		}
	}
}
