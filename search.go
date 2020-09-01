package zoo

import (
	"fmt"
	"log"
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
	if p.moveNum == 1 {
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
	moves := e.getRootMovesLen(p, n)
	sortedMoves := e.sortMoves(p, moves)
	for _, entry := range sortedMoves {
		if err := p.Move(entry.move); err != nil {
			if err != errRecurringPosition {
				log.Printf("search: %v", err)
			}
			continue
		}
		score := -e.search(p, -inf, inf, len(entry.move), depth)
		if err := p.Unmove(entry.move); err != nil {
			log.Printf("unmove: %v", err)
		}
		if score > best.Score {
			best.Score = score
			best.Move = entry.move
			fmt.Printf("log depth %d\n", depth)
			fmt.Printf("log score %d\n", score)
			fmt.Printf("log pv %s\n", MoveString(entry.move))
			fmt.Printf("log transpositions %d\n", e.table.Len())
		}
	}
	return best
}

func (e *Engine) search(p *Pos, alpha, beta, depth, maxDepth int) int {
	alphaOrig := alpha

	// Step 1: Check the transposition table.
	if e.useTable {
		if entry, ok := e.table.ProbeDepth(p.zhash, maxDepth); ok {
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
	var best int
	for _, step := range p.Steps() {
		if err := p.Step(step); err != nil {
			log.Println(err)
			continue
		}
		var score int
		if len(p.steps) == 0 {
			p.NullMove()
			score = -e.search(p, -beta, -alpha, depth+1, maxDepth)
		} else {
			score = e.search(p, alpha, beta, depth+1, maxDepth)
		}
		if err := p.Unstep(step); err != nil {
			log.Println(err)
			continue
		}
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
		entry := &TableEntry{
			ZHash: p.zhash,
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
