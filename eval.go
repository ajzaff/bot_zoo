package zoo

const (
	Inf Value = 20000
	Win Value = 10000
)

type Value int16

func (v Value) Infinite() bool {
	return v >= Inf || v <= -Inf
}

func (v Value) Terminal() bool {
	return v >= Win || v <= -Win
}

func (v Value) Winning() bool {
	return v >= Win
}

func (v Value) Losing() bool {
	return v <= -Win
}

var pieceValue = []Value{
	0,
	0,
	200,  // Cat
	300,  // Dog
	500,  // Horse
	800,  // Camel
	1300, // Elephant
}

var rabbitMaterialValue = []Value{
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

var hostageValue = []Value{
	0,
	5,  // Rabbit
	8,  // Dog
	20, // Horse
	50, // Camel
	0,
}

func (p *Pos) hostageScore(side Color) (value Value) {
	(p.frozen[side] & p.presence[side.Opposite()]).Each(func(b Bitboard) {
		if t := p.board[b.Square()]; p.frozenB(t, b) {
			value += hostageValue[t&decolorMask]
		}
	})
	return value
}

func (p *Pos) positionScore(side Color) (value Value) {
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
			value += ps[b.Square()]
		}
	}
	return value
}

func (p *Pos) valueOf(side Color) (value Value) {
	if v := p.bitboards[GRabbit.MakeColor(side)].Count(); v <= 8 {
		value += rabbitMaterialValue[v]
	} else {
		value += rabbitMaterialValue[8] + Value(v) - 8
	}
	for s := GCat; s <= GElephant; s++ {
		value += pieceValue[s] * Value(p.bitboards[s.MakeColor(side)].Count())
	}
	value += p.hostageScore(side)
	value += p.positionScore(side)
	return value
}

func (p *Pos) terminalGoalValue() Value {
	myGoal, theirGoal := ^NotRank8, ^NotRank1
	if p.side != Gold {
		myGoal, theirGoal = theirGoal, myGoal
	}
	if p.bitboards[GRabbit.MakeColor(p.side)]&myGoal != 0 {
		return Win
	}
	if p.bitboards[GRabbit.MakeColor(p.side.Opposite())]&theirGoal != 0 {
		return -Win
	}
	return 0
}

func (p *Pos) terminalEliminationValue() Value {
	if p.bitboards[GRabbit.MakeColor(p.side.Opposite())] == 0 {
		return Win
	}
	if p.bitboards[GRabbit.MakeColor(p.side)] == 0 {
		return -Win
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

func (p *Pos) terminalImmobilizedValue() Value {
	if p.immobilized(p.side.Opposite()) {
		return Win
	}
	if p.immobilized(p.side) {
		return -Win
	}
	return 0
}

func (p *Pos) terminalValue() Value {
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

func (p *Pos) Value() Value {
	if v := p.terminalValue(); v != 0 {
		return v
	}
	return p.valueOf(p.side) - p.valueOf(p.side.Opposite())
}
