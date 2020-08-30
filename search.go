package zoo

import (
	"fmt"
	"time"
)

const inf = 1000000

type SearchResult struct {
	ZHash int64
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
	e.table.Clear() // To clear or not to clear?
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
	moves := e.getMovesLen(p, n)
	sortedMoves := e.sortMoves(p, moves)
	for _, entry := range sortedMoves {
		t, mseq, err := p.Move(entry.move, false)
		if err != nil {
			continue
		}
		score := -e.search(t, -inf, inf, CountSteps(entry.move), depth)
		if score > best.Score {
			best.Score = score
			best.Move = mseq
			fmt.Printf("log depth %d\n", depth)
			fmt.Printf("log score %d\n", score)
			fmt.Printf("log pv %s\n", MoveString(mseq))
			fmt.Printf("log transpositions %d\n", e.table.Len())
		}
	}
	return best
}

func assert(message string, cond bool) {
	if !cond {
		panic(message)
	}
}

func (e *Engine) search(p *Pos, alpha, beta, depth, maxDepth int) int {
	alphaOrig := alpha

	// Step 1: Check the transposition table.
	if e.useTable {
		if entry, ok := e.table.ProbeDepth(p.ZHash, maxDepth); ok {
			switch entry.Bound {
			case ExactBound:
				return entry.Value
			case LowerBound:
				if alpha < entry.Value {
					alpha = entry.Value
				}
			case UpperBound:
				if beta > entry.Value {
					beta = entry.Value
				}
			}
			if alpha >= beta {
				return entry.Value
			}
		}
	}

	// Step 2: Is this a terminal node or depth==0?
	if depth == maxDepth || p.Terminal() {
		// TODO(ajzaff): Add quiescence search.
		// Among other things, check statically whether any pieces
		// can be flipped into a trap on the next turn.
		return p.Score()
	}

	// Step 2a: Assertions.
	assert("!(0 < depth && depth < maxDepth)", 0 < depth && depth < maxDepth)

	// Step 3: Main search.
	n := maxDepth - depth
	if n > 4 {
		n = 4
	}
	var best int
	moves := e.getMovesLen(p, n)
	sortedMoves := e.sortMoves(p, moves)
	for _, entry := range sortedMoves {
		t, _, err := p.Move(entry.move, false)
		if err != nil {
			continue
		}
		score := -e.search(t, -beta, -alpha, depth+CountSteps(entry.move), maxDepth)
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

	// Step 4: Store transposition table entry.
	if e.useTable {
		entry := &Entry{
			ZHash: p.ZHash,
			Depth: maxDepth,
			Value: best,
		}
		switch {
		case best <= alphaOrig:
			entry.Bound = UpperBound
		case best >= beta:
			entry.Bound = LowerBound
		default:
			entry.Bound = ExactBound
		}
		e.table.Store(entry)
	}

	// Return best score.
	return best
}
