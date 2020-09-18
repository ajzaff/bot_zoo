package zoo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type aei struct {
	id           []string
	protoVersion string
	log          io.Writer
	w            io.Writer
}

func newV1AEI(log, w io.Writer) *aei {
	return &aei{
		log:          log,
		w:            w,
		protoVersion: "1",
		id: []string{
			"name zoo",
			"author Alan Zaffetti",
			"version 0",
		},
	}
}

func (e *Engine) SetLog(w io.Writer) {
	e.aei.log = w
}

func (e *Engine) SetOutput(w io.Writer) {
	e.aei.w = w
}

var errAEIQuit = errors.New("quit")

// IsQuit tests whether the error message matches the AEI "quit" flag error value.
// The error is returned by ExecuteCommand after receiving the "quit" AEI command.
func IsQuit(err error) bool {
	return err == errAEIQuit
}

func (e *Engine) ExecuteCommand(text string) error {
	switch {
	case text == "aei":
		e.writePreamble()
		return nil
	case text == "isready":
		e.writef("readyok\n")
		return nil
	case text == "newgame":
		e.NewGame()
		return nil
	case text == "stop":
		e.Stop()
	case text == "quit":
		return errAEIQuit
	case strings.HasPrefix(text, "setposition "):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			return fmt.Errorf("expected position matching /%s/", shortPosPattern)
		}
		pos, err := ParseShortPosition(parts[1])
		if err != nil {
			return err
		}
		pos.moveNum = 2
		e.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "setoption "):
		// name, value, err := ParseSetOption()
	case strings.HasPrefix(text, "makemove "):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			return fmt.Errorf("expected steps")
		}
		e.Stop()
		parts[1] = strings.TrimSpace(parts[1])
		move, err := ParseMove(parts[1])
		if err != nil {
			return err
		}
		if err := e.Pos().Move(move); err != nil {
			return err
		}
		if e.opts.Verbose {
			e.verbosePos()
		}
		return nil
	case text == "go":
		e.Go()
		return nil
	case strings.HasPrefix(text, "go "):
		parts := strings.SplitN(text, " ", 2)
		switch cmd := strings.TrimSpace(parts[1]); cmd {
		case "ponder":
			e.GoPonder()
		case "infinite":
			e.GoInfinite()
		default:
			return fmt.Errorf("unsupported go command: %q", cmd)
		}
	default:
		return e.ExecuteExtendedCommand(text)
	}
	return nil
}

func (e *Engine) verboseLog(send bool, s string) {
	if !e.opts.Verbose {
		return
	}
	sc := bufio.NewScanner(strings.NewReader(s))
	prefix := ">"
	if send {
		prefix = "<"
	}
	for sc.Scan() {
		log.Println(prefix, sc.Text())
	}
}

func (e *Engine) writef(format string, as ...interface{}) {
	s := fmt.Sprintf(format, as...)
	e.aei.w.Write([]byte(s))
	e.verboseLog(true, s)
}

func (e *Engine) writePreamble() {
	e.writef("protocol-version %s\n", e.aei.protoVersion)
	for _, id := range e.aei.id {
		e.writef("id %s\n", id)
	}
	e.writef("aeiok\n")
}

func (e *Engine) Logf(format string, as ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = fmt.Sprint(format, "\n")
	}
	s := fmt.Sprintf(format, as...)
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		if e.aei.log != nil {
			s := fmt.Sprintf(format, as...)
			e.aei.log.Write([]byte(s))
		}
		e.writef("log %s\n", sc.Text())
	}
}

func (e *Engine) logEval() {
	// e.Logf("eval: %d", e.Pos().Value())
}

func (e *Engine) verbosePos() {
	e.Logf(e.Pos().String())
	e.logEval()
}
