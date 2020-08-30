package zoo

import "sort"

func (e *Engine) sortMoves(p *Pos, moves [][]Step) []scoredMove {
	a := make([]scoredMove, len(moves))
	for i := range a {
		a[i].move = moves[i]
		a[i].score = -terminalEval
	}

	for i, move := range moves {
		t, _, err := p.Move(move, false)
		if err != nil {
			continue
		}

		// Step 1a: Table lookup.
		if e.useTable {
			if entry, ok := e.table.ProbeDepth(t.ZHash, 0); ok {
				score := entry.Value
				if entry.Bound == ExactBound {
					a[i].score = inf + score
				}
				continue
			}
		}

		// Step 1b: Fallback to eval score.
		a[i].score = -t.Score()
	}
	sort.Sort(byLen(a))
	sort.Stable(byScore(a))

	// Experimental pruning of non-best moves.
	{
		best := a[0].score
		if best > 0 {
			n := 0
			for ; n < len(a); n++ {
				if a[n].score-best < 0 {
					break
				}
			}
			a = a[:n]
		} else {
			n := len(a) - 1
			for ; n > 0; n-- {
				if a[n].score-best < 100 {
					break
				}
			}
			a = a[:n]
		}
	}

	return a
}

type scoredMove struct {
	score int
	move  []Step
}

type byLen []scoredMove

func (a byLen) Len() int           { return len(a) }
func (a byLen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLen) Less(i, j int) bool { return len(a[i].move) > len(a[j].move) }

type byScore []scoredMove

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].score > a[j].score }
