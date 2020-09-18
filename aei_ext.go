package zoo

import (
	"fmt"
	"strings"
)

func (e *Engine) executeExtendedCommand(text string) error {
	switch {
	case text == "new":
		e.NewGame()
		pos, err := ParseShortPosition(posStandard)
		if err != nil {
			panic(err)
		}
		pos.moveNum = 2
		e.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "unmove"):
		if err := e.Pos().Unmove(); err != nil {
			return err
		}
		if e.opts.Verbose {
			e.verbosePos()
		}
		return nil
	case text == "hash":
		e.Logf("%d", e.Pos().zhash)
		return nil
	case text == "depth":
		e.Logf("%d", e.Pos().Depth())
		return nil
	case strings.HasPrefix(text, "step"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			var stepList StepList
			stepList.Generate(e.Pos())
			for i := 0; i < stepList.Len(); i++ {
				step := stepList.At(i)
				e.Logf("[%d] %s", step.Value, step.Step)
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
		if e.opts.Verbose {
			e.verbosePos()
		}
		return nil
	case strings.HasPrefix(text, "unstep"):
		if err := e.Pos().Unstep(); err != nil {
			return err
		}
		if e.opts.Verbose {
			e.verbosePos()
		}
		return nil
	case text == "pass":
		e.Pos().Pass()
		if e.opts.Verbose {
			e.verbosePos()
		}
		return nil
	case text == "unpass":
		if err := e.Pos().Unpass(); err != nil {
			return err
		}
		if e.opts.Verbose {
			e.verbosePos()
		}
		return nil
	case text == "eval":
		e.logEval()
		return nil
	case strings.HasPrefix(text, "print"):
		parts := strings.SplitN(text, " ", 2)
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
				e.Logf(t.String())
				e.Logf(b.String())
			}
			return nil
		case "stronger":
			for t := GRabbit; t <= GElephant; t++ {
				b = e.Pos().stronger[t]
				e.Logf(t.String())
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
		e.Logf(b.String())
		return nil
	default:
		return fmt.Errorf("unsupported command: %q", text)
	}
	return nil
}
