package zoo

import "sort"

var terminalValue = 122

var pieceValue = []int{
	0,
	0,
	2,  // Cat
	3,  // Dog
	5,  // Horse
	8,  // Camel
	13, // Elephant
}

var rabbitValue = []int{
	0,
	terminalValue,
	129,
	135,
	140,
	144,
	147,
	149,
	150,
}

func (p *Pos) Score() (score int) {
	if v := p.Bitboards[GRabbit.MakeColor(p.Side)].Count(); v <= 8 {
		score += rabbitValue[v]
	} else {
		score += rabbitValue[8] + v - 8
	}
	for s := GCat; s <= GElephant; s++ {
		score += pieceValue[s] * p.Bitboards[s.MakeColor(p.Side)].Count()
	}
	return score
}

func (e *Engine) Sort(steps []Step) (scores []int) {
	a := byScore{
		steps:  steps,
		scores: make([]int, len(steps)),
	}
	for i, step := range steps {
		t, _, err := e.Pos().Step(step)
		if err != nil {
			a.scores[i] = -terminalValue
			continue
		}
		a.scores[i] = t.Score()
	}
	sort.Stable(sort.Reverse(a))
	return a.scores
}

type byScore struct {
	scores []int
	steps  []Step
}

func (a byScore) Len() int { return len(a.steps) }
func (a byScore) Swap(i, j int) {
	a.scores[i], a.scores[j] = a.scores[j], a.scores[i]
	a.steps[i], a.steps[j] = a.steps[j], a.steps[i]
}
func (a byScore) Less(i, j int) bool {
	return a.scores[i] < a.scores[j]
}
