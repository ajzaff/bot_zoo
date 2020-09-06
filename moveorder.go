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

	for i, move := range moves {
		func() {
			defer func() {
				if err := p.Unmove(); err != nil {
					panic(fmt.Sprintf("moveorder: %v", err))
				}
			}()
			if err := p.Move(move); err != nil {
				if err != errRecurringPosition {
					panic(fmt.Sprintf("moveorder: %v", err))
				}
				a[i].score = -inf
				numIllegal++
				return
			}

			var score int

			// Step 1a: Table lookup.
			if e.useTable {
				if entry, ok := e.table.ProbeDepth(p.zhash, 0); ok {
					score := entry.Value
					if entry.Bound == ExactBound {
						a[i].score = inf + score
					}
				} else {
					score = -p.Score()
				}
			} else {
				// Step 1b: Fallback to eval score.
				score = -p.Score()
			}

			a[i].score = score
		}()
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

	numIllegal := 0

	for i, step := range steps {
		func() {
			defer func() {
				if err := p.Unstep(); err != nil {
					panic(fmt.Sprintf("steporder: %v", err))
				}
			}()

			initSide := p.side

			if err := p.Step(step); err != nil {
				a[i].score = -inf
				numIllegal++
				return
			}

			var score int

			// Step 1a: Table lookup.
			if e.useTable {
				if entry, ok := e.table.ProbeDepth(p.zhash, 0); ok {
					score := entry.Value
					if entry.Bound == ExactBound {
						a[i].score = inf + score
					}
				} else {
					score = p.Score()
				}
			} else {
				// Step 1b: Fallback to eval score.
				score = p.Score()
			}

			if initSide != p.side {
				score = -score
			}

			a[i].score = score
		}()
	}

	sort.Stable(byStepScore(a))
	a = a[:len(a)-numIllegal]

	return a
}
