package zoo

import "sort"

var terminalEval = 14884

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
	122,
	129,
	135,
	140,
	144,
	147,
	149,
	150,
}

func (p *Pos) score(side Color) (score int) {
	if v := p.Bitboards[GRabbit.MakeColor(side)].Count(); v <= 8 {
		score += rabbitValue[v]
	} else {
		score += rabbitValue[8] + v - 8
	}
	for s := GCat; s <= GElephant; s++ {
		score += pieceValue[s] * p.Bitboards[s.MakeColor(side)].Count()
	}
	return score
}

func (p *Pos) Score() int {
	if v := p.Goal(); v == p.Side {
		return terminalEval
	} else if v == p.Side.Opposite() {
		return -terminalEval
	}
	if v := p.Eliminated(); v == p.Side {
		return -terminalEval
	} else if v == p.Side.Opposite() {
		return terminalEval
	}
	if v := p.Immobilized(); v == p.Side {
		return -terminalEval
	} else if v == p.Side.Opposite() {
		return terminalEval
	}
	return p.score(p.Side) - p.score(p.Side.Opposite())
}

func SortMoves(p *Pos, moves [][]Step) (scores []int) {
	a := byScore{
		moves:  moves,
		scores: make([]int, len(moves)),
	}
	for i, move := range moves {
		t, _, err := p.Move(move, false)
		if err != nil {
			a.scores[i] = -terminalEval
			continue
		}
		a.scores[i] = t.Score()
	}
	sort.Stable(sort.Reverse(a))
	return a.scores
}

type byScore struct {
	scores []int
	moves  [][]Step
}

func (a byScore) Len() int { return len(a.moves) }
func (a byScore) Swap(i, j int) {
	a.scores[i], a.scores[j] = a.scores[j], a.scores[i]
	a.moves[i], a.moves[j] = a.moves[j], a.moves[i]
}
func (a byScore) Less(i, j int) bool {
	return a.scores[i] < a.scores[j]
}
