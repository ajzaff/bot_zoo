package zoo

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const inf = 1000000

type SearchResult struct {
	Score int
	Move  []Step
}

type SearchInfo struct {
	// cumulative nodes at the given ply
	// which can be used to compute the EBF.
	nodes []int // guarded by m

	// ply start times (unix nanoseconds) per ply.
	// Used to determine remaining time to search.
	times []int64

	done chan struct{}

	// best move
	// time (unix nanoseconds) when best was last set.
	// Corresponds to end of search.
	bestTime int64        // guarded by m
	best     SearchResult // guarded by m

	m sync.Mutex
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
	s.m.Lock()
	defer s.m.Unlock()
	return s.best
}

// setBest sets the current best move.
func (s *SearchInfo) setBest(best SearchResult) {
	s.m.Lock()
	defer s.m.Unlock()
	s.best = best
	s.bestTime = time.Now().UnixNano()
}

func (s *SearchInfo) newPly() {
	n := 0
	if len(s.nodes) > 0 {
		n = s.nodes[len(s.nodes)-1]
	}
	s.nodes = append(s.nodes, n)
	s.times = append(s.times, time.Now().UnixNano())
}

func (s *SearchInfo) addNodes(n int) {
	s.m.Lock()
	defer s.m.Unlock()
	s.nodes[len(s.nodes)-1] += n
}

// Nodes returns the number of nodes searched in all ply.
func (s *SearchInfo) Nodes() int {
	return s.nodes[len(s.nodes)-1]
}

// Seconds returns the number of seconds searched in all ply.
func (s *SearchInfo) Seconds() int64 {
	s.m.Lock()
	end := s.bestTime
	s.m.Unlock()
	if end == 0 {
		end = s.times[len(s.times)-1]
	}
	return int64(time.Duration(end-s.times[0]) / time.Second)
}

func computebfNd(d int, b float64) float64 {
	n := b
	for i := 2; i <= d; i++ {
		n += math.Pow(b, float64(i))
	}
	return n
}

func (s *SearchInfo) ebf(d int, N float64) (b float64, err float64) {
	const (
		tol   = 1000.
		small = 1e-4
	)
	var n = 0.
	for lo, hi := 1., 20.; hi-lo > small; {
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
	return b, n - N
}

// EBF experimentally computes the effective branching factor using per-ply node count.
// This helps compute the time required to search to depth d.
func (s *SearchInfo) EBF() (b float64, err float64) {
	d := len(s.nodes) - 1
	if d < 4 {
		// Avoid EBF instability in small values.
		return 1, 0
	}
	return s.ebf(d, float64(s.nodes[d]))
}

func (s *SearchInfo) guessNextPlyDuration() time.Duration {
	d := len(s.times) - 1
	if d < 4 {
		// These searches are basically free.
		// We use len < 5 (instead of < 3) to avoid EBF instability in these small values.
		return 0
	}
	// Compute the ratio of the amount of time spent and nodes and solve for X.
	//  Time[d-1]        X
	// ----------  =  --------
	// Nodes[d-1]     Nodes[d]

	b, _ := s.EBF()
	lastDuration := float64(time.Now().UnixNano() - s.times[d])
	lastNodes := float64(s.nodes[d])
	nextNodes := lastNodes + math.Pow(b, float64(1+d))
	v := lastDuration / lastNodes * nextNodes
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return time.Duration(v)
}

func (e *Engine) startNow() {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("log SEARCH_ERROR recovered: %v\n", r))
		}
	}()
	go e.iterativeDeepeningRoot()
}

// Go starts the search routine in a new goroutine.
func (e *Engine) Go() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		e.ponder = false
		e.startNow()
	}
}

// GoFixed starts a fixed-depth search routine and blocks until it finishes.
func (e *Engine) GoFixed(fixedDepth int) {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		e.ponder = false
		prevDepth := e.fixedDepth
		e.fixedDepth = fixedDepth
		defer func() {
			if r := recover(); r != nil {
				panic(fmt.Sprintf("log SEARCH_ERROR_FIXED recovered: %v\n", r))
			}
		}()
		e.iterativeDeepeningRoot()
		e.fixedDepth = prevDepth
	}
}

// GoPonder starts the ponder search in a new goroutine.
func (e *Engine) GoPonder() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		e.ponder = true
		e.startNow()
	}
}

// GoInfinite starts an infinite routine (same as GoPonder).
func (e *Engine) GoInfinite() {
	e.GoPonder()
}

// Stop stops the search immediately and blocks until stopped.
func (e *Engine) Stop() {
	if atomic.LoadInt32(&e.running) == 1 {
		e.stopping = 1
		<-e.searchInfo.done
		atomic.SwapInt32(&e.running, 0)
	}
}

func (s *SearchInfo) searchRateKNps() int64 {
	ns := s.times[len(s.times)-1] - s.times[0]
	if ns == 0 {
		return 0
	}
	n := s.nodes[len(s.nodes)-1]
	return int64(float64(n) / (float64(ns) / float64(time.Second)) / 1000)
}

func (e *Engine) printSearchInfo() {
	s := e.searchInfo
	if e.ponder {
		fmt.Printf("log ponder\n")
	}
	fmt.Printf("info depth %d\n", len(s.nodes))
	fmt.Printf("info time %d\n", s.Seconds())
	s.m.Lock()
	fmt.Printf("info score %d\n", s.best.Score)
	fmt.Printf("info move %s\n", MoveString(s.best.Move))
	s.m.Unlock()
	if pv, _, _ := e.table.PV(e.p); len(pv) > 0 {
		fmt.Printf("info pv %s\n", MoveString(pv))
	}
	fmt.Printf("info nodes %d\n", s.nodes[len(s.nodes)-1])
	fmt.Printf("log rate %d kN/s\n", s.searchRateKNps())
	fmt.Printf("log transpositions %d\n", e.table.Len())
}

func (e *Engine) stopInternal() {
	e.printSearchInfo()
	if !e.ponder {
		best := e.searchInfo.Best()
		fmt.Printf("bestmove %s\n", MoveString(best.Move))
	}
	e.searchInfo.done <- struct{}{}
}

func initWindowDelta(x int) int {
	v := x
	if v < 0 {
		v = -v
	}
	return 21 + v/256
}

func (e *Engine) iterativeDeepeningRoot() {
	e.stopping = 0
	defer func() { e.stopInternal() }()

	if !e.lastPonder {
		e.table.Clear()
	}
	e.lastPonder = e.ponder
	e.searchInfo = newSearchInfo()
	if !e.ponder {
		if e.timeInfo == nil {
			e.timeInfo = e.timeControl.newTimeInfo()
		} else {
			e.timeControl.resetTurn(e.timeInfo, e.p.side)
		}
	}

	p := e.Pos()
	if p.moveNum == 1 {
		// TODO(ajzaff): Find best setup moves using a specialized search.
		// For now, choose a random setup.
		e.searchInfo.setBest(SearchResult{
			Move: e.RandomSetup(),
		})
		go e.Stop()
		return
	}

	// First ply searched on main goroutine.
	rootSteps := e.scoreMoves(p, e.getRootMovesLen(p, 1))
	sortMoves(rootSteps)
	best := e.searchRoot(p, rootSteps, -inf, +inf, 1)
	e.printSearchInfo()

	const maxPly = 256

	// Main search loop.
	for depth := 2; depth < maxPly; depth++ {
		if !e.ponder && depth > e.minDepth && e.fixedDepth == 0 {
			if next, rem := e.searchInfo.guessNextPlyDuration(), e.timeControl.TurnTimeRemaining(e.timeInfo, p.side); rem <= next {
				go e.Stop()
				b, err := e.searchInfo.EBF()
				fmt.Printf("log stop deepening before ply=%d (ebf=%f(err=%f), cost=%s, budget=%s)\n", depth, b, err, next, rem)
				break
			}
		}
		e.searchInfo.newPly()

		bestChan := make(chan SearchResult)
		var wg sync.WaitGroup
		wg.Add(e.concurrency)
		for i := 0; i < e.concurrency; i++ {
			// Implementation of Lazy SMP which runs parallel searches over
			// the same nodes.
			go func() {
				newPos := p.Clone()
				best := e.searchInfo.Best()

				// Copy all moves for use in this goroutine and add rootOrderNoise if configured.
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				n := depth
				if n > 4 {
					n = 4
				}
				moves := e.getRootMovesLen(newPos, n)
				// Shuffling moves increases variance in concurrent search.
				r.Shuffle(len(moves), func(i, j int) {
					moves[i], moves[j] = moves[j], moves[i]
				})
				scoredMoves := e.scoreMoves(newPos, moves)
				if e.rootOrderNoise > 0 {
					e.perturbMoves(r, e.rootOrderNoise, scoredMoves)
				}

				// Aspiration window tracks the previous score for search.
				// Whenever we fail high or low widen the window.

				// Start with a small aspiration window and, in the case of a fail
				// high/low, re-search with a bigger window until we don't fail
				// high/low anymore.
				delta := initWindowDelta(best.Score)
				alpha, beta := best.Score-delta, best.Score+delta
				if alpha < -inf {
					alpha = -inf
				}
				if beta > inf {
					beta = inf
				}

				for e.stopping == 0 {

					// Rescore the PV move and sort stably.
					e.rescorePVMoves(newPos, scoredMoves)
					sortMoves(scoredMoves)

					if res := e.searchRoot(newPos, scoredMoves, alpha, beta, depth); res.Score <= alpha {
						beta = (alpha + beta) / 2
						alpha = res.Score - delta
						if alpha < -inf {
							alpha = -inf
						}
						fmt.Printf("log failed_LOW %d: retry with new aspiration window: [%d, %d] (delta=%d)\n", res.Score, alpha, beta, delta)
					} else if res.Score >= beta {
						beta = res.Score + delta
						if beta > inf {
							beta = inf
						}
						fmt.Printf("log failed_HIGH %d: retry with new aspiration window: [%d, %d] (delta=%d)\n", res.Score, alpha, beta, delta)
					} else {
						best = res
						break
					}

					// Update aspiration window delta
					delta += delta/4 + 5

					assert("!(alpha >= -inf && beta <= inf)", alpha >= -inf && beta <= inf)
				}
				bestChan <- best
				wg.Done()
			}()
		}
		go func() {
			wg.Wait()
			close(bestChan)
		}()
		for b := range bestChan {
			if b.Score > best.Score {
				best = b
			}
		}
		e.searchInfo.setBest(best)

		if e.stopping == 1 || atomic.LoadInt32(&e.running) == 0 {
			break
		}

		if isTerminal(best.Score) {
			go e.Stop()
			fmt.Printf("log stop search at depth=%d with terminal eval\n", depth)
			break
		}
		if e.fixedDepth > 0 && depth >= e.fixedDepth {
			go e.Stop()
			fmt.Printf("log stop fixed_depth search at depth=%d\n", depth)
			break
		}

		// Print search info for depth d.
		e.printSearchInfo()
	}
	e.searchInfo.setBest(best)
}

func (e *Engine) searchRoot(p *Pos, scoredMoves []ScoredMove, alpha, beta, depth int) SearchResult {
	best := SearchResult{Score: alpha}
	for _, entry := range scoredMoves {
		if e.stopping != 0 {
			break
		}
		n := MoveLen(entry.move)
		if n > depth {
			continue
		}
		if err := p.Move(entry.move); err != nil {
			panic(fmt.Sprintf("search_move_root: %s: %v", entry.move, err))
		}
		score := -e.search(p, -beta, -alpha, n, depth)
		if err := p.Unmove(); err != nil {
			panic(fmt.Sprintf("search_unmove_root: %s: %v", entry.move, err))
		}
		if score > alpha {
			alpha = score
		}
		if score > best.Score {
			best.Score = score
			best.Move = entry.move
		}
	}

	// Store transposition table entry steps.
	if e.useTable {
		for _, step := range best.Move {
			entry := &TableEntry{
				Bound: ExactBound,
				ZHash: p.zhash,
				Depth: depth,
				Value: best.Score,
				Step:  new(Step),
				pv:    true,
			}
			*entry.Step = step
			e.table.Store(entry)
			if err := p.Step(step); err != nil {
				panic(fmt.Sprintf("root_store_step: %s: %s: %v", best.Move, step, err))
			}
		}
		for i := len(best.Move) - 1; i >= 0; i-- {
			if err := p.Unstep(); err != nil {
				panic(fmt.Sprintf("root_store_unstep: %s: %s: %v", best.Move, best.Move[i], err))
			}
		}
	}
	return best
}

func (e *Engine) search(p *Pos, alpha, beta, depth, maxDepth int) int {
	alphaOrig := alpha
	var bestStep *Step

	// Step 1: Check the transposition table.
	if e.useTable {
		if entry, ok := e.table.ProbeDepth(p.zhash, maxDepth-depth); ok {
			bestStep = entry.Step
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
		}
	}

	if alpha >= beta {
		return alpha // fail-hard cutoff
	}

	// Step 2: Is this a terminal node or depth==0?
	// Start quiescense search.
	if depth == maxDepth || p.Terminal() {
		return e.quiescence(p, alpha, beta)
	}

	// Step 2a: Assertions.
	assert("!(0 < depth && depth < maxDepth)", 0 < depth && depth < maxDepth)

	// Step 2c. Try null move pruning.
	if e.nullMoveR > 0 && depth+e.nullMoveR < maxDepth {
		p.Pass()
		score := -e.search(p, -beta, -alpha, depth+e.nullMoveR+1, maxDepth)
		p.Unpass()
		if score >= beta {
			return beta // null move pruning
		}
	}

	steps := p.Steps()

	// Check if immobilized.
	if len(steps) == 0 {
		return -terminalEval
	}

	scoredSteps := e.scoreSteps(p, steps)
	sortSteps(scoredSteps)

	// Step 3: Main search.
	best := alpha
	nodes := 0
	for _, entry := range scoredSteps {
		n := entry.step.Len()
		if depth+n > maxDepth {
			continue
		}
		nodes++
		initSide := p.side

		if err := p.Step(entry.step); err != nil {
			panic(fmt.Sprintf("search_step: %v", err))
		}
		var score int
		if p.side == initSide {
			score = e.search(p, alpha, beta, depth+n, maxDepth)
		} else {
			score = -e.search(p, -beta, -alpha, depth+n, maxDepth)
		}
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("search_unstep: %v", err))
		}

		assert("!(score > -inf && score < inf)", score > -inf && score < inf)

		if score > best {
			best = score
			if bestStep == nil {
				bestStep = new(Step)
			}
			*bestStep = entry.step
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
			Depth: maxDepth - depth,
			Value: best,
			Step:  bestStep,
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
	assert("!(alpha >= -inf && alpha < beta && beta <= inf)",
		alpha >= -inf && alpha < beta && beta <= inf)

	eval := p.Score()
	if eval >= beta {
		return beta
	}
	if alpha < eval {
		alpha = eval
	}

	steps := p.Steps()
	scoredSteps := e.scoreSteps(p, steps)
	sortSteps(scoredSteps)
	nodes := 0
	for _, entry := range scoredSteps {
		if !entry.step.Capture() {
			continue
		}

		nodes++
		initSide := p.side

		if err := p.Step(entry.step); err != nil {
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
