package zoo

import (
	"fmt"
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

func (e *Engine) sortMoves(p *Pos, moves [][]Step) []ScoredMove {
	a := make([]ScoredMove, len(moves))
	for i := range a {
		a[i].move = moves[i]
		a[i].score = -terminalEval
	}

	numIllegal := 0

	// 1: Check the table to get the best move if any.
	var best []Step
	if e.useTable {
		if v, _, _ := e.table.Best(p); len(v) > 0 {
			best = v
		}
	}

	// 2: Range over all moves.
	for i, move := range moves {

		// 2a: Update the best move from the table to +inf.
		if e.useTable && MoveEqual(move, best) {
			a[i].score = +inf
			continue
		}

		// 2a. Try the move.
		// In the unlikely case it's illegal give a score of -inf.
		// We want to prune this move from the results.
		if err := p.Move(move); err != nil {
			if err != errRecurringPosition {
				panic(fmt.Sprintf("moveorder_move: %v", err))
			}
			a[i].score = -inf
			numIllegal++
			continue
		}

		// Step 2b: Get the evaluation and set the score.
		// Negate the score since sides have changed.
		a[i].score = -p.Score()

		// Step 2c. Undo the move.
		if err := p.Unmove(); err != nil {
			panic(fmt.Sprintf("moveorder_unmove: %v", err))
		}
	}
	sort.Sort(byLen(a))
	sort.Stable(byScore(a))
	a = a[:len(a)-numIllegal]
	return a
}

func (e *Engine) sortSteps(p *Pos, steps []Step) []ScoredStep {
	a := make([]ScoredStep, len(steps))
	for i := range a {
		a[i].step = steps[i]
		a[i].score = -terminalEval
	}

	// 1: Check the table to get the PV step if any.
	var pv Step
	if e.useTable {
		entry, ok := e.table.ProbeDepth(p.zhash, 0)
		if ok && entry.Bound == ExactBound && entry.Step != nil {
			pv = *entry.Step
		}
	}

	// 2: Range over all steps.
	for i, step := range steps {

		// 2a: Update the PV score to +inf.
		if e.useTable && step == pv {
			a[i].score = +inf
			continue
		}

		initSide := p.Side()

		// 2a. Try the step.
		// In the unlikely case it's illegal give a score of -inf.
		if err := p.Step(step); err != nil {
			a[i].score = -inf
			continue
		}

		// Step 2b: Get the evaluation and set the score.
		// Negate the score if sides changed.
		score := p.Score()
		if p.Side() != initSide {
			score = -score
		}
		a[i].score = score

		// Step 2c. Undo the step.
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("steporder: %v", err))
		}
	}
	sort.Stable(byStepScore(a))
	return a
}
