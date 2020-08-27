package zoo

import (
	"fmt"
	"regexp"
)

const EmptyPositionShortString = `g [                                                                ]`

var shortPosPattern = regexp.MustCompile(`^([wbgs]) \[([ RCDHMErcdhme]{64})\]$`)

func ParseShortPosition(s string) (*Pos, error) {
	matches := shortPosPattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("input does not match /%s/", shortPosPattern)
	}
	side := ParseColor(matches[1])
	pos := NewPos(nil, nil, side, 4, false, Empty, 0, 0)
	for i, r := range matches[2] {
		square := Square(8*(7-i/8) + i%8)
		piece, err := ParsePiece(string(r))
		if err != nil {
			return nil, fmt.Errorf("at %s: %v", square.String(), err)
		}
		pos, err = pos.Place(piece, square)
		if err != nil {
			return nil, fmt.Errorf("at %s: %v", square.String(), err)
		}
	}
	return pos, nil
}
