package zoo

import (
	"fmt"
	"strconv"
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
	case text == "movenow":
		a.engine.GoFixed(8)
		move, _, _ := a.engine.table.Best(a.engine.Pos())
		if len(move) > 0 {
			a.engine.Pos().Move(move)
		} else {
			a.Logf("no best move found in table")
		}
	case strings.HasPrefix(text, "moves"):
		parts := strings.SplitN(text, " ", 2)
		n := 0
		if len(parts) == 2 {
			n, _ = strconv.Atoi(parts[1])
		}
		moves := a.engine.getRootMovesLen(a.engine.Pos(), 4)
		scoredMoves := make([]ScoredMove, len(moves))
		for i, move := range moves {
			scoredMoves[i] = ScoredMove{score: -inf, move: move}
		}
		a.engine.rescorePVMoves(a.engine.Pos(), scoredMoves)
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
	case text == "pv":
		pv, score, err := a.engine.table.PV(a.engine.Pos())
		if err != nil {
			a.Logf("PV error: %v", err)
		}
		if len(pv) > 0 {
			a.Logf("[%d] %s", score, MoveString(pv))
		}
		return nil
	case strings.HasPrefix(text, "step"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			var steps []Step
			a.engine.Pos().Steps(&steps)
			selector := a.engine.stepSelector(steps)
			for score, step, ok := selector.SelectScore(); ok; score, step, ok = selector.SelectScore() {
				a.Logf("[%d] %s", score, step)
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
		case "fg":
			b = a.engine.Pos().frozen[Gold]
		case "fs":
			b = a.engine.Pos().frozen[Silver]
		case "g":
			b = a.engine.Pos().presence[Silver]
		case "s":
			b = a.engine.Pos().presence[Silver]
		case "short":
			a.Logf(a.engine.Pos().ShortString())
			return nil
		default:
			p, _ := ParsePiece(parts[1][0])
			bs := a.engine.Pos().bitboards
			if bs != nil {
				b = bs[p]
			}
		}
		a.Logf(b.String())
		return nil
	default:
		return fmt.Errorf("unsupported command: %q", text)
	}
	return nil
}

func (a *AEI) logEval() {
	a.Logf("eval: %d", a.engine.Pos().Score())
}

func (a *AEI) verbosePos() {
	a.Logf(a.engine.Pos().String())
	a.logEval()
}
