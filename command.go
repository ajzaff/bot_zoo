package zoo

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ExecuteCommand parses AEI command and executes the handler or returns an error.
func (e *Engine) ExecuteCommand(s string) error {
	s = strings.TrimSpace(s)

	if e.LogProtocolTraffic {
		e.Debugf("> %s", s)
	}

	i := strings.IndexByte(s, ' ')
	if i == -1 {
		i = len(s)
	}
	if i == 0 {
		return nil
	}
	command := s[:i]

	handler := globalAEIHandlers[command]
	if handler == nil {
		return fmt.Errorf("unrecognized command: %s", command)
	}
	args := strings.TrimSpace(s[i:])
	if err := handler(e, args); err != nil {
		return err
	}
	return nil
}

var errAEIQuit = errors.New("quit")

// IsQuit tests whether the error message matches the AEI "quit" flag error value.
// The error is returned by ExecuteCommand after receiving the "quit" AEI command.
func IsQuit(err error) bool {
	return err == errAEIQuit
}

var globalAEIHandlers = make(map[string]func(e *Engine, args string) error)

// RegisterAEIHandler registers a handler to be invoked upon calling ExecuteCommand with the given command.
func RegisterAEIHandler(command string, handler func(e *Engine, args string) error) {
	globalAEIHandlers[command] = handler
}

func extendedHandler(handler func(e *Engine, args string) error) func(e *Engine, args string) error {
	return handler
}

func init() {
	RegisterAEIHandler("aei", func(e *Engine, args string) error {
		e.Outputf("protocol-version %s", e.ProtoVersion)
		e.Outputf("id name %s", e.BotName)
		e.Outputf("id version %s", e.BotVersion)
		e.Outputf("id author %s", e.BotAuthor)
		e.Outputf("aeiok")
		return nil
	})
	RegisterAEIHandler("isready", func(e *Engine, args string) error {
		e.Outputf("readyok")
		return nil
	})
	RegisterAEIHandler("newgame", func(e *Engine, args string) error {
		e.NewGame()
		return nil
	})
	RegisterAEIHandler("stop", func(e *Engine, args string) error {
		e.Stop()
		return nil
	})
	RegisterAEIHandler("quit", func(e *Engine, args string) error {
		return errAEIQuit
	})
	RegisterAEIHandler("setposition", func(e *Engine, args string) error {
		p, err := ParseShortPosition(args)
		if err != nil {
			return err
		}
		e.Pos = p
		return nil
	})
	RegisterAEIHandler("setoption", func(e *Engine, args string) error {
		return e.ExecuteSetOption(args)
	})
	RegisterAEIHandler("makemove", func(e *Engine, args string) error {
		e.Stop()
		move, err := ParseMove(args)
		if err != nil {
			return err
		}
		e.Move(move)
		return nil
	})
	RegisterAEIHandler("go", func(e *Engine, args string) error {
		switch args {
		case "":
			e.Go()
		case "ponder":
			e.GoPonder()
		case "infinite":
			e.GoInfinite()
		default:
			return fmt.Errorf("unsupported go command: %s", args)
		}
		return nil
	})

	// Extended AEI handlers:

	RegisterAEIHandler("new", extendedHandler(func(e *Engine, args string) error {
		e.NewGame()
		p, err := ParseShortPosition(`g [rrrrrrrrhdcemcdh                                HDCMECDHRRRRRRRR]`)
		if err != nil {
			return err
		}
		e.Pos = p
		return nil
	}))
	RegisterAEIHandler("unmove", extendedHandler(func(e *Engine, args string) error {
		e.Unmove()
		return nil
	}))
	RegisterAEIHandler("hash", extendedHandler(func(e *Engine, args string) error {
		e.Logf("%d", e.Hash())
		return nil
	}))
	RegisterAEIHandler("hashafter", extendedHandler(func(e *Engine, args string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing step")
		}
		step, err := ParseStep(args)
		if err != nil {
			return err
		}
		e.Debugf("%d", e.HashAfter(step))
		return nil
	}))
	RegisterAEIHandler("steps", extendedHandler(func(e *Engine, args string) error {
		var stepList StepList
		stepList.Generate(e.Pos)
		for i := 0; i < stepList.Len(); i++ {
			step := stepList.At(i)
			illegalStr := ""
			if !e.Legal(step.Step) {
				illegalStr = " (illegal)"
			}
			if cap := step.Step.DebugCaptureContext(e.Pos); cap != 0 {
				e.Logf("[%f] %s %s%s", step.Value, step.Step, cap, illegalStr)
			} else {
				e.Logf("[%f] %s%s", step.Value, step.Step, illegalStr)
			}
		}
		return nil
	}))
	RegisterAEIHandler("legal", extendedHandler(func(e *Engine, args string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing step")
		}
		step, err := ParseStep(args)
		if err != nil {
			return err
		}
		if e.Legal(step) {
			e.Logf("legal")
		} else {
			e.Logf("illegal")
		}
		return nil
	}))
	RegisterAEIHandler("place", extendedHandler(func(e *Engine, args string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing setup")
		}
		step, err := ParseStep(args)
		if err != nil {
			return err
		}
		e.Place(step.Piece(), step.Dest())
		return nil
	}))
	RegisterAEIHandler("remove", extendedHandler(func(e *Engine, args string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing square")
		}
		i, err := ParseSquare(args)
		if err != nil {
			return err
		}
		e.Remove(e.At(i), i)
		return nil
	}))
	RegisterAEIHandler("step", extendedHandler(func(e *Engine, args string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing step")
		}
		step, err := ParseStep(args)
		if err != nil {
			return err
		}
		if !e.Legal(step) {
			return fmt.Errorf("illegal step")
		}
		e.Step(step)
		return nil
	}))
	RegisterAEIHandler("unstep", extendedHandler(func(e *Engine, args string) error {
		e.Unstep()
		return nil
	}))
	RegisterAEIHandler("pass", extendedHandler(func(e *Engine, args string) error {
		e.Pass()
		return nil
	}))
	RegisterAEIHandler("unpass", extendedHandler(func(e *Engine, args string) error {
		e.Unpass()
		return nil
	}))
	RegisterAEIHandler("movelist", extendedHandler(func(e *Engine, args string) error {
		e.Debugf(e.moves.String())
		return nil
	}))
	RegisterAEIHandler("eval", extendedHandler(func(e *Engine, args string) error {
		e.Logf("%f", e.Terminal())
		return nil
	}))
	RegisterAEIHandler("random", extendedHandler(func(e *Engine, args string) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		e.RandomSetup(r)
		return nil
	}))
	RegisterAEIHandler("playbatch", extendedHandler(func(e *Engine, args string) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		e.RandomSetup(r)
		for n := 0; ; {
			e.GoWait()
			if v := e.Terminal(); v != 0 {
				n++
				e.Debugf("%s", e.Pos.String())
				if c := e.Side(); v == 1 {
					e.Debugf("%c won game %d of %d", c.Byte(), n, e.PlayBatchGames)
				} else {
					e.Debugf("%c lost game %d of %d", c.Byte(), n, e.PlayBatchGames)
				}
				e.NewGame()
				if n == e.PlayBatchGames {
					break
				}
				continue
			}
			e.Move(e.bestMove)
		}
		e.batchWriter.Flush()
		return nil
	}))
	RegisterAEIHandler("options", extendedHandler(func(e *Engine, args string) error {
		e.Options.Range(func(name string, value interface{}) {
			e.Debugf("%v=%v", name, value)
		})
		return nil
	}))
	RegisterAEIHandler("debug", extendedHandler(func(e *Engine, args string) error {
		e.Debug()
		return nil
	}))
	RegisterAEIHandler("print", extendedHandler(func(e *Engine, args string) error {
		var b Bitboard
		switch args {
		case "":
			e.Logf(e.Pos.String())
			return nil
		case "weaker":
			for t := GRabbit; t <= GElephant; t++ {
				e.Logf(string(t.Byte()))
				e.Logf(e.weaker[t].String())
			}
			return nil
		case "stronger":
			for t := GRabbit; t <= GElephant; t++ {
				e.Logf(string(t.Byte()))
				e.Logf(e.stronger[t].String())
			}
			return nil
		case "tg":
			b = e.touching[Gold]
		case "ts":
			b = e.touching[Silver]
		case "dg":
			b = e.dominating[Gold]
		case "ds":
			b = e.dominating[Silver]
		case "fg":
			b = e.frozen[Gold]
		case "fs":
			b = e.frozen[Silver]
		case "g":
			b = e.presence[Gold]
		case "s":
			b = e.presence[Silver]
		case "short":
			e.Logf(e.ShortString())
			return nil
		default:
			p, err := ParsePiece(args[0])
			if err != nil {
				return fmt.Errorf("printing piece bitboard: %v", err)
			}
			b = e.bitboards[p]
		}
		e.Logf(b.String())
		return nil
	}))
}
