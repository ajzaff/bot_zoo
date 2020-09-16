package zoo

import (
	"fmt"
	"strings"
)

func (a *AEI) handleExt(text string) error {
	switch {
	case text == "new", text == "newstandard":
		a.engine.NewGame()
		pos, err := ParseShortPosition(posStandard)
		if err != nil {
			panic(err)
		}
		pos.moveNum = 2
		a.engine.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "unmove"):
		if err := a.engine.Pos().Unmove(); err != nil {
			return err
		}
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case text == "hash":
		a.Logf("%d", a.engine.Pos().zhash)
		return nil
	case text == "depth":
		a.Logf("%d", a.engine.Pos().Depth())
		return nil
	case strings.HasPrefix(text, "step"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			var stepList StepList
			stepList.Generate(a.engine.Pos())
			for i := 0; i < stepList.Len(); i++ {
				step := stepList.At(i)
				a.Logf("[%d] %s", step.Value, step.Step)
			}
			return nil
		}
		step, err := ParseStep(parts[1])
		if err != nil {
			return err
		}
		if err := a.engine.Pos().Step(step); err != nil {
			return err
		}
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case strings.HasPrefix(text, "unstep"):
		if err := a.engine.Pos().Unstep(); err != nil {
			return err
		}
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case text == "pass":
		a.engine.Pos().Pass()
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case text == "unpass":
		if err := a.engine.Pos().Unpass(); err != nil {
			return err
		}
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case text == "eval":
		a.logEval()
		return nil
	case strings.HasPrefix(text, "print"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			a.Logf(a.engine.Pos().String())
			return nil
		}
		var b Bitboard
		switch parts[1] {
		case "weaker":
			b = a.engine.Pos().touching[Gold]
			for t := GRabbit; t <= GElephant; t++ {
				b = a.engine.Pos().weaker[t]
				a.Logf(t.String())
				a.Logf(b.String())
			}
			return nil
		case "stronger":
			for t := GRabbit; t <= GElephant; t++ {
				b = a.engine.Pos().stronger[t]
				a.Logf(t.String())
				a.Logf(b.String())
			}
			return nil
		case "tg":
			b = a.engine.Pos().touching[Gold]
		case "ts":
			b = a.engine.Pos().touching[Silver]
		case "dg":
			b = a.engine.Pos().dominating[Gold]
		case "ds":
			b = a.engine.Pos().dominating[Silver]
		case "fg":
			b = a.engine.Pos().frozen[Gold]
		case "fs":
			b = a.engine.Pos().frozen[Silver]
		case "g":
			b = a.engine.Pos().presence[Gold]
		case "s":
			b = a.engine.Pos().presence[Silver]
		case "short":
			a.Logf(a.engine.Pos().ShortString())
			return nil
		default:
			p, err := ParsePiece(parts[1][0])
			if err != nil {
				return fmt.Errorf("printing piece bitboard: %v", err)
			}
			b = a.engine.Pos().bitboards[p]
		}
		a.Logf(b.String())
		return nil
	default:
		return fmt.Errorf("unsupported command: %q", text)
	}
	return nil
}

func (a *AEI) logEval() {
	a.Logf("eval: %d", a.engine.Pos().Value())
}

func (a *AEI) verbosePos() {
	a.Logf(a.engine.Pos().String())
	a.logEval()
}
