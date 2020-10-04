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

func NewEngine(settings *EngineSettings, aeiSettings *AEISettings) (*Engine, error) {
	e := &Engine{
		EngineSettings: settings,
		AEISettings:    aeiSettings,
		Options:        newOptions(),
		timeControl:    makeTimeControl(),
		Pos:            NewEmptyPosition(),
		log:            log.New(os.Stdout, "log ", 0),
		out:            log.New(os.Stdout, "", 0),
		debug:          log.New(os.Stderr, "", 0),
	}
	if settings.UseTFRecordWriter {
		e.batchWriter = NewBatchWriter(settings.TFRecordEpoch, settings.TFRecordBatchSize)
	}
	if err := e.EngineSettings.Options.Execute(e.Options); err != nil {
		return nil, err
	}
	e.searchState.Reset()
	return e, nil
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

// GoWait starts the search routine and waits for it to finish.
func (e *Engine) GoWait() {
	if atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		defer func() {
			if r := recover(); r != nil {
				panic(fmt.Sprintf("SEARCH_ERROR recovered: %v", r))
			}
		}()
		e.searchRoot(false)
	}
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
		e.wg.Wait()
		e.running = 0
		e.stopping = 0
	}
}

// Debug engine info.
func (e *Engine) Debug() {
	e.Debugf(e.Pos.String())
	e.Debugf("short=%s", e.ShortString())
	e.Debugf("hash=%v", e.Hash())

	src, piece, ok := e.Push()
	e.Debugf("can_pass=%v", e.CanPass())
	e.Debugf("src=%v", src)
	e.Debugf("piece=%c", piece.Byte())
	e.Debugf("push=%v", ok)
	e.debugStack(e.debug)
	e.threefold.Debug(e.debug)
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
