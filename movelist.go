package zoo

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MoveList represents a mutable list of game moves.
type MoveList []Move

var turnPattern = regexp.MustCompile(`^(\d+)([gswb])`)

// ParseMoveList reads a move list from the string s or returns an error.
// The movelist always starts at 1g including setup moves.
func ParseMoveList(s string) (MoveList, error) {
	var (
		sc          = bufio.NewScanner(strings.NewReader(s))
		turnNumber  = 1
		side        = Gold
		currentMove string
		res         MoveList
	)

	for i := 1; sc.Scan(); i++ {
		text := strings.TrimSpace(sc.Text())
		if currentMove != "" {
			return nil, fmt.Errorf("line %d: unexpected move after current move %s: %s", i, currentMove, text)
		}
		match := turnPattern.FindStringSubmatch(text)
		if len(match) < 3 {
			return nil, fmt.Errorf("line %d: expected turn number and color matching /%s/: %s", i, turnPattern, text)
		}
		v, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: bad turn number: %s: %s", i, match[1], text)
		}
		if v != turnNumber {
			return nil, fmt.Errorf("line %d: wrong turn number: %d: %s", i, v, text)
		}
		if match[2][0] != side.Byte() {
			return nil, fmt.Errorf("line %d: wrong color: %s: %s", i, match[2], text)
		}
		text = strings.TrimSpace(text[len(match[0]):])
		if text == "" {
			currentMove = match[1]
			continue
		}
		move, err := ParseMove(text)
		if err != nil {
			return nil, fmt.Errorf("line %d: bad move: %s: %s: %v", i, match[0], text, err)
		}
		res = append(res, move)
		side = side.Opposite()
		if side == Gold {
			turnNumber++
		}
	}
	return res, nil
}

// Last returns the last move in the list.
// If the last move is empty, the previous is returned.
func (l MoveList) Last() Move {
	n := len(l)
	if n == 0 {
		return nil
	}
	m := l[n-1]
	if len(m) > 0 {
		return m
	}
	if n == 1 {
		return nil
	}
	return l[n-2]
}

// appendString appends the MoveList string to sb.
func (l MoveList) appendString(sb *strings.Builder) {
	var (
		moveNumber = 1
		side       = Gold
	)
	for _, m := range l {
		fmt.Fprintf(sb, "%d%c ", moveNumber, side.Byte())
		m.appendString(sb)
		sb.WriteByte('\n')
		side = side.Opposite()
		if side == Gold {
			moveNumber++
		}
	}
}

// String returns the string representation of this MoveList.
// With turn numbers (e.g. 1g, 1s, 2g, ...) followed by the
// move played in standard notation. The last line has the
// current move number and side with an empty move.
func (l MoveList) String() string {
	var sb strings.Builder
	l.appendString(&sb)
	return sb.String()
}
