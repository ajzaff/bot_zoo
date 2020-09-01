package zoo

import (
	"strconv"
	"strings"
)

func (a *AEI) handleZoo(text string) error {
	text = text[4:]
	switch {
	case text == "new", text == "newstandard":
		pos, _ := ParseShortPosition(PosStandard)
		pos.moveNum = 2
		a.engine.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "move"), strings.HasPrefix(text, "moves"):
		parts := strings.SplitN(text, " ", 2)
		n := 0
		if len(parts) == 2 {
			n, _ = strconv.Atoi(parts[1])
		}
		moves := a.engine.getRootMovesLen(a.engine.Pos(), 4)
		scoredMoves := a.engine.sortMoves(a.engine.Pos(), moves)
		if n == 0 {
			n = len(scoredMoves)
		}
		for i, e := range scoredMoves {
			if i >= n {
				break
			}
			a.Logf("[%d] %s", e.score, MoveString(e.move))
		}
		a.Logf("%d", len(scoredMoves))
		return nil
	case text == "hash":
		a.Logf("%X", a.engine.Pos().zhash)
		return nil
	case strings.HasPrefix(text, "step"), strings.HasPrefix(text, "makestep"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			for _, step := range a.engine.Pos().Steps() {
				a.Logf("%s", step)
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
	case text == "nullmove":
		a.engine.Pos().NullMove()
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
		case "w", "g":
			b = a.engine.Pos().presence[Gold]
		case "b", "s":
			b = a.engine.Pos().presence[Silver]
		case "short":
			a.Logf(a.engine.Pos().ShortString())
			return nil
		default:
			p, _ := ParsePiece(parts[1])
			bs := a.engine.Pos().bitboards
			if bs != nil {
				b = bs[p]
			}
		}
		a.Logf(b.String())
		return nil
	default:
		return nil
	}
}

func (a *AEI) logEval() {
	a.Logf("eval: %d", a.engine.Pos().Score())
}

func (a *AEI) verbosePos() {
	a.Logf(a.engine.Pos().String())
	a.logEval()
}
