package zoo

import "sort"

var terminalEval = 100000

var pieceValue = []int{
	0,
	0,
	200,  // Cat
	300,  // Dog
	500,  // Horse
	800,  // Camel
	1300, // Elephant
}

var rabbitValue = []int{
	0,
	12200,
	12900,
	13500,
	14000,
	14400,
	14700,
	14900,
	15000,
}

var mobilityScore = []int{
	-800,
	-600,
	-400,
	-200,
	-180,
	-160,
	-140,
	-120,
	-100,
	-80,
	-60,
	-40,
	-30,
	-20,
	-10,
	0,
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
	score += p.mobilityScore(side)
	return score
}

func (p *Pos) mobilityScore(side Color) (score int) {
	var count int
	b := p.Presence[side]
	for b > 0 {
		atB := b & -b
		if !p.frozenB(atB) {
			count++
		}
		b &= ^atB
	}
	if count >= 16 {
		count = 15
	}
	return mobilityScore[count]
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

var moveLenPenalty = []int{
	-terminalEval,
	-30,
	-20,
	-10,
	0,
	0,
	0,
	0,
	0,
}

func moveLengthPenalty(n int) int {
	if n < 8 {
		return moveLenPenalty[n]
	}
	return 0
}

func (e *Engine) SortMoves(p *Pos, moves [][]Step) (scores []int) {
	e.r.Shuffle(len(moves), func(i, j int) {
		moves[i], moves[j] = moves[j], moves[i]
	})
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
		score := -t.Score()
		score += moveLengthPenalty(len(move))
		a.scores[i] = score
	}
	sort.Sort(a)
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
	return a.scores[i] > a.scores[j]
}
