package zoo

import (
	"fmt"
	"regexp"
	"strings"
)

type Step struct {
	Src, Dest Square
	Piece
	Dir string
}

func (s Step) Valid() bool {
	return s.Setup() || s.Src.Valid() && s.Dest.Valid() && s.Dir != ""
}

func (s Step) Setup() bool {
	return !s.Src.Valid() && s.Dir == "" && s.Dest.Valid() && s.Piece != Empty
}

func (s Step) Capture() bool {
	return s.Dir == "x"
}

func (s Step) String() string {
	if s.Setup() {
		return fmt.Sprintf("%c%s", s.Piece.Byte(), s.Dest.String())
	}
	return fmt.Sprintf("%c%s%s", s.Piece.Byte(), s.Src.String(), s.Dir)
}

func ParseMove(s string) ([]Step, error) {
	parts := strings.Split(s, " ")
	var res []Step
	for _, part := range parts {
		step, err := ParseStep(part)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", part, err)
		}
		res = append(res, step)
	}
	return res, nil
}

var stepPattern = regexp.MustCompile(`^([RCDHMErcdhme])([a-h][1-8])([nsewx])?$`)

func ParseStep(s string) (Step, error) {
	matches := stepPattern.FindStringSubmatch(s)
	if matches == nil {
		return Step{}, fmt.Errorf("input does not match /%s/", stepPattern)
	}
	piece, err := ParsePiece(matches[1])
	if err != nil {
		return Step{}, err
	}
	at := ParseSquare(matches[2])
	if matches[3] == "" {
		return Step{
			Src:   invalidSquare,
			Dest:  at,
			Piece: piece,
		}, nil
	}
	dest := at.Translate(ParseDelta(matches[3]))
	if !dest.Valid() {
		return Step{}, fmt.Errorf("destination is invalid")
	}
	return Step{
		Src:   at,
		Dest:  dest,
		Piece: piece,
		Dir:   matches[3],
	}, nil
}

// MoveString outputs a legal move string by adding captures if missing.
func (p *Pos) MoveString(move []Step) string {
	var sb strings.Builder
	for i, step := range move {
		if !step.Capture() {
			sb.WriteString(step.String())
		}
		if c := p.capture(step); c.Capture() {
			fmt.Fprintf(&sb, " %s", c.String())
		}
		if i+1 < len(move) {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}
