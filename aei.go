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
	case text == "isready":
		a.write("readyok")
	case text == "newgame":
		pos, _ := ParseShortPosition(PosEmpty)
		a.engine.SetPos(pos)
	case text == "stop":
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
	case strings.HasPrefix(text, "setoption"):
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
	case strings.HasPrefix(text, "go"):
	case strings.HasPrefix(text, "zzz_"):
		return a.handleZoo(text)
	default:
		return fmt.Errorf("unsupported command: %q", text)
	}
	return nil
}

func (a *AEI) write(format string, as ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = fmt.Sprint(format, "\n")
	}
	a.w.Write([]byte(fmt.Sprintf(format, as...)))
}

func (a *AEI) writePreamble() {
	a.write("protocol-version %s", a.protoVersion)
	for _, id := range a.id {
		a.write("id %s", id)
	}
	a.write("aeiok")
}

func (a *AEI) Logf(format string, as ...interface{}) {
	a.write(fmt.Sprintf("log %s", format), as...)
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
