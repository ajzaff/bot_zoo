package zoo

import (
	"fmt"
	"math/rand"
	"sort"
)

type ScoredStep struct {
	score int
	step  Step
}

type byStepScore []ScoredStep

func (a byStepScore) Len() int           { return len(a) }
func (a byStepScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStepScore) Less(i, j int) bool { return a[i].score > a[j].score }

type ScoredMove struct {
	score int
	move  []Step
}

type byLen []ScoredMove

func (a byLen) Len() int           { return len(a) }
func (a byLen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLen) Less(i, j int) bool { return len(a[i].move) > len(a[j].move) }

type byScore []ScoredMove

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].score > a[j].score }

// perturbMoves adds noise to the scoredMoves in [-f, +f].
// Intended to be used for parallel Lazy-SMP search.
func (e *Engine) perturbMoves(r *rand.Rand, f float64, scoredMoves []ScoredMove) {
	for i := range scoredMoves {
		scoredMoves[i].score += int(f * r.NormFloat64())
	}
}

func sortMoves(a []ScoredMove) {
	sort.Stable(byLen(a))
	sort.Stable(byScore(a))
}

// sortMoves computes move scores based on naive eval.
// call rescorePVMoves to update the score after finding a PV.
func (e *Engine) scoreMoves(p *Pos, moves [][]Step) []ScoredMove {
	a := make([]ScoredMove, 0, len(moves))

	// 2: Range over all moves.
	for _, move := range moves {

		// 2a. Try the move.
		// In the unlikely case it's illegal give a score of -inf.
		// We want to prune this move from the results.
		if err := p.Move(move); err != nil {
			if err != errRecurringPosition {
				panic(fmt.Sprintf("moveorder_move: %v", err))
			}
		} else {
			// Step 2b: Get the evaluation and set the score.
			// Negate the score since sides have changed.
			a = append(a, ScoredMove{
				score: -p.Score(),
				move:  move,
			})
		}

		// Step 2c. Undo the move.
		if err := p.Unmove(); err != nil {
			panic(fmt.Sprintf("moveorder_unmove: %v", err))
		}
	}
	return a
}

// rescorePVMoves updates the scored moves for the PV
// to +inf and sets everything else to -inf.
func (e *Engine) rescorePVMoves(p *Pos, scoredMoves []ScoredMove) {

	// 1: Check the table to get the PV move if any.
	if !e.useTable {
		return
	}
	var best []Step
	v, _, _ := e.table.Best(p)
	if len(v) == 0 {
		return
	}
	best = v

	// 2: Range over all moves.
	for i := range scoredMoves {
		scoredMove := scoredMoves[i]

		// 2a: Update the best move from the table to +inf.
		if MoveEqual(scoredMove.move, best) {
			scoredMoves[i].score = +inf
			continue
		}

		scoredMoves[i].score = -inf
	}
}

// perturbMoves adds noise to the scoredMoves in [-f, +f].
// Intended to be used for parallel Lazy-SMP search.
func (e *Engine) perturbSteps(r *rand.Rand, f float64, scoredSteps []ScoredStep) {
	for i := range scoredSteps {
		scoredSteps[i].score += int(f * r.NormFloat64())
	}
}

func sortSteps(a []ScoredStep) {
	sort.Stable(byStepScore(a))
}

func (e *Engine) scoreSteps(p *Pos, steps []Step) []ScoredStep {
	a := make([]ScoredStep, 0, len(steps))

	// 1: Check the table to get the PV step if any.
	var pv Step
	if e.useTable {
		entry, ok := e.table.ProbeDepth(p.zhash, 0)
		if ok && entry.Bound == ExactBound && entry.Step != nil {
			pv = *entry.Step
		}
	}

	// 2: Range over all steps.
	for _, step := range steps {

		// 2a: Update the PV score to +inf.
		if e.useTable && step == pv {
			a = append(a, ScoredStep{
				score: +inf,
				step:  step,
			})
			continue
		}

		initSide := p.Side()

		// 2a. Try the step.
		// In the unlikely case it's illegal give a score of -inf.
		if err := p.Step(step); err != nil {
			// TODO(ajzaff): Handle error.
		} else {
			// Step 2b: Get the evaluation and set the score.
			// Negate the score if sides changed.
			score := p.Score()
			if p.Side() != initSide {
				score = -score
			}
			a = append(a, ScoredStep{
				score: score,
				step:  step,
			})
		}

		// Step 2c. Undo the step.
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("steporder: %v", err))
		}
	}
	return a
}
