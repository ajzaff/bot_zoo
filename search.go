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

type searchInfo struct {
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

func newSearchInfo() *searchInfo {
	s := &searchInfo{}
	s.addNodes(1, 0)
	return s
}

// locks excluded: s.m
func (s *searchInfo) depth() int {
	return len(s.nodes)
}

func (s *searchInfo) Depth() int {
	s.m.Lock()
	defer s.m.Unlock()
	return s.depth()
}

func (s *searchInfo) addNodes(depth, v int) {
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
func (s *searchInfo) Nodes() int {
	return s.nodes[len(s.nodes)-1]
}

// Seconds returns the number of seconds searched in all ply.
func (s *searchInfo) Seconds() int64 {
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
func (s *searchInfo) ebfInternal(d int, N float64) (b float64, err float64) {
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
func (s *searchInfo) ebf() (b float64, err float64) {
	d := s.depth() - 1
	if d < 4 {
		// Avoid EBF instability in small values.
		return 1, 0
	}
	return s.ebfInternal(d, float64(s.nodes[d]))
}

// GuessPlyDuration guesses the duration of the next ply of search.
func (s *searchInfo) GuessPlyDuration() time.Duration {
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
	if b == 1 {
		// EFB is not set, we should be ok to proceed.
		return 0
	}
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
			panic(fmt.Sprintf("SEARCH_ERROR recovered: %v\n", r))
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
				panic(fmt.Sprintf("SEARCH_ERROR_FIXED recovered: %v\n", r))
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

// Stop signals the search to stop immediately.
func (e *Engine) Stop() {
	if atomic.CompareAndSwapInt32(&e.stopping, 0, 1) {
	loop:
		for {
			select {
			case <-e.resultChan:
			default:
				break loop
			}
		}
		e.wg.Wait()
		e.running = 0
		e.stopping = 0
	}
}

func (s *searchInfo) searchRateKNps() int64 {
	ns := s.times[len(s.times)-1] - s.times[0]
	if ns == 0 {
		return 0
	}
	n := s.nodes[len(s.nodes)-1]
	return int64(float64(n) / (float64(ns) / float64(time.Second)) / 1000)
}

func (e *Engine) printSearchInfo(pv []Step, best searchResult) {
	s := e.searchInfo
	if e.ponder {
		e.Logf("ponder")
	}
	e.writef("info depth %d\n", s.Depth())
	e.writef("info time %d\n", s.Seconds())
	s.m.Lock()
	e.writef("info score %d\n", best.Value)
	s.m.Unlock()
	e.writef("info pv %s\n", MoveString(pv))
	e.writef("info nodes %d\n", s.nodes[len(s.nodes)-1])
	e.Logf("rate %d kN/s", s.searchRateKNps())
	e.Logf("transpositions %d", e.table.Len())
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
	Value int
	Move  []Step
	PV    []Step
}

type NodeType int

const (
	PV NodeType = iota
	NonPV
)

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
		e.printSearchInfo(nil, best)
		if !e.ponder {
			e.writef("bestmove %s\n", MoveString(best.Move))
		}
		return
	}

	e.best = searchResult{Value: -inf}
	e.resultChan = make(chan searchResult)
	e.wg.Add(e.concurrency)

	// Implementation of Lazy SMP which runs parallel searches
	// with slightly varied root node orderings. This leads to
	// arriving at a given (deeper) position faster on average.
	for i := 0; i < e.concurrency; i++ {
		go func(rootScore int) {
			defer func() {
				// FIXME(ajzaff):
				// We would like to use this final result, but it's very unstable due
				// to cancelled search. Should we fix the instability we can use this.
				// Reduce the tracking count of running goroutines by 1.
				e.wg.Done()
			}()
			p := p.Clone() // shadowed

			// Generate all moves for use in this goroutine and add rootOrderNoise if configured.
			// Shuffling moves increases variance in concurrent search.
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			moves := e.getRootMovesLen(p, 4)
			r.Shuffle(len(moves), func(i, j int) {
				moves[i], moves[j] = moves[j], moves[i]
			})
			scoredMoves := make([]ScoredMove, len(moves))
			for i, move := range moves {
				if e.rootOrderNoise != 0 {
					scoredMoves[i] = ScoredMove{score: int(float64(e.rootOrderNoise) * r.Float64()), move: move}
				}
			}

			alpha, beta := -inf, +inf
			stack := make([]Stack, 1, maxPly)
			best := searchResult{Value: -inf}

			// Main search loop:
		mainLoop:
			for depth := 1; depth < maxPly; depth++ {
				// Aspiration window tracks the previous score for search.
				// Whenever we fail high or low widen the window.
				// Start with a small aspiration window and, in the case of a fail
				// high/low, re-search with a bigger window until we don't fail
				// high/low anymore.
				var delta int
				if depth >= 4 {
					delta = initWindowDelta(best.Value)
					alpha, beta = best.Value-delta, best.Value+delta
					if alpha < -inf {
						alpha = -inf
					}
					if beta > inf {
						beta = inf
					}
				}

				stack = append(stack, Stack{})

				// Search all root moves using aspiration and call search on each.
				for failHighCount := 0; ; {
					adjustedDepth := depth - failHighCount
					if adjustedDepth < 1 {
						adjustedDepth = 1
					}
					result := e.searchRoot(p, &stack, scoredMoves, alpha, beta, adjustedDepth)

					if Terminal(result.Value) {
						// Stop the search goroutine if a terminal value is achieved.
						best.Value = result.Value
						best.Move = result.Move
						best.PV = result.PV
						break
					}

					if result.Value <= alpha {
						e.Logf("[%d] [<=%d] %s", depth, alpha, MoveString(result.PV))
						beta = (alpha + beta) / 2
						alpha = result.Value - delta
						if alpha < -inf {
							alpha = -inf
						}
						failHighCount = 0
					} else if result.Value >= beta {
						e.Logf("[%d] [>=%d] %s", depth, beta, MoveString(result.PV))
						beta = result.Value + delta
						if beta > inf {
							beta = inf
						}
						failHighCount++
					} else {
						// The result was within the window.
						// Ready to search the next ply.
						best.Value = result.Value
						best.Move = result.Move
						best.Depth = depth

						e.Logf("[%d] [==%d] %s", depth, best.Value, MoveString(result.PV))
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
						break mainLoop
					}
				}

				if e.stopping == 1 {
					break mainLoop
				}

				// Send best move from (possibly cancelled) last ply to the done chan.
				e.resultChan <- best
			}
		}(e.best.Value)
	}

	// Track the number of running goroutines.
	// Collect search results and manage timeout.
	for e.running == 1 {
		next, rem := e.searchInfo.GuessPlyDuration(), e.timeControl.FixedOptimalTimeRemaining(e.timeInfo, p.side)
		select {
		case b := <-e.resultChan:
			if b.Depth > e.best.Depth || (b.Depth == e.best.Depth && b.Value > e.best.Value) {
				e.best = b

				e.Logf("[%d] [%d] %s", e.best.Depth, e.best.Value, MoveString(e.best.PV))

				if Winning(e.best.Value) {
					// Stop after finding a winning move.
					// We keep searching losing moves until MaxPly.
					// The only winning move is not to play.
					e.Stop()
				}
			}
		case <-time.After(time.Second):
			if e.fixedDepth == 0 && !e.ponder {
				if e.timeControl.GameTimeRemaining(e.timeInfo, p.side) < 3*time.Second {
					e.Logf("stop search now to avoid timeout")
					e.Stop()
					break
				}
				if e.searchInfo.Depth() >= e.minDepth && rem < next {
					// Time will soon be up! Stop the search.
					e.Logf("stop search now (cost=%s, budget=%s)", next, rem)
					e.Stop()
				}
			}
		}

		if e.fixedDepth > 0 && e.searchInfo.Depth() >= e.fixedDepth {
			e.Logf("stop search after fixed depth")
			e.Stop()
		}
	}

	// Print search info and "bestmove" if not pondering.
	e.printSearchInfo(e.best.PV, e.best)
	if !e.ponder {
		e.writef("bestmove %s\n", MoveString(e.best.Move))
	}
}

func (e *Engine) searchRoot(p *Pos, stack *[]Stack, scoredMoves []ScoredMove, alpha, beta, depth int) searchResult {
	best := searchResult{
		Value: -terminalEval,
		Depth: depth,
	}
	for _, entry := range scoredMoves {
		if e.stopping == 1 {
			break
		}
		n := MoveLen(entry.move)
		if n > depth {
			continue
		}
		err := p.Move(entry.move)
		if err != nil {
			if err != errRecurringPosition {
				panic(fmt.Sprintf("search_move_root: %s: %v", entry.move, err))
			}
		} else {
			var stepList StepList
			value := -e.search(PV, p, stack, &stepList, -beta, -alpha, n, depth)
			if value > alpha {
				alpha = value
			}
			if value > best.Value || best.Move == nil {
				best.Value = value
				best.Move = entry.move
			}
		}
		if err := p.Unmove(); err != nil {
			panic(fmt.Sprintf("search_unmove_root: %s: %v", entry.move, err))
		}
	}
	// Update PV
	pv := make([]Step, len(best.Move), len(best.Move)+len(*stack)-2)
	copy(pv, best.Move)
	for _, e := range (*stack)[1 : len(*stack)-1] {
		pv = append(pv, e.Step)
	}
	best.PV = pv
	return best
}

func (e *Engine) search(nt NodeType, p *Pos, stack *[]Stack, stepList *StepList, alpha, beta, depth, maxDepth int) int {
	alphaOrig := alpha
	var bestStep *Step

	// Step 1: Check the transposition table.
	var tableMove bool
	if e.useTable {
		if entry, ok := e.table.ProbeDepth(p.zhash, maxDepth-depth); ok {
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

	if alpha >= beta {
		return alpha // fail-hard cutoff
	}

	// Step 2: Is this a terminal node or depth==0?
	// Start quiescense search.
	if depth >= maxDepth || p.Terminal() {
		return e.quiescence(nt, p, alpha, beta)
	}

	// Step 2a: Assertions.
	assert("!(0 < depth && depth < maxDepth)", 0 < depth && depth < maxDepth)

	// Step 2c. Try null move pruning.
	if eval := p.Score(); nt != PV && eval >= beta && e.nullMoveR > 0 && depth+e.nullMoveR < maxDepth {
		p.Pass()
		nullValue := -e.search(NonPV, p, stack, stepList, -beta, -alpha, depth+e.nullMoveR+1, maxDepth)
		p.Unpass()
		if nullValue >= beta {
			return nullValue // null move pruning
		}
	}

	if nt != PV && !tableMove && maxDepth-depth >= 8 {
		// If position is not in table, and is PV line, decrease maxDepth by 2.
		maxDepth -= 2
	}

	// Generate steps at this ply and add them to the list.
	// We will start search from move l and later truncate the list to this initial length.
	l := stepList.Len()
	stepList.Generate(p)

	// Step 3: Main search.
	selector := e.stepSelector(stepList.steps[l:])
	nodes := 0

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
		var value int
		if p.side == initSide {
			value = e.search(nt, p, stack, stepList, alpha, beta, depth+n, maxDepth)
		} else {
			value = -e.search(nt, p, stack, stepList, -beta, -alpha, depth+n, maxDepth)
		}
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("search_unstep: %v", err))
		}

		assert("!(value > -inf && value < inf)", value > -inf && value < inf)

		if value > alpha {
			alpha = value
			if bestStep == nil {
				bestStep = new(Step)
			}
			*bestStep = step
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

	// Update stack and tracing info.
	e.searchInfo.addNodes(depth, nodes)
	(*stack)[depth].Depth = depth
	(*stack)[depth].Eval = alpha

	// Update PV.
	if nt == PV {
		var pv []Step
		if bestStep != nil {
			(*stack)[depth].Step = *bestStep
			pv = append(pv, *bestStep)
		}
		if depth+1 < len(*stack) {
			pv = append(pv, (*stack)[depth+1].PV...)
		}
		(*stack)[depth].PV = pv
	}

	// Step 4: Store transposition table entry.
	if e.useTable {
		entry := &TableEntry{
			ZHash: p.zhash,
			Depth: maxDepth - depth,
			Value: alpha,
			Step:  bestStep,
		}
		switch {
		case alpha <= alphaOrig:
			entry.Bound = UpperBound
		case alpha >= beta:
			entry.Bound = LowerBound
		default:
			entry.Bound = ExactBound
		}
		e.table.Store(entry)
	}

	// Return best score.
	return alpha
}

// TODO(ajzaff): Measure the effect of counting quienscence nodes on the EBF.
// This has direct consequences on move timings.
func (e *Engine) quiescence(nt NodeType, p *Pos, alpha, beta int) int {
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

	for step, ok := selector.SelectCapture(); ok; step, ok = selector.SelectCapture() {
		initSide := p.side

		if err := p.Step(step); err != nil {
			panic(fmt.Sprintf("quiescense_step: %v", err))
		}
		var score int
		if p.side == initSide {
			score = e.quiescence(nt, p, alpha, beta)
		} else {
			score = -e.quiescence(nt, p, -beta, -alpha)
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
