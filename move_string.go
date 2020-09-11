package zoo

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var turnPattern = regexp.MustCompile(`^(\d+)([gswb])`)

func (e *Engine) NewGameFromMoveList(r io.Reader) error {
	sc := bufio.NewScanner(r)
	e.NewGame()
	var currentMove string
	for i := 1; sc.Scan(); i++ {
		text := strings.TrimSpace(sc.Text())
		if currentMove != "" {
			return fmt.Errorf("line %d: unexpected move after current move %s: %s", i, currentMove, text)
		}
		match := turnPattern.FindStringSubmatch(text)
		if len(match) < 3 {
			return fmt.Errorf("line %d: expected turn number and color matching /%s/: %s", i, turnPattern, text)
		}
		v, err := strconv.Atoi(match[1])
		if err != nil {
			return fmt.Errorf("line %d: bad turn number: %s: %s", i, match[1], text)
		}
		if v != e.Pos().moveNum {
			return fmt.Errorf("line %d: wrong turn number: %d: %s", i, v, text)
		}
		if match[2] != e.Pos().Side().String() {
			return fmt.Errorf("line %d: wrong color: %s: %s", i, match[2], text)
		}
		text = strings.TrimSpace(text[len(match[0]):])
		if text == "" {
			currentMove = match[1]
			continue
		}
		move, err := ParseMove(text)
		if err != nil {
			return fmt.Errorf("%s: bad move: %s: %v", match[0], text, err)
		}
		if err := e.Pos().Move(move); err != nil {
			return fmt.Errorf("%s: bad move: %s: %v", match[0], text, err)
		}
	}
	return nil
}
