package zoo

import (
	"errors"
	"fmt"
	"strings"
)

// ExecuteCommand parses AEI command and executes the handler or returns an error.
func (e *Engine) ExecuteCommand(s string) error {
	s = strings.TrimSpace(s)

	if e.LogProtocolTraffic {
		e.Logf("> %s", s)
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
		parts := strings.SplitN(args, " ", 2)
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
	})
	RegisterAEIHandler("setoption", func(e *Engine, args string) error {
		return e.ExecuteSetOption(args)
	})
	RegisterAEIHandler("makemove", func(e *Engine, args string) error {
		parts := strings.SplitN(args, " ", 2)
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
		return nil
	})
	RegisterAEIHandler("go", func(e *Engine, args string) error {
		switch args {
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
		pos, err := ParseShortPosition(posStandard)
		if err != nil {
			panic(err)
		}
		pos.moveNum = 2
		e.SetPos(pos)
		return nil
	}))
	RegisterAEIHandler("unmove", extendedHandler(func(e *Engine, args string) error {
		if err := e.Pos().Unmove(); err != nil {
			return err
		}
		return nil
	}))
	RegisterAEIHandler("hash", extendedHandler(func(e *Engine, args string) error {
		e.Logf("%d", e.Pos().zhash)
		return nil
	}))
	RegisterAEIHandler("depth", extendedHandler(func(e *Engine, args string) error {
		e.Logf("%d", e.Pos().Depth())
		return nil
	}))
	RegisterAEIHandler("step", extendedHandler(func(e *Engine, args string) error {
		parts := strings.SplitN(args, " ", 2)
		if len(parts) < 2 {
			var stepList StepList
			stepList.Generate(e.Pos())
			for i := 0; i < stepList.Len(); i++ {
				step := stepList.At(i)
				e.Logf("[%f] %s", step.Value, step.Step)
			}
			return nil
		}
		step, err := ParseStep(parts[1])
		if err != nil {
			return err
		}
		if err := e.Pos().Step(step); err != nil {
			return err
		}
		return nil
	}))
	RegisterAEIHandler("unstep", extendedHandler(func(e *Engine, args string) error {
		if err := e.Pos().Unstep(); err != nil {
			return err
		}
		return nil
	}))
	RegisterAEIHandler("pass", extendedHandler(func(e *Engine, args string) error {
		e.Pos().Pass()
		return nil
	}))
	RegisterAEIHandler("unpass", extendedHandler(func(e *Engine, args string) error {
		if err := e.Pos().Unpass(); err != nil {
			return err
		}
		return nil
	}))
	RegisterAEIHandler("eval", extendedHandler(func(e *Engine, args string) error {
		// TODO(ajzaff): Output current eval.
		e.Logf("0")
		return nil
	}))
	RegisterAEIHandler("print", extendedHandler(func(e *Engine, args string) error {
		parts := strings.SplitN(args, " ", 2)
		if len(parts) < 2 {
			e.Logf(e.Pos().String())
			return nil
		}
		var b Bitboard
		switch parts[1] {
		case "weaker":
			b = e.Pos().touching[Gold]
			for t := GRabbit; t <= GElephant; t++ {
				b = e.Pos().weaker[t]
				e.Logf(string(t.Byte()))
				e.Logf(b.String())
			}
			return nil
		case "stronger":
			for t := GRabbit; t <= GElephant; t++ {
				b = e.Pos().stronger[t]
				e.Logf(string(t.Byte()))
				e.Logf(b.String())
			}
			return nil
		case "tg":
			b = e.Pos().touching[Gold]
		case "ts":
			b = e.Pos().touching[Silver]
		case "dg":
			b = e.Pos().dominating[Gold]
		case "ds":
			b = e.Pos().dominating[Silver]
		case "fg":
			b = e.Pos().frozen[Gold]
		case "fs":
			b = e.Pos().frozen[Silver]
		case "g":
			b = e.Pos().presence[Gold]
		case "s":
			b = e.Pos().presence[Silver]
		case "short":
			e.Logf(e.Pos().ShortString())
			return nil
		default:
			p, err := ParsePiece(parts[1][0])
			if err != nil {
				return fmt.Errorf("printing piece bitboard: %v", err)
			}
			b = e.Pos().bitboards[p]
		}
		return nil
	}))
}
