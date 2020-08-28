package zoo

import (
	"fmt"
	"strings"
)

func (a *AEI) handleZoo(text string) error {
	text = text[4:]
	switch {
	case text == "new", text == "newstandard":
		pos, _ := ParseShortPosition(PosStandard)
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
		a.engine.SetPos(a.engine.Pos().Step(step))
		return nil
	case text == "nullmove":
		a.engine.SetPos(a.engine.Pos().NullMove())
		return nil
	case strings.HasPrefix(text, "print"):
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			fmt.Println(a.engine.Pos().String())
		} else {
			var b Bitboard
			switch parts[1] {
			case "w", "g":
				b = a.engine.Pos().Presence[Gold]
			case "b", "s":
				b = a.engine.Pos().Presence[Silver]
			default:
				p, _ := ParsePiece(parts[1])
				bs := a.engine.Pos().Bitboards
				if bs != nil {
					b = bs[p]
				}
			}
			fmt.Println(b)
		}
		return nil
	default:
		return nil
	}
}
