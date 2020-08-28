package zoo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var errAEIQuit = errors.New("quit")

type AEI struct {
	engine       *Engine
	id           []string
	protoVersion string
	w            io.Writer
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
		a.engine.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "setoption"):
		return fmt.Errorf("not implemented")
	case strings.HasPrefix(text, "makemove"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			return fmt.Errorf("expected steps matching /%s/", stepPattern)
		}
		move, err := ParseMove(parts[1])
		if err != nil {
			return err
		}
		pos, err := a.engine.Pos().Move(move, true)
		if err != nil {
			return err
		}
		a.engine.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "go"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			move := a.engine.Pos().RandomMove()
			if len(move) == 0 {
				a.Logf("no moves")
				return nil
			}
			a.writef("bestmove %s\n", MoveString(move))
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

func (a *AEI) writef(format string, as ...interface{}) {
	a.w.Write([]byte(fmt.Sprintf(format, as...)))
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
	a.writef(fmt.Sprintf("log %s", format), as...)
}

func (a *AEI) Run(w io.Writer, r io.Reader) error {
	sc := bufio.NewScanner(r)
	a.w = w
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
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
