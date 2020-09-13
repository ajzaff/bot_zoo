package zoo

const (
	Inf Value = 30000
	Win Value = 20000
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

// rabbitMaterialValue value of each rabbit with i rabbits remaining.
var rabbitMaterialValue = []Value{
	0,
	900,
	1600,
	2200,
	2700,
	3100,
	3400,
	3600,
	3700,
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
		for b := p.Bitboard(t.MakeColor(side)); b > 0; b &= b - 1 {
			value += ps[b.Square()]
		}
	}
	return value
}

func (p *Pos) valueOf(side Color) (value Value) {
	n := p.bitboards[GRabbit.MakeColor(side)].Count()
	if n > 8 {
		n = 8
	}
	value += rabbitMaterialValue[n]
	for t := GCat; t <= GElephant; t++ {
		value += pieceValue[t] * Value(p.Bitboard(t.MakeColor(side)).Count())
	}
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
