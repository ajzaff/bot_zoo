package zoo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

// Engine implements game control structures around Pos and keeps track of game state.
type Engine struct {
	*EngineSettings
	*AEISettings

	*Options

	log   *log.Logger
	out   *log.Logger
	debug *log.Logger

	*Pos

	timeControl TimeControl
	timeInfo    *TimeInfo

	searchState
}

func NewEngine() *Engine {
	e := &Engine{
		AEISettings:    &AEISettings{},
		EngineSettings: &EngineSettings{},
		Options:        newOptions(),
		timeControl:    makeTimeControl(),
		Pos:            NewEmptyPosition(),
		log:            log.New(os.Stdout, "log ", 0),
		out:            log.New(os.Stdout, "", 0),
		debug:          log.New(os.Stderr, "", 0),
	}
	e.searchState.Reset()
	return e
}

func (e *Engine) NewGame() {
	e.Pos = NewEmptyPosition()
	e.tt.Clear()
	e.timeInfo = e.timeControl.newTimeInfo()
}

func (e *Engine) startNow(ponder bool) {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("SEARCH_ERROR recovered: %v", r))
		}
	}()
	go e.searchRoot(ponder)
}

// Go starts the search routine in a new goroutine.
func (e *Engine) Go() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		e.startNow(false)
	}
}

// GoPonder starts the ponder search in a new goroutine.
func (e *Engine) GoPonder() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		e.startNow(true)
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

// Logf logs the formatted message to the configured log writer.
// This is used for all logging as well as AEI protocol logging
// which requires the prefix be set to "log" on each line.
func (e *Engine) Logf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		e.log.Println(sc.Text())
	}
}

// Outputf outputs the formatted message to the configured output log.
// This should be used only for AEI protocol messages.
func (e *Engine) Outputf(format string, a ...interface{}) {
	if e.LogVerbosePosition {
		e.Debugf(e.Pos.String())
		e.Debugf(e.Pos.ShortString())
	}
	s := fmt.Sprintf(format, a...)
	if e.LogProtocolTraffic {
		e.Debugf("< %s", s)
	}
	e.out.Print(s)
}

// Debugf logs the formatted message to stderr.
func (e *Engine) Debugf(format string, a ...interface{}) {
	e.debug.Printf(format, a...)
}
