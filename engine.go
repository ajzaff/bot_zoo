package zoo

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
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
	if settings.UseDatasetWriter {
		e.batchWriter = NewBatchWriter(settings.DatasetEpoch)
	}
	if err := e.EngineSettings.Options.Execute(e.Options); err != nil {
		return nil, err
	}
	if err := e.searchState.Reset(settings); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) NewGame() {
	e.Pos = NewEmptyPosition()
	e.tt.Clear()
	e.searchState.Reset(e.EngineSettings)
	e.timeInfo = e.timeControl.newTimeInfo()
}

// RandomSetup initializes the game with random setup moves.
func (e *Engine) RandomSetup(r *rand.Rand) {
	e.NewGame()
	var stepList StepList
	for i := 0; i < 32; i++ {
		stepList.Generate(e.Pos)
		j := 0
		for i := 0; i < stepList.Len(); i++ {
			if e.Legal(stepList.At(i).Step) {
				stepList.Swap(i, j)
				j++
			}
		}
		stepList.Truncate(j)
		s := stepList.At(r.Intn(stepList.Len())).Step
		if e.UseDatasetWriter {
			e.batchWriter.WriteExample(e.Pos, &Tree{
				root: &TreeNode{
					children: []*TreeNode{{
						step: s,
						runs: 1,
					}},
				},
			})
		}
		e.Step(s)
		stepList.Truncate(0)
	}
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

// Close closes the engine and all dependencies.
func (e *Engine) Close() (err error) {
	if err1 := e.searchState.model.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
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
	e.tree.Debug(e.debug)
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
