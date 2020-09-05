package zoo

import (
	"math/rand"
)

const transposeTableSize = 2000000

type Engine struct {
	timeControl TimeControl
	timeInfo    *TimeInfo

	searchInfo *SearchInfo

	p *Pos
	r *rand.Rand

	// depth != 0 implies fixed depth.
	depth int

	table    *Table
	useTable bool

	stopping int32
	running  int32 // atomic
}

func NewEngine(seed int64) *Engine {
	return &Engine{
		p:        NewEmptyPosition(),
		r:        rand.New(rand.NewSource(seed)),
		table:    NewTable(transposeTableSize),
		useTable: true,
	}
}

func (e *Engine) NewGame() {
	pos := NewEmptyPosition()
	e.SetPos(pos)
	e.table.Clear()
	e.timeInfo = e.timeControl.newTimeInfo()
}

func (e *Engine) Pos() *Pos {
	return e.p
}

func (e *Engine) SetPos(p *Pos) {
	*e.p = *p
}
