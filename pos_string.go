package zoo

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	PosEmpty     = `g [                                                                ]`
	PosStandard  = `g [rrrrrrrrhdcemcdh                                HDCMECDHRRRRRRRR]`
	PosStandardG = `g [rrrrrrrrhdcemcdh                                                ]`
)

var shortPosPattern = regexp.MustCompile(`^([wbgs]) \[([ RCDHMErcdhme]{64})\]$`)

func ParseShortPosition(s string) (*Pos, error) {
	matches := shortPosPattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("input does not match /%s/", shortPosPattern)
	}
	side := ParseColor(matches[1])
	pos := newPos(nil, nil, side, 2, nil, 0)
	for i, r := range matches[2] {
		square := Square(8*(7-i/8) + i%8)
		piece, err := ParsePiece(string(r))
		if err != nil {
			return nil, fmt.Errorf("at %s: %v", square.String(), err)
		}
		if piece == Empty {
			continue
		}
		if err := pos.Place(piece, square); err != nil {
			return nil, fmt.Errorf("at %s: %v", square.String(), err)
		}
	}
	return pos, nil
}

func (p *Pos) ShortString() string {
	if p == nil {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s [", p.side.String())
	for i := 7; i >= 0; i-- {
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			sb.WriteByte(p.At(at).Byte())
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (p *Pos) String() string {
	if p == nil {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d%c", p.moveNum, p.side.Byte())
	if len(p.steps) > 0 {
		fmt.Fprintf(&sb, " %s", p.MoveString(p.steps))
	}
	sb.WriteString("\n +-----------------+\n")
	for i := 7; i >= 0; i-- {
		fmt.Fprintf(&sb, "%d| ", i+1)
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			atB := at.Bitboard()
			piece := p.At(at)
			if piece == Empty {
				if atB&Traps != 0 {
					sb.WriteByte('x')
				} else {
					sb.WriteByte('.')
				}
			} else {
				sb.WriteByte(piece.Byte())
			}
			sb.WriteByte(' ')
		}
		sb.WriteString("|\n")
	}
	sb.WriteString(" +-----------------+\n   a b c d e f g h")
	return sb.String()
}
