package zoo

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	stepsB       [64]Bitboard
	rabbitStepsB [2][64]Bitboard
)

func init() {
	for i := Square(0); i < 64; i++ {
		b := i.Bitboard()
		steps := b.Neighbors()
		stepsB[i] = steps
		grSteps := steps
		if b&NotRank1 != 0 { // rabbits can't move backwards.
			grSteps ^= b << 8
		}
		srSteps := steps
		if b&NotRank8 != 0 {
			srSteps ^= b << 8
		}
		rabbitStepsB[0][i] = grSteps
		rabbitStepsB[1][i] = srSteps
	}
}

func StepsFor(p Piece, i Square) Bitboard {
	if p.SamePiece(GRabbit) {
		return rabbitStepsB[p.Color()][i]
	}
	return stepsB[i]
}

type Step struct {
	Src, Dest Square
	Piece
	Dir string
}

func (s Step) String() string {
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

var setupPattern = regexp.MustCompile(`^([RCDHME rcdhme])([a-h][1-8])$`)

func ParseSetup(s string) (Piece, Square, error) {
	matches := setupPattern.FindStringSubmatch(s)
	if matches == nil {
		return Empty, 0, fmt.Errorf("input does not match /%s/", setupPattern)
	}
	piece, err := ParsePiece(matches[1])
	if err != nil {
		return Empty, 0, err
	}
	return piece, ParseSquare(s), nil
}

var stepPattern = regexp.MustCompile(`^([RCDHMErcdhme])([a-h][1-8])([nsewx])$`)

func ParseStep(s string) (Step, error) {
	matches := stepPattern.FindStringSubmatch(s)
	if matches == nil {
		return Step{}, fmt.Errorf("input does not match /%s/", stepPattern)
	}
	piece, err := ParsePiece(matches[1])
	if err != nil {
		return Step{}, err
	}
	src := ParseSquare(matches[2])
	dest := src.Translate(ParseDelta(matches[3]))
	if !dest.Valid() {
		return Step{}, fmt.Errorf("destination is invalid")
	}
	return Step{
		Src:   src,
		Dest:  dest,
		Piece: piece,
		Dir:   matches[3],
	}, nil
}

func MoveString(move []Step) string {
	var sb strings.Builder
	for i, step := range move {
		sb.WriteString(step.String())
		if i+1 < len(move) {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}
