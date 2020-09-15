package zoo

const (
	Inf Value = 999
	Win Value = 1
)

type Value float64

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
	return 0
}
