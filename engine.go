package zoo

import (
	"math/rand"
	"sync"
)

const transposeTableSize = 2000000

type Engine struct {
	TimeLimits TimeLimits
	TimeInfo   TimeInfo

	p *Pos
	r *rand.Rand

	// depth != 0 implies fixed depth.
	depth int

	table    *Table
	useTable bool

	stopping int32 // atomic
	running  int32 // atomic

	best SearchResult // guarded by mu
	mu   sync.Mutex
}

func NewEngine(seed int64) *Engine {
	return &Engine{
		p:        NewEmptyPosition(),
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
