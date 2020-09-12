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

var hostageScore = []int{
	0,
	5,  // Rabbit
	8,  // Dog
	20, // Horse
	50, // Camel
	0,
}

func isInfinite(score int) bool {
	return score >= inf || score <= -inf
}

func Terminal(score int) bool {
	return score >= terminalEval || score <= -terminalEval
}

func Winning(score int) bool {
	return score >= terminalEval
}

func Losing(score int) bool {
	return score <= -terminalEval
}

func (p *Pos) hostageScore(side Color) (value int) {
	(p.frozen[side] & p.presence[side.Opposite()]).Each(func(b Bitboard) {
		if t := p.board[b.Square()]; p.frozenB(t, b) {
			value += hostageScore[t&decolorMask]
		}
	})
	return value
}

func (p *Pos) positionScore(side Color) (score int) {
	ps := positionValue[side]
	for _, t := range []Piece{
		GRabbit,
		GCat,
		GDog,
		GHorse,
		GCamel,
		GElephant,
	} {
		ps := ps[t]
		for b := p.bitboards[t]; b > 0; b &= b - 1 {
			score += ps[b.Square()]
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
	score += p.hostageScore(side)
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
