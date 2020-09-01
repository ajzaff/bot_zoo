package zoo

import (
	"log"
	"sort"
)

type ScoredStep struct {
	score int
	step  Step
}

type byStepScore []ScoredMove

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

	for i, move := range moves {
		if err := p.Move(move); err != nil {
			if err != errRecurringPosition {
				log.Printf("moveorder: %v", err)
			}
			a[i].score = -inf
			continue
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

		if err := p.Unmove(move); err != nil {
			panic(err)
		}
	}
	sort.Sort(byLen(a))
	sort.Stable(byScore(a))

	// Experimental pruning of non-best moves.
	if len(a) > 0 {
		// best := a[0].score
		// 	if best > 0 {
		// n := 0
		// for ; n < len(a); n++ {
		// 	if a[n].score-best < 0 {
		// 		break
		// 	}
		// }
		// a = a[:n]
		// 	} else {
		// n := len(a) - 1
		// for ; n > 0; n-- {
		// 	if best-a[n].score < 100 {
		// 		break
		// 	}
		// }
		// a = a[:n]
		// 	}
		// if len(a) > 20 {
		// 	a = a[:20]
		// }
	}

	return a
}