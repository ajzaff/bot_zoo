package zoo

import "strings"

func (a *AEI) handleZoo(text string) error {
	text = text[4:]
	switch {
	case text == "new", text == "newstandard":
		pos, _ := ParseShortPosition(PosStandard)
		pos.MoveNum = 2
		a.engine.SetPos(pos)
		return nil
	case strings.HasPrefix(text, "step"), strings.HasPrefix(text, "makestep"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			for _, step := range a.engine.Pos().GetSteps(true) {
				a.Logf("%s", step)
			}
			return nil
		}
		step, err := ParseStep(parts[1])
		if err != nil {
			return err
		}
		pos, _, err := a.engine.Pos().Step(step)
		if err != nil {
			return err
		}
		a.engine.SetPos(pos)
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case text == "nullmove":
		a.engine.SetPos(a.engine.Pos().NullMove())
		if a.verbose {
			a.verbosePos()
		}
		return nil
	case text == "movenow":
		move := a.engine.BestMove()
		if len(move) == 0 {
			a.Logf("no moves")
			return nil
		}
		a.Logf("move is %s", MoveString(move))
		pos, _, err := a.engine.Pos().Move(move, true)
		if err != nil {
			return err
		}
		a.engine.SetPos(pos)
		return nil
	case text == "search":
		move, score := a.engine.Search()
		a.Logf("search: %d %s [depth=6]:", score, MoveString(move))
		return nil
	case text == "eval":
		a.logScore()
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
			b = a.engine.Pos().Presence[Gold]
		case "b", "s":
			b = a.engine.Pos().Presence[Silver]
		case "short":
			a.Logf(a.engine.Pos().ShortString())
			return nil
		default:
			p, _ := ParsePiece(parts[1])
			bs := a.engine.Pos().Bitboards
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

func (a *AEI) logScore() {
	a.Logf("eval: %d", a.engine.Pos().Score())
}

func (a *AEI) verbosePos() {
	a.Logf(a.engine.Pos().String())
	a.logScore()
}
