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
	// Search won't stop unless a terminal score is achieved.
	fixedDepth int

	// minDepth for time based iterative deepening.
	minDepth int

	// ponder implies we will search until we're asked explicitly to stop.
	// We don't set the best move after a ponder.
	// We don't clear the transposition table when we're done.
	// Ponder will stop terminal score is achieved.
	ponder bool

	table    *Table
	useTable bool

	stopping int32
	running  int32 // atomic
}

func NewEngine(seed int64) *Engine {
	return &Engine{
		p:        NewEmptyPosition(),
		r:        rand.New(rand.NewSource(seed)),
		minDepth: 8,
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
