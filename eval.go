package zoo

const terminalEval = 100000

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
var positionValue = [][]int{{}, { // Rabbit
	999, 999, 999, 999, 999, 999, 999, 999,
	10, 10, 5, 10, 10, 5, 10, 10,
	1, -2, -10, -5, -5, -10, -2, 1,
	1, 0, -5, 5, -5, -5, 0, 1,
	1, 0, 0, -5, -5, 0, 0, 1,
	1, -1, -5, -5, -5, -5, 0, 1,
	1, 0, 0, -5, -5, 0, 0, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
}, { // Cat
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, -8, -10, -8, -8, -10, -8, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -2, 0, 0, -2, 0, 0,
	5, 8, 10, 5, 5, 10, 8, 10,
	5, 10, 10, 5, 5, 10, 10, 10,
}, { // Dog
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, -8, -10, -8, -8, -10, -8, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 10, 10, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // Horse
	-13, -8, -8, -8, -8, -8, -8, -13,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 5, 10, 2, 2, 10, 5, -8,
	-8, 5, 5, 2, 2, 5, 5, -8,
	-8, 5, 0, 0, 5, 0, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-13, -8, -8, -8, -8, -8, -8, -13,
}, { // Camel
	-13, -8, -8, -8, -8, -8, -8, -13,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 5, 10, 2, 2, 10, 5, -8,
	-8, 5, 5, 2, 2, 5, 5, -8,
	-8, 5, 0, 5, 5, 0, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-13, -8, -8, -8, -8, -8, -8, -13,
}, { // Elephant
	-22, -13, -13, -13, -13, -13, -13, -22,
	-13, 0, 0, 0, 0, 0, 0, -13,
	-13, 5, -8, 10, 10, -8, 5, -13,
	-13, 5, 10, 10, 10, 10, 5, -13,
	-13, 5, 5, 10, 10, 5, 5, -13,
	-13, 5, 0, 5, 5, 0, 5, -13,
	-13, 0, -1, 0, 0, -1, 0, -13,
	-22, -13, -13, -13, -13, -13, -13, -22,
}}

func isInfinite(score int) bool {
	return score >= inf || score <= -inf
}

func isTerminal(score int) bool {
	return score >= terminalEval || score <= -terminalEval
}

func (p *Pos) mobilityScore(side Color) (score int) {
	var count int
	for b := p.presence[side]; b > 0; b &= (b - 1) {
		if !p.Frozen(b.Square()) {
			count++
		}
	}
	if count >= 16 {
		count = 15
	}
	return mobilityScore[count]
}

func (p *Pos) positionScore(side Color) (score int) {
	c := 7
	m := -1
	if side != Gold {
		c = 0
		m = 1
	}
	for _, t := range []Piece{
		GRabbit.MakeColor(side),
		GCat.MakeColor(side),
		GDog.MakeColor(side),
		GHorse.MakeColor(side),
		GCamel.MakeColor(side),
		GElephant.MakeColor(side),
	} {
		ps := positionValue[t&decolorMask]
		for b := p.bitboards[t]; b > 0; b &= b - 1 {
			at := b.Square()
			score += ps[8*(c+m*int(at)/8)+c+m*(int(at)%8)]
		}
	}
	return score
}

func (p *Pos) score(side Color) (score int) {
	if v := p.bitboards[GRabbit.MakeColor(side)].Count(); v <= 8 {
		score += rabbitMaterialValue[v]
	} else {
		score += rabbitMaterialValue[8] + v - 8
	}
	for s := GCat; s <= GElephant; s++ {
		score += pieceValue[s] * p.bitboards[s.MakeColor(side)].Count()
	}
	score += p.mobilityScore(side)
	score += p.positionScore(side)
	return score
}

func (p *Pos) terminalGoalValue() int {
	myGoal, theirGoal := ^NotRank8, ^NotRank1
	if p.side != Gold {
		myGoal, theirGoal = theirGoal, myGoal
	}
	if p.bitboards[GRabbit.MakeColor(p.side)]&myGoal != 0 {
		return terminalEval
	}
	if p.bitboards[GRabbit.MakeColor(p.side.Opposite())]&theirGoal != 0 {
		return -terminalEval
	}
	return 0
}

func (p *Pos) terminalEliminationValue() int {
	if p.bitboards[GRabbit.MakeColor(p.side.Opposite())] == 0 {
		return terminalEval
	}
	if p.bitboards[GRabbit.MakeColor(p.side)] == 0 {
		return -terminalEval
	}
	return 0
}

func (p *Pos) immobilized(c Color) bool {
	for b := p.presence[c]; b > 0; b &= (b - 1) {
		if !p.Frozen(b.Square()) {
			return false
		}
	}
	return true
}

func (p *Pos) terminalImmobilizedValue() int {
	if p.immobilized(p.side.Opposite()) {
		return terminalEval
	}
	if p.immobilized(p.side) {
		return -terminalEval
	}
	return 0
}

func (p *Pos) terminalValue() int {
	if v := p.terminalGoalValue(); v != 0 {
		return v
	}
	if v := p.terminalEliminationValue(); v != 0 {
		return v
	}
	if v := p.terminalImmobilizedValue(); v != 0 {
		return v
	}
	return 0
}

func (p *Pos) Score() int {
	if v := p.terminalValue(); v != 0 {
		return v
	}
	return p.score(p.side) - p.score(p.side.Opposite())
}
