package zoo

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const maxPly = 1024

const inf = 1000000

type SearchInfo struct {
	// cumulative nodes at the given ply
	// which can be used to compute the EBF.
	nodes []int // guarded by m

	// ply start times (unix nanoseconds) per ply.
	// Used to determine remaining time to search.
	times []int64

	// best move
	// time (unix nanoseconds) when best was last set.
	// Corresponds to end of search.
	bestTime int64 // guarded by m

	m sync.Mutex
}

func newSearchInfo() *SearchInfo {
	s := &SearchInfo{}
	s.addNodes(1, 0)
	return s
}

// locks excluded: s.m
func (s *SearchInfo) depth() int {
	return len(s.nodes)
}

func (s *SearchInfo) Depth() int {
	s.m.Lock()
	defer s.m.Unlock()
	return s.depth()
}

func (s *SearchInfo) addNodes(depth, v int) {
	s.m.Lock()
	defer s.m.Unlock()
	if depth > 0 {
		if n := len(s.nodes); n <= depth {
			// Start recording the next ply.
			s.times = append(s.times, time.Now().UnixNano())
			if n > 0 {
				v += s.nodes[n-1]
			}
			s.nodes = append(s.nodes, v)
			return
		}
		s.nodes[depth] += v
	}
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

// locks excluded: s.m.
func (s *SearchInfo) ebfInternal(d int, N float64) (b float64, err float64) {
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
// locks excluded: s.m.
func (s *SearchInfo) ebf() (b float64, err float64) {
	d := s.depth() - 1
	if d < 4 {
		// Avoid EBF instability in small values.
		return 1, 0
	}
	return s.ebfInternal(d, float64(s.nodes[d]))
}

// GuessPlyDuration guesses the duration of the next ply of search.
func (s *SearchInfo) GuessPlyDuration() time.Duration {
	s.m.Lock()
	defer s.m.Unlock()

	d := s.depth() - 1
	if d < 4 {
		// These searches are basically free.
		return 0
	}
	// Compute the ratio of the amount of time spent and nodes and solve for X.
	//  Time[d-1]        X
	// ----------  =  --------
	// Nodes[d-1]     Nodes[d]

	b, _ := s.ebf()
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

// Stop stops the search immediately and collects the final search results
// and blocks until all goroutines have stoped.
func (e *Engine) Stop() {
	if e.running == 1 && atomic.LoadInt32(&e.stopping) == 0 {
		e.stopping = 1
		for i := 0; i < e.concurrency; i++ {
			<-e.done
			// FIXME(ajzaff):
			// We would like to use this final result, but it's very unstable due
			// to cancelled search. Should we fix the instability we can use this.
		}
		e.stopping = 0
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

func (e *Engine) printSearchInfo(best searchResult) {
	s := e.searchInfo
	if e.ponder {
		e.Logf("ponder")
	}
	e.writef("info depth %d\n", s.Depth())
	e.writef("info time %d\n", s.Seconds())
	s.m.Lock()
	e.writef("info score %d\n", best.Score)
	s.m.Unlock()
	if pv, _, _ := e.table.PV(e.p); len(pv) > 0 {
		e.writef("info pv %s\n", MoveString(pv))
	}
	e.writef("info nodes %d\n", s.nodes[len(s.nodes)-1])
	e.Logf("log rate %d kN/s", s.searchRateKNps())
	e.Logf("log transpositions %d", e.table.Len())
}

func initWindowDelta(x int) int {
	v := x
	if v < 0 {
		v = -v
	}
	return 21 + v/256
}

type searchResult struct {
	Depth int
	Score int
	Move  []Step
}

func (e *Engine) iterativeDeepeningRoot() {
	if !e.lastPonder {
		e.table.Clear()
	}
	e.lastPonder = e.ponder
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
		best := searchResult{
			Move:  e.RandomSetup(),
			Depth: 1,
		}
		e.printSearchInfo(best)
		if !e.ponder {
			e.writef("bestmove %s\n", MoveString(best.Move))
		}
		return
	}

	// First ply searched on main goroutine.
	moves := e.getRootMovesLen(p, 1)
	rootSteps := make([]ScoredMove, len(moves))
	for i, move := range moves {
		rootSteps[i] = ScoredMove{score: -inf, move: move}
	}

	e.best = e.searchRoot(p, rootSteps, -inf, +inf, 1)
	e.printSearchInfo(e.best)
	resultChan := make(chan searchResult)

	// Implementation of Lazy SMP which runs parallel searches
	// with slightly varied root node orderings. This leads to
	// arriving at a given (deeper) position faster on average.
	for i := 0; i < e.concurrency; i++ {
		go func(rootScore int) {
			best := searchResult{Score: rootScore}

			defer func() {
				// Send the best move from the last (possibly partially cancelled) search.
				// TODO(ajzaff): In principle these searches may not be totally finished.
				// In practice, these move are often better than the last finished ply.
				// Find out if there's a better solution.
				e.done <- best
			}()

			newPos := p.Clone()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			// Generate all moves for use in this goroutine and add rootOrderNoise if configured.
			moves := e.getRootMovesLen(newPos, 4)
			// Shuffling moves increases variance in concurrent search.
			r.Shuffle(len(moves), func(i, j int) {
				moves[i], moves[j] = moves[j], moves[i]
			})
			scoredMoves := make([]ScoredMove, len(moves))
			for i, move := range moves {
				if e.rootOrderNoise != 0 {
					scoredMoves[i] = ScoredMove{score: int(float64(e.rootOrderNoise) * r.Float64()), move: move}
				}
			}

			// Main search loop:
			for depth := 2; depth < maxPly; depth++ {
				best.Depth = depth

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

				// Main aspiration loop:
				for {
					if res := e.searchRoot(newPos, scoredMoves, alpha, beta, depth); res.Score <= alpha {
						beta = (alpha + beta) / 2
						alpha = res.Score - delta
						if alpha < -inf {
							alpha = -inf
						}
						pv, _, _ := e.table.PV(newPos)
						if len(pv) > 0 {
							e.Logf("[%d] [<=%d] %s", res.Depth, res.Score, MoveString(pv))
						}
					} else if res.Score >= beta {
						beta = res.Score + delta
						if beta > inf {
							beta = inf
						}
						pv, _, _ := e.table.PV(newPos)
						if len(pv) > 0 {
							e.Logf("[%d] [>=%d] %s", res.Depth, res.Score, MoveString(pv))
						}
					} else {
						// The result was within the window.
						// Ready to search the next ply.
						best = res

						pv, _, _ := e.table.PV(newPos)
						if len(pv) > 0 {
							e.Logf("[%d] [==%d] %s", res.Depth, res.Score, MoveString(pv))
						}
						break
					}

					// Rescore the PV move and sort stably.
					for i := range scoredMoves {
						if MoveEqual(scoredMoves[i].move, best.Move) {
							scoredMoves[i].score = +inf
						} else {
							scoredMoves[i].score = -inf
						}
					}
					sortMoves(scoredMoves)

					// Update aspiration window delta
					delta += delta/4 + 5

					assert("!(alpha >= -inf && beta <= inf)", alpha >= -inf && beta <= inf)

					if e.stopping == 1 {
						break
					}
				}

				if e.stopping == 1 {
					break
				}

				// Send best move from (possibly cancelled) last ply to the done chan.
				resultChan <- best
			}
		}(e.best.Score)
	}

	// Collect search results and manage timeout.
	for e.running == 1 {
		next, rem := e.searchInfo.GuessPlyDuration(), e.timeControl.FixedOptimalTimeRemaining(e.timeInfo, p.side)
		select {
		case b := <-resultChan:
			if b.Depth > e.best.Depth || (b.Depth == e.best.Depth && b.Score > e.best.Score) {
				e.best = b

				// Log the new best PV.
				pv, _, _ := e.table.PV(p)
				if len(pv) > 0 {
					e.Logf("[%d] [%d] %s", e.best.Depth, e.best.Score, MoveString(pv))
				}
			}
		case <-time.After(time.Second):
			if e.fixedDepth == 0 && !e.ponder {
				if rem < 3*time.Second {
					fmt.Println("log stop search now to avoid timeout")
					e.Stop()
					break
				}
				if rem < next {
					// Time will soon be up! Stop the search.
					b, errv := e.searchInfo.ebf()
					e.Logf("log stop search now (b=%f{err=%f} cost=%s, budget=%s)", b, errv, next, rem)
					e.Stop()
				}
			}
		}

		if e.fixedDepth > 0 && e.searchInfo.Depth() >= e.fixedDepth {
			e.Logf("log stop search after fixed depth")
			e.Stop()
			break
		}
	}

	// Print search info and "bestmove" if not pondering.
	e.printSearchInfo(e.best)
	if !e.ponder {
		e.writef("bestmove %s\n", MoveString(e.best.Move))
	}
}

func (e *Engine) searchRoot(p *Pos, scoredMoves []ScoredMove, alpha, beta, depth int) searchResult {
	best := searchResult{
		Score: alpha,
		Depth: depth,
	}
	for i, entry := range scoredMoves {
		if e.stopping == 1 {
			break
		}
		n := MoveLen(entry.move)
		if n > depth {
			continue
		}
		if err := p.Move(entry.move); err != nil {
			panic(fmt.Sprintf("search_move_root: %s: %v", entry.move, err))
		}
		var stepList StepList
		score := -e.search(p, &stepList, i == 0, -beta, -alpha, n, depth)
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
	e.table.StoreMove(p, depth, best.Score, best.Move)
	return best
}

func (e *Engine) search(p *Pos, stepList *StepList, pv bool, alpha, beta, depth, maxDepth int) int {
	alphaOrig := alpha
	var bestStep *Step

	// Step 1: Check the transposition table.
	var tableMove bool
	if e.useTable {
		var entry *TableEntry
		if entry, pv = e.table.ProbeDepth(p.zhash, maxDepth-depth); pv {
			tableMove = true
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
	pv = pv && tableMove

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
		score := -e.search(p, stepList, pv, -beta, -alpha, depth+e.nullMoveR+1, maxDepth)
		p.Unpass()
		if score >= beta {
			return score // null move pruning
		}
	}

	if pv && maxDepth-depth >= 6 {
		// If position is not in table, and is PV line, decrease maxDepth by 2.
		// FIXME(ajzaff): I copied this 'optimization' blindly from Stockfish
		// but I haven't validated if it works yet.
		maxDepth -= 2
	}

	// Generate steps at this ply and add them to the list.
	// We will start search from move l and later truncate the list to this initial length.
	l := stepList.Len()
	stepList.Generate(p)

	// Check if immobilized.
	if l == stepList.Len() {
		return -terminalEval
	}

	// Step 3: Main search.
	selector := e.stepSelector(stepList.steps[l:])
	best := alpha
	nodes := 0
	first := true
	for step, ok := selector.Select(); ok; step, ok = selector.Select() {
		n := step.Len()
		if depth+n > maxDepth {
			continue
		}
		nodes++
		initSide := p.side

		if err := p.Step(step); err != nil {
			panic(fmt.Sprintf("search_step: %v", err))
		}
		var score int
		if p.side == initSide {
			score = e.search(p, stepList, pv && first, alpha, beta, depth+n, maxDepth)
		} else {
			score = -e.search(p, stepList, pv && first, -beta, -alpha, depth+n, maxDepth)
		}
		first = false
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("search_unstep: %v", err))
		}

		assert("!(score > -inf && score < inf)", score > -inf && score < inf)

		if score > best {
			best = score
			if bestStep == nil {
				bestStep = new(Step)
			}
			*bestStep = step
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break // fail-hard cutoff
		}
		if e.stopping == 1 {
			break
		}
	}

	// Truncate steps generated at this ply.
	stepList.Truncate(l)

	e.searchInfo.addNodes(depth, nodes)

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

	steps := make([]Step, 0, 20)
	selector := e.stepSelector(steps)

	nodes := 0
	for step, ok := selector.SelectCapture(); ok; step, ok = selector.SelectCapture() {
		nodes++
		initSide := p.side

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
	return alpha
}
