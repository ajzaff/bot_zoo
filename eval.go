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

var rabbitMaterialValue = []int{
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

// Position values are symmetrical and represented as gold side.
// When evaluating Silver, the board must be reflected on Rank-axis.
// They are presented with A1 in the bottom left corner for clarity.
var positionValue = [][]int{{ // Empty
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 50, 0, 0, 50, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 50, 0, 0, 50, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // Rabbit
	999, 999, 999, 999, 999, 999, 999, 999,
	199, 199, 50, 199, 199, 50, 199, 199,
	20, -5, -50, -10, -10, -50, -5, 20,
	10, 0, -10, 10, -10, -10, 0, 10,
	5, 0, 0, -10, -10, 0, 0, 5,
	5, 0, -10, -10, -10, -10, 0, 5,
	0, 0, -5, -10, -10, -5, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // Cat
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -50, 0, 0, -50, 0, 0,
	0, -50, -100, -50, -50, -100, -50, 0,
	0, 0, -50, 0, 0, -50, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -5, 0, 0, -5, 0, 0,
	10, 20, 50, 10, 10, 50, 20, 50,
	10, 50, 50, 10, 10, 50, 50, 50,
}, { // Dog
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -20, 0, 0, -20, 0, 0,
	0, -20, -100, -20, -20, -100, -20, 0,
	0, 0, -20, 0, 0, -20, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 50, 50, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // Horse
	-20, -20, -20, -20, -20, -20, -20, -20,
	-20, 0, 0, 0, 0, 0, 0, -20,
	-20, 10, -20, 50, 50, -20, 10, -20,
	-20, 10, 50, 5, 5, 50, 10, -20,
	-20, 10, 10, 5, 5, 10, 10, -20,
	-20, 10, 0, 0, 10, 0, 10, -20,
	-20, 0, 0, 0, 0, 0, 0, -20,
	-20, -20, -20, -20, -20, -20, -20, -20,
}, { // Camel
	-20, -20, -20, -20, -20, -20, -20, -20,
	-20, 0, 0, 0, 0, 0, 0, -20,
	-20, 10, -20, 50, 50, -20, 10, -20,
	-20, 10, 50, 5, 5, 50, 10, -20,
	-20, 10, 10, 5, 5, 10, 10, -20,
	-20, 10, 0, 10, 10, 0, 10, -20,
	-20, 0, 0, 0, 0, 0, 0, -20,
	-20, -20, -20, -20, -20, -20, -20, -20,
}, { // Elephant
	-20, -20, -20, -20, -20, -20, -20, -20,
	-20, 0, 0, 0, 0, 0, 0, -20,
	-20, 10, -20, 50, 50, -20, 10, -20,
	-20, 10, 50, 50, 50, 50, 10, -20,
	-20, 10, 10, 50, 50, 10, 10, -20,
	-20, 10, 0, 10, 10, 0, 10, -20,
	-20, 0, 0, 0, 0, 0, 0, -20,
	-20, -20, -20, -20, -20, -20, -20, -20,
}}

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

func (p *Pos) positionScore(side Color) (score int) {
	c := 7
	m := -1
	if side == Silver {
		c = 0
		m = 1
	}
	for _, t := range []Piece{
		Empty,
		GRabbit.MakeColor(side),
		GCat.MakeColor(side),
		GDog.MakeColor(side),
		GHorse.MakeColor(side),
		GCamel.MakeColor(side),
		GElephant.MakeColor(side),
	} {
		ps := positionValue[t&decolorMask]
		for b := p.Bitboards[t]; b > 0; b &= b - 1 {
			at := b.Square()
			score += ps[8*(c+m*int(at)/8)+c+m*(int(at)%8)]
		}
	}
	return score
}

func (p *Pos) score(side Color) (score int) {
	if v := p.Bitboards[GRabbit.MakeColor(side)].Count(); v <= 8 {
		score += rabbitMaterialValue[v]
	} else {
		score += rabbitMaterialValue[8] + v - 8
	}
	for s := GCat; s <= GElephant; s++ {
		score += pieceValue[s] * p.Bitboards[s.MakeColor(side)].Count()
	}
	score += p.mobilityScore(side)
	score += p.positionScore(side)
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
