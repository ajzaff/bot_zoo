package zoo

import (
	"os"
	"sync"
)

const transposeTableSize = 2000000

type Engine struct {
	timeControl TimeControl
	timeInfo    *TimeInfo

	*AEI

	searchInfo *searchInfo

	best       searchResult
	resultChan chan searchResult

	p *Pos

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

	lastPonder bool

	// rootOrderNoise applied during search to root moves.
	// works in conjunction with concurrency to take
	// advantage of the increased randomness.
	rootOrderNoise float64

	// Null move depth reduction factor R.
	// TODO(ajzaff): Use an adaptive value of R.
	nullMoveR int

	// concurrency setting of Lazy-SMP search in number of goroutines.
	concurrency int
	wg          sync.WaitGroup

	table    *Table
	useTable bool

	// semi-atomic
	stopping int32
	running  int32
}

func NewEngine(seed int64) *Engine {
	e := &Engine{
		timeControl:    makeTimeControl(),
		p:              NewEmptyPosition(),
		minDepth:       12,
		concurrency:    4,
		rootOrderNoise: 200,
		nullMoveR:      4,
		table:          NewTable(transposeTableSize),
		useTable:       true,
	}
	e.AEI = NewAEI(e, nil, os.Stdout)
	return e
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
