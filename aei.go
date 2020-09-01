package zoo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

var errAEIQuit = errors.New("quit")

type AEI struct {
	engine       *Engine
	id           []string
	protoVersion string
	w            io.Writer
	verbose      bool
}

func NewAEI(engine *Engine) *AEI {
	return &AEI{
		engine: engine,
		id: []string{
			"name zoo",
			"author Alan Zaffetti",
			"version 0",
		},
		protoVersion: "1",
	}
}

func (a *AEI) SetVerbose(verbose bool) {
	a.verbose = verbose
}

func (a *AEI) handle(text string) error {
	switch {
	case text == "aei":
		a.writePreamble()
		return nil
	case text == "isready":
		a.writef("readyok\n")
		return nil
	case text == "newgame":
		pos, _ := ParseShortPosition(PosEmpty)
		a.engine.SetPos(pos)
		return nil
	case text == "stop":
		return fmt.Errorf("not implemented")
	case text == "quit":
		return errAEIQuit
	case strings.HasPrefix(text, "setposition"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			return fmt.Errorf("expected position matching /%s/", shortPosPattern)
		}
		pos, err := ParseShortPosition(parts[1])
		if err != nil {
			return err
		}
		pos.moveNum = 2
		a.engine.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "setoption"):
		return fmt.Errorf("not implemented")
	case strings.HasPrefix(text, "makemove"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			return fmt.Errorf("expected steps")
		}
		parts[1] = strings.TrimSpace(parts[1])
		move, err := parseMove(parts[1])
		if err != nil {
			return err
		}
		if err := a.engine.Pos().Move(move); err != nil {
			return err
		}
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case strings.HasPrefix(text, "go"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			best := a.engine.SearchFixedDepth(4)
			if len(best.Move) == 0 {
				a.Logf("no moves")
				return nil
			}
			a.writef("bestmove %s\n", a.engine.Pos().MoveString(best.Move))
			a.writef("info score %d\n", best.Score)
			a.writef("info nodes %d\n", best.Nodes)
			a.writef("info pv %s\n", a.engine.Pos().MoveString(best.PV))
			a.writef("info time %d\n", int(best.Time.Seconds()))
			a.writef("info curmovenumber %d\n", a.engine.Pos().moveNum)
			return nil
		}
		switch cmd := parts[1]; cmd {
		case "ponder":
			return fmt.Errorf("not implemented")
		default:
			return fmt.Errorf("unsupported go command: %q", cmd)
		}
	case strings.HasPrefix(text, "zzz_"):
		return a.handleZoo(text)
	default:
		return fmt.Errorf("unsupported command: %q", text)
	}
}

func (a *AEI) verboseLog(send bool, s string) {
	if !a.verbose {
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

func (a *AEI) writef(format string, as ...interface{}) {
	s := fmt.Sprintf(format, as...)
	a.w.Write([]byte(s))
	a.verboseLog(true, s)
}

func (a *AEI) writePreamble() {
	a.writef("protocol-version %s\n", a.protoVersion)
	for _, id := range a.id {
		a.writef("id %s\n", id)
	}
	a.writef("aeiok\n")
}

func (a *AEI) Logf(format string, as ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = fmt.Sprint(format, "\n")
	}
	s := fmt.Sprintf(format, as...)
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		a.writef("log %s\n", sc.Text())
	}
}

func (a *AEI) Run(w io.Writer, r io.Reader) error {
	sc := bufio.NewScanner(r)
	a.w = w
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		a.verboseLog(false, text)
		if err := a.handle(text); err != nil {
			if err == errAEIQuit {
				return nil
			}
			a.Logf("ERROR: %v", err)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}
