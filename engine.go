package zoo

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Engine struct {
	opts *Options
	aei  *AEISettings
	AEI

	timeControl TimeControl
	timeInfo    *TimeInfo

	// Search result from the last search.
	best       []ExtStep
	resultChan chan ExtStep

	p *Pos

	// ponder implies we will search until we're asked explicitly to stop.
	// We don't set the best move after a ponder.
	// We don't clear the transposition table when we're done.
	// Ponder will stop terminal score is achieved.
	ponder bool

	// concurrency setting of Lazy-SMP search in number of goroutines.
	concurrency int
	wg          sync.WaitGroup

	table    *Table
	useTable bool

	// controls whether to pick the best move or sample.
	pickBest bool

	// semi-atomic
	stopping int32
	running  int32
}

func NewEngine(seed int64) *Engine {
	e := &Engine{
		timeControl: makeTimeControl(),
		p:           NewEmptyPosition(),
		concurrency: 4,
		table:       NewTable(),
		useTable:    true,
	}
	var opts Options
	e.opts = &opts
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

func (e *Engine) startNow() {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("SEARCH_ERROR recovered: %v\n", r))
		}
	}()
	go e.searchRoot()
}

// Go starts the search routine in a new goroutine.
func (e *Engine) Go() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		e.ponder = false
		e.startNow()
	}
}

// GoPonder starts the ponder search in a new goroutine.
func (e *Engine) GoPonder() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		e.ponder = true
		e.startNow()
	}
}

// GoInfinite starts an infinite routine (same as GoPonder).
func (e *Engine) GoInfinite() {
	e.GoPonder()
}

// Stop signals the search to stop immediately.
func (e *Engine) Stop() {
	if atomic.CompareAndSwapInt32(&e.stopping, 0, 1) {
	loop:
		for {
			select {
			case <-e.resultChan:
			default:
				break loop
			}
		}
		e.wg.Wait()
		e.running = 0
		e.stopping = 0
	}
}

func searchRateKNps(nodes int, start time.Time) int64 {
	return int64(float64(nodes) / (float64(time.Now().Sub(start)) / float64(time.Second)) / 1000)
}

func (e *Engine) printSearchInfo(nodes int, depth uint8, start time.Time, best ExtStep) {
	if e.ponder {
		e.Logf("ponder")
	}
	e.writef("info depth %d\n", depth)
	e.writef("info time %d\n", int(time.Now().Sub(start).Seconds()))
	e.writef("info score %d\n", best.Value)
	e.writef("info nodes %d\n", nodes)
	e.Logf("rate %d kN/s", searchRateKNps(nodes, start))
	e.Logf("hashfull %d", e.table.Hashfull())
}
