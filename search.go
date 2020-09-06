package zoo

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const inf = 1000000

type SearchResult struct {
	ZHash int64
	Depth int
	Score int
	Move  []Step
	PV    []Step
}

type SearchInfo struct {
	// cumulative nodes at the given ply
	// which can be used to compute the EBF.
	nodes []int

	// ply start times (unix seconds) per ply.
	// Used to determine remaining time to search.
	times []int64

	done chan struct{}

	// best move
	// time (unix seconds) when best was last set.
	// Corresponds to end of search.
	bestTime int64        // guarded by mu
	best     SearchResult // guarded by mu
	mu       sync.Mutex
}

func newSearchInfo() *SearchInfo {
	s := &SearchInfo{
		done: make(chan struct{}),
	}
	s.newPly()
	return s
}

// Best returns the current best move.
func (s *SearchInfo) Best() SearchResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.best
}

// setBest sets the current best move.
func (s *SearchInfo) setBest(best SearchResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.best = best
	s.bestTime = time.Now().Unix()
}

func (s *SearchInfo) newPly() {
	n := 0
	if len(s.nodes) > 0 {
		n = s.nodes[len(s.nodes)-1]
	}
	s.nodes = append(s.nodes, n)
	s.times = append(s.times, time.Now().Unix())
}

func (s *SearchInfo) addNodes(n int) {
	s.nodes[len(s.nodes)-1] += n
}

// Nodes returns the number of nodes searched in all ply.
func (s *SearchInfo) Nodes() int {
	return s.nodes[len(s.nodes)-1]
}

// Seconds returns the number of seconds searched in all ply.
func (s *SearchInfo) Seconds() int64 {
	s.mu.Lock()
	end := s.bestTime
	s.mu.Unlock()
	if end == 0 {
		end = s.times[len(s.times)-1]
	}
	return end - s.times[0]
}

func computebfNd(d int, b float64) float64 {
	n := b
	for i := 2; i <= d; i++ {
		n *= math.Pow(b, float64(i))
	}
	return n
}

// EBF experimentally computes the effective branching factor.
// This helps compute the nodes required to search to depth d
// and by extension the duration of time required to complete
// another ply.
func (s *SearchInfo) EBF() float64 {
	const (
		tol   = 10000.
		small = 1e-4
	)
	d := len(s.nodes) - 1
	N := float64(s.Nodes())
	b := 0.
	for n, lo, hi := 0., 1., 50.; hi-lo > small; {
		mid := (hi-lo)/2 + lo
		b = mid
		n = computebfNd(d, b)
		e := n - N
		if math.Abs(e) < tol {
			break
		}
		if e < 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	return b
}

func (s *SearchInfo) guessNextPlyDuration() time.Duration {
	now := time.Now().Unix()
	d := len(s.nodes)
	v := float64(now-s.times[d-1]) * math.Pow(s.EBF(), float64(d))
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return time.Duration(v * float64(time.Second))
}

// Go starts the search routine.
func (e *Engine) Go() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		go e.iterativeDeepeningRoot()
	}
}

// GoPonder starts the search routine in ponder mode.
func (e *Engine) GoPonder() {
}

// GoInfinite starts an infinite routine.
func (e *Engine) GoInfinite() {
}

// Stop stops the search immediately and blocks until stopped.
func (e *Engine) Stop() {
	if atomic.LoadInt32(&e.running) == 1 {
		e.stopping = 1
		<-e.searchInfo.done
	}
}

// Stop search immediately and print the best move info.
func (e *Engine) stop() {
	best := e.searchInfo.Best()
	fmt.Printf("bestmove %s\n", MoveString(best.Move))
	fmt.Printf("info score %d\n", best.Score)
	fmt.Printf("info nodes %d\n", e.searchInfo.Nodes())
	fmt.Printf("info time %d\n", e.searchInfo.Seconds())
	fmt.Printf("info ebf %f\n", e.searchInfo.EBF())
	fmt.Printf("info pv %s\n", MoveString(best.PV))
	fmt.Printf("info curmovenumber %d\n", e.Pos().moveNum)
	atomic.SwapInt32(&e.running, 0)
	e.searchInfo.done <- struct{}{}
}

func (e *Engine) iterativeDeepeningRoot() {
	e.stopping = 0
	defer func() { e.stop() }()

	e.searchInfo = newSearchInfo()
	if e.timeInfo == nil {
		e.timeInfo = e.timeControl.newTimeInfo()
	} else {
		e.timeControl.resetTurn(e.timeInfo, e.p.side)
	}

	p := e.Pos()
	if p.moveNum == 1 {
		// TODO(ajzaff): Find best setup moves using a specialized search.
		// For now, choose a random setup.
		e.searchInfo.setBest(SearchResult{
			Move: e.RandomSetup(),
		})
		return
	}

	var best SearchResult
	for d := 1; e.stopping == 0 && atomic.LoadInt32(&e.running) == 1; d++ {
		if best.Score >= terminalEval {
			go e.Stop()
			return
		}
		if next, rem := e.searchInfo.guessNextPlyDuration(), e.timeControl.GameTimeRemaining(e.timeInfo, p.side); d > 4 && rem <= next {
			fmt.Printf("log stop deepening before ply=%d (ebf=%f, cost=%s, budget=%s)\n", d, e.searchInfo.EBF(), next, rem)
			go e.Stop()
			return
		}
		e.searchInfo.newPly()
		best = e.searchRoot(p, d)
		e.searchInfo.setBest(best)
	}
	e.table.Clear()
	e.searchInfo.setBest(best)
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
		if e.stopping != 0 {
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
	// Start quiescense search.
	if depth == maxDepth || p.Terminal() {
		return e.quiescence(p, alpha, beta)
	}

	// Step 2a: Assertions.
	assert("!(0 < depth && depth < maxDepth)", 0 < depth && depth < maxDepth)

	steps := p.Steps()

	// Step 3: Main search.
	best := -inf
	nodes := 0
	for _, step := range steps {
		nodes++
		initSide := p.side

		if e.stopping != 0 {
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

	e.searchInfo.addNodes(nodes)

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

// TODO(ajzaff): Measure the effect of counting quienscence nodes on the EBF.
// This has direct consequences on move timings.
func (e *Engine) quiescence(p *Pos, alpha, beta int) int {
	eval := p.Score()
	if eval >= beta {
		return beta
	}
	if alpha < eval {
		alpha = eval
	}

	steps := p.Steps()
	nodes := 0
	for _, step := range steps {
		nodes++
		initSide := p.side

		if !step.Capture() {
			continue
		}

		if err := p.Step(step); err != nil {
			panic(fmt.Sprintf("quiescense_step: %v", err))
		}
		var score int
		if p.side == initSide {
			score = e.quiescence(p, alpha, beta)
		} else {
			score = -e.quiescence(p, -beta, -alpha)
		}
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("quiescense_unstep: %v", err))
		}
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	e.searchInfo.addNodes(nodes)
	return alpha
}
