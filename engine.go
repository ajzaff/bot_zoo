package zoo

import "math/rand"

type Engine struct {
	p     *Pos
	r     *rand.Rand
	table Table
}

func NewEngine(seed int64) *Engine {
	pos, _ := ParseShortPosition(PosEmpty)
	return &Engine{
		p:     pos,
		r:     rand.New(rand.NewSource(seed)),
		table: NewTable(),
	}
}

func (e *Engine) Pos() *Pos {
	return e.p
}

func (e *Engine) SetPos(p *Pos) {
	*e.p = *p
}
