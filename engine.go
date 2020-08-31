package zoo

import "math/rand"

const transposeTableSize = 2000000

type Engine struct {
	p        *Pos
	r        *rand.Rand
	table    *Table
	useTable bool
}

func NewEngine(seed int64) *Engine {
	pos, _ := ParseShortPosition(PosEmpty)
	return &Engine{
		p:        pos,
		r:        rand.New(rand.NewSource(seed)),
		table:    NewTable(transposeTableSize),
		useTable: true,
	}
}

func (e *Engine) Pos() *Pos {
	return e.p
}

func (e *Engine) SetPos(p *Pos) {
	*e.p = *p
}
