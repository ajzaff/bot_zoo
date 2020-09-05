package zoo

import (
	"fmt"
	"sync/atomic"
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

// Go starts the search routine.
func (e *Engine) Go() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		go e.startSearch()
	}
}

// GoPonder starts the search routine in ponder mode.
func (e *Engine) GoPonder() {
}

// GoInfinite starts an infinite routine.
func (e *Engine) GoInfinite() {
}

// Best returns the current best move.
func (e *Engine) Best() (best SearchResult, ok bool) {
	if atomic.LoadInt32(&e.running) == 0 {
		return SearchResult{}, false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.best, true
}

func (e *Engine) Stop() {
	if atomic.LoadInt32(&e.running) == 1 {
		atomic.SwapInt32(&e.stopping, 1)
	}
}

// Stop search immediately and print the best move info.
func (e *Engine) stop() {
	best, ok := e.Best()
	if !ok {
		fmt.Printf("log search terminated before stop\n")
		return
	}
	fmt.Printf("bestmove %s\n", MoveString(best.Move))
	fmt.Printf("info score %d\n", best.Score)
	fmt.Printf("info nodes %d\n", best.Nodes)
	fmt.Printf("info pv %s\n", MoveString(best.PV))
	fmt.Printf("info time %d\n", int(best.Time.Seconds()))
	fmt.Printf("info curmovenumber %d\n", e.Pos().moveNum)
	atomic.SwapInt32(&e.running, 0)
}

func (e *Engine) startSearch() {
	atomic.SwapInt32(&e.stopping, 0)
	atomic.SwapInt32(&e.running, 1)
	defer func() { e.stop() }()

	p := e.Pos()
	if p.moveNum == 1 {
		// TODO(ajzaff): Find best setup moves using a specialized search.
		// For now, choose a random setup.
		e.mu.Lock()
		e.best = SearchResult{
			Move: e.RandomSetup(),
		}
		e.mu.Unlock()
		return
	}
	start := time.Now()
	e.TimeInfo.Start[p.side] = start
	for d := 1; atomic.LoadInt32(&e.stopping) == 0 && atomic.LoadInt32(&e.running) == 1; d++ {
		best := e.searchRoot(p, d)
		e.mu.Lock()
		e.best = best
		e.mu.Unlock()
	}
	e.table.Clear()
	e.best.Time = time.Now().Sub(start)
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
		if atomic.LoadInt32(&e.stopping) != 0 {
			break
		}
		func() {
			err := p.Move(entry.move)
			defer func() {
				if err := p.Unmove(); err != nil {
					panic(fmt.Sprintf("search_unmove_root: %v", err))
				}
			}()
			if err != nil {
				if err != errRecurringPosition {
					panic(fmt.Sprintf("search_move_root: %v", err))
				}
				return
			}
			score := -e.search(p, -inf, inf, MoveLen(entry.move), depth)
			if score > best.Score {
				best.Score = score
				best.Move = entry.move
				fmt.Printf("log depth %d\n", depth)
				fmt.Printf("log score %d\n", score)
				fmt.Printf("log pv %s\n", entry.move)
				fmt.Printf("log transpositions %d\n", e.table.Len())
			}
		}()
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
	best := -inf
	for _, step := range p.Steps() {
		initSide := p.side

		if atomic.LoadInt32(&e.stopping) != 0 {
			break
		}

		if err := p.Step(step); err != nil {
			panic(fmt.Sprintf("search_step: %v", err))
		}
		var score int
		if p.side == initSide {
			score = e.search(p, alpha, beta, depth+step.Len(), maxDepth)
		} else {
			score = -e.search(p, -beta, -alpha, depth+step.Len(), maxDepth)
		}
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("search_unstep: %v", err))
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
