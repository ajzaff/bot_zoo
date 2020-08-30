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
var positionValue = [][]int{{ // Empty
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 10, 0, 0, 10, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 10, 0, 0, 10, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // Rabbit
	999, 999, 999, 999, 999, 999, 999, 999,
	199, 199, 50, 199, 199, 50, 199, 199,
	8, -2, -10, -5, -5, -10, -2, 8,
	5, 0, -5, 5, -5, -5, 0, 5,
	2, 0, 0, -5, -5, 0, 0, 2,
	2, 0, -5, -5, -5, -5, 0, 2,
	0, 0, -2, -5, -5, -2, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
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
	-8, -8, -8, -8, -8, -8, -8, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 5, 10, 2, 2, 10, 5, -8,
	-8, 5, 5, 2, 2, 5, 5, -8,
	-8, 5, 0, 0, 5, 0, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, -8, -8, -8, -8, -8, -8, -8,
}, { // Camel
	-8, -8, -8, -8, -8, -8, -8, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 5, 10, 2, 2, 10, 5, -8,
	-8, 5, 5, 2, 2, 5, 5, -8,
	-8, 5, 0, 5, 5, 0, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, -8, -8, -8, -8, -8, -8, -8,
}, { // Elephant
	-8, -8, -8, -8, -8, -8, -8, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 5, 10, 10, 10, 10, 5, -8,
	-8, 5, 5, 10, 10, 5, 5, -8,
	-8, 5, 0, 5, 5, 0, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, -8, -8, -8, -8, -8, -8, -8,
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
	if side != Gold {
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

func (p *Pos) terminalGoalValue() int {
	myGoal, theirGoal := ^NotRank8, ^NotRank1
	if p.Side != Gold {
		myGoal, theirGoal = theirGoal, myGoal
	}
	if p.Bitboards[GRabbit.MakeColor(p.Side)]&myGoal != 0 {
		return terminalEval
	}
	if p.Bitboards[GRabbit.MakeColor(p.Side.Opposite())]&theirGoal != 0 {
		return -terminalEval
	}
	return 0
}

func (p *Pos) terminalEliminationValue() int {
	if p.Bitboards[GRabbit.MakeColor(p.Side.Opposite())] == 0 {
		return terminalEval
	}
	if p.Bitboards[GRabbit.MakeColor(p.Side)] == 0 {
		return -terminalEval
	}
	return 0
}

func (p *Pos) terminalImmobilizedValue() int {
	if p.immobilized(p.Side.Opposite()) {
		return terminalEval
	}
	if p.immobilized(p.Side) {
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
	return p.score(p.Side) - p.score(p.Side.Opposite())
}
