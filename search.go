package zoo

import (
	"fmt"
	"math/rand"
	"time"
)

const maxPly int16 = 1024

type nodeType int

const (
	PV nodeType = iota
	NonPV
)

type searchResult struct {
	ID    int
	Depth int16
	Value Value
	Nodes int
	Move  []Step
	PV    []Step
}

func initWindowDelta(x Value) Value {
	v := x
	if v < 0 {
		v = -v
	}
	return 21 + v/256
}

func (e *Engine) iterativeDeepeningRoot() {
	if !e.lastPonder {
		e.table.Clear()
	}
	e.lastPonder = e.ponder
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
		if !e.ponder {
			e.writef("bestmove %s\n", MoveString(best.Move))
		}
		return
	}

	e.best = searchResult{Value: -Inf}
	e.resultChan = make(chan searchResult)
	e.wg.Add(e.concurrency)

	// Implementation of Lazy SMP which runs parallel searches
	// with slightly varied root node orderings. This leads to
	// arriving at a given (deeper) position faster on average.
	for i := 0; i < e.concurrency; i++ {
		go func(id int, rootValue Value) {
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
					scoredMoves[i] = ScoredMove{score: Value(float64(e.rootOrderNoise) * r.Float64()), move: move}
				}
			}

			alpha, beta := -Inf, +Inf
			stack := make([]Stack, 1, maxPly)
			best := searchResult{ID: id, Value: -Inf}

			// Main search loop:
		mainLoop:
			for depth := int16(1); depth < maxPly; depth++ {
				// Aspiration window tracks the previous score for search.
				// Whenever we fail high or low widen the window.
				// Start with a small aspiration window and, in the case of a fail
				// high/low, re-search with a bigger window until we don't fail
				// high/low anymore.
				var delta Value
				if depth >= 4 {
					delta = initWindowDelta(best.Value)
					alpha, beta = best.Value-delta, best.Value+delta
					if alpha < -Inf {
						alpha = -Inf
					}
					if beta > Inf {
						beta = Inf
					}
				}

				stack = append(stack, Stack{})

				// Search all root moves using aspiration and call search on each.
				for {
					adjustedDepth := depth
					if adjustedDepth < 1 {
						adjustedDepth = 1
					}
					move, value := e.searchRoot(p, &stack, scoredMoves, alpha, beta, adjustedDepth)

					if value.Terminal() {
						// Stop the search goroutine if a terminal value is achieved.
						best.Value = value
						best.Move = move
						best.Nodes = stack[0].Nodes
						best.PV = stack[0].PV
						break
					}

					if value <= alpha {
						e.Logf("%d (%d] %s", depth, alpha, MoveString(stack[0].PV))

						beta = (alpha + beta) / 2
						alpha = value - delta
						if alpha < -Inf {
							alpha = -Inf
						}

						// TODO(ajzaff): We want to extend the search after failing low.
						// Notify the main thread of a fail low immediately to avoid
						// submitting a suboptimal value from another search.
					} else if value >= beta {
						e.Logf("%d [%d) %s", depth, beta, MoveString(stack[0].PV))

						beta = value + delta
						if beta > Inf {
							beta = Inf
						}
					} else {
						// The result was within the window.
						// Ready to search the next ply.
						e.Logf("%d [%d] %s", depth, value, MoveString(stack[0].PV))

						best.Value = value
						best.Move = move
						best.Depth = depth
						best.Nodes = stack[0].Nodes
						best.PV = stack[0].PV
						break
					}

					// Rescore the PV move and sort stably.
					for i := range scoredMoves {
						if MoveEqual(scoredMoves[i].move, best.Move) {
							scoredMoves[i].score = +Inf
						} else {
							scoredMoves[i].score = -Inf
						}
					}
					sortMoves(scoredMoves)

					// Update aspiration window delta
					delta += delta/4 + 5

					assert("!(alpha >= -inf && beta <= inf)", alpha >= -Inf && beta <= Inf)

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
		}(i, e.best.Value)
	}

	// Track the depth for each goroutine to validate mindepth.
	goroutineDepths := make(map[int]int16) // goroutine => depth
	var mindepth = func() int16 {
		min := int16(-1)
		for _, d := range goroutineDepths {
			if min == -1 || d < min {
				min = d
			}
		}
		return min
	}

	// Track the number of running goroutines.
	// Collect search results and manage timeout.
	for e.running == 1 {
		budget := e.timeControl.GameTimeRemaining(e.timeInfo, p.side)
		if budget > 30*time.Second {
			budget = 30 * time.Second
		}
		deadline := e.timeInfo.Start[e.Pos().Side()].Add(budget)
		select {
		case b := <-e.resultChan:
			if b.Depth > e.best.Depth || (b.Depth == e.best.Depth && b.Value > e.best.Value) {
				e.best.Depth = b.Depth
				e.best.Value = b.Value
				e.best.Nodes += b.Nodes
				e.best.Move = b.Move
				e.best.PV = b.PV
				goroutineDepths[b.ID] = b.Depth

				e.Logf("%d [%d] %s", e.best.Depth, e.best.Value, MoveString(e.best.PV))

				if e.best.Value.Winning() {
					// Stop after finding a winning move.
					// We keep searching losing moves until MaxPly.
					// The only winning move is not to play.
					e.Stop()
				}
			}
		case <-time.After(time.Second):
			if e.fixedDepth == 0 && !e.ponder {
				if rem := e.timeControl.GameTimeRemaining(e.timeInfo, e.Pos().Side()); rem < 3*time.Second {
					e.Logf("stop search now to avoid timeout (remaining=%s)", rem)
					e.Stop()
					break
				}
				rem := deadline.Sub(time.Now())
				if mindepth() > e.minDepth && rem < 1*time.Second {
					e.Logf("stop search now (budget=%s, remaining=%s)", budget, rem)
					e.Stop()
				}
			}
		}
		if e.fixedDepth > 0 && e.best.Depth >= e.fixedDepth {
			e.Logf("stop search after fixed depth")
			e.Stop()
		}
	}

	// Print search info and "bestmove" if not pondering.
	e.printSearchInfo(e.best.Nodes, e.best.Depth, e.timeInfo.Start[e.Pos().Side()], e.best)
	if !e.ponder {
		e.writef("bestmove %s\n", MoveString(e.best.Move))
	}
}

func (e *Engine) searchRoot(p *Pos, stack *[]Stack, scoredMoves []ScoredMove, alpha, beta Value, depth int16) (bestMove []Step, bestValue Value) {

	// Step 1. Terminal node? We have no moves.
	eval := p.Value()
	if eval.Terminal() {
		return nil, eval
	}

	// Step 2. Full root search.
	bestValue = -Win
	for _, entry := range scoredMoves {
		if e.stopping == 1 {
			break
		}
		n := MoveLen(entry.move)
		if err := p.Move(entry.move); err != nil {
			if err != errRecurringPosition {
				panic(fmt.Sprintf("search_move_root: %s: %v", entry.move, err))
			}
		} else {
			var stepList StepList
			value := -e.search(PV, p, stack, &stepList, -beta, -alpha, int16(n), depth)
			if value > alpha {
				alpha = value
			}
			if value > bestValue || bestMove == nil {
				// Update PV and best move.
				(*stack)[0].PV = make([]Step, len(entry.move))
				copy((*stack)[0].PV, entry.move)
				if n < len(*stack) {
					(*stack)[0].PV = append((*stack)[0].PV, (*stack)[n].PV...)
				}
				bestMove = entry.move
				bestValue = value
			}
		}
		if err := p.Unmove(); err != nil {
			panic(fmt.Sprintf("search_unmove_root: %s: %v", entry.move, err))
		}
	}

	// Store best move
	if e.useTable {
		e.table.StoreMove(p, depth, bestValue, bestMove)
	}

	(*stack)[0].Nodes = 1
	for i := 1; i < 5 && i < len(*stack); i++ {
		(*stack)[0].Nodes += (*stack)[i].Nodes
	}
	return bestMove, bestValue
}

func (e *Engine) search(nt nodeType, p *Pos, stack *[]Stack, stepList *StepList, alpha, beta Value, depth, maxDepth int16) Value {
	alphaOrig := alpha
	var bestStep *Step

	// Step 1: Check the transposition table.
	var tableMove bool
	if e.useTable {
		if entry, ok := e.table.ProbeDepth(p.zhash, maxDepth-depth); ok {
			// tableMove = true
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

	// Step 1a. Terminal node?
	eval := p.Value()
	if eval.Terminal() {
		return eval
	}

	// Step 2: Is this a terminal node or depth==0?
	// Start quiescense search.
	if depth >= maxDepth || p.Terminal() {
		return e.quiescence(nt, p, stepList, alpha, beta)
	}

	// Step 2a: Assertions.
	assert("!(0 < depth && depth < maxDepth)", 0 < depth && depth < maxDepth)

	// Step 2c. Try null move pruning.
	if nt != PV && eval >= beta && e.nullMoveR > 0 && depth+e.nullMoveR < maxDepth {
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
	selector := newStepSelector(p.side, stepList.steps[l:])

	for step, ok := selector.Select(); ok; step, ok = selector.Select() {
		n := step.Len()
		if depth+int16(n) > maxDepth {
			continue
		}
		initSide := p.side

		if err := p.Step(step); err != nil {
			panic(fmt.Sprintf("search_step: %v", err))
		}
		var value Value
		if p.side == initSide {
			value = e.search(nt, p, stack, stepList, alpha, beta, depth+int16(n), maxDepth)
		} else {
			value = -e.search(nt, p, stack, stepList, -beta, -alpha, depth+int16(n), maxDepth)
		}
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("search_unstep: %v", err))
		}

		assert("!(value > -inf && value < inf)", value > -Inf && value < Inf)

		if value > alpha {
			alpha = value
			if bestStep == nil {
				bestStep = new(Step)
			}
			*bestStep = step

			// Update PV.
			if nt == PV {
				var pv []Step
				if bestStep != nil {
					(*stack)[depth].Step = *bestStep
					pv = append(pv, *bestStep)
				}
				if int(depth+1) < len(*stack) {
					pv = append(pv, (*stack)[depth+1].PV...)
				}
				(*stack)[depth].PV = pv
			}

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
	(*stack)[depth].Depth = depth
	(*stack)[depth].Eval = alpha

	if nt == PV {
		(*stack)[depth].Nodes = 1
		if int(depth+1) < len(*stack) {
			(*stack)[depth].Nodes += (*stack)[depth+1].Nodes
		}
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
func (e *Engine) quiescence(nt nodeType, p *Pos, stepList *StepList, alpha, beta Value) Value {
	assert("!(alpha >= -inf && alpha < beta && beta <= inf)",
		alpha >= -Inf && alpha < beta && beta <= Inf)

	eval := p.Value()
	return eval
	// if eval >= beta {
	// 	return beta
	// }
	// if alpha < eval {
	// 	alpha = eval
	// }

	// // Generate steps and add them to the list.
	// // We will start search from move l and later truncate the list to this initial length.
	// l := stepList.Len()
	// stepList.Generate(p)

	// selector := newStepSelector(p.side, stepList.steps[l:])

	// for step, ok := selector.SelectCapture(); ok; step, ok = selector.SelectCapture() {
	// 	initSide := p.side

	// 	if err := p.Step(step); err != nil {
	// 		panic(fmt.Sprintf("quiescense_step: %v", err))
	// 	}
	// 	var value Value
	// 	if p.side == initSide {
	// 		value = e.quiescence(nt, p, stepList, alpha, beta)
	// 	} else {
	// 		value = -e.quiescence(nt, p, stepList, -beta, -alpha)
	// 	}
	// 	if err := p.Unstep(); err != nil {
	// 		panic(fmt.Sprintf("quiescense_unstep: %v", err))
	// 	}
	// 	if value >= beta {
	// 		return beta
	// 	}
	// 	if value > alpha {
	// 		alpha = value
	// 	}
	// }

	// // Truncate steps generated at this ply.
	// stepList.Truncate(l)

	// return alpha
	// return alpha
}
