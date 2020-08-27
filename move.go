package zoo

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	Steps       [64]Bitboard
	RabbitSteps [2][64]Bitboard
)

func init() {
	for i := 0; i < 64; i++ {
		b := Bitboard(1) << i
		steps := b.Neighbors()
		Steps[i] = steps
		grSteps := steps
		if b&NotRank1 != 0 { // rabbits can't move backwards.
			grSteps ^= b << 8
		}
		srSteps := steps
		if b&NotRank8 != 0 {
			srSteps ^= b << 8
		}
		RabbitSteps[0][i] = grSteps
		RabbitSteps[1][i] = srSteps
	}
}

type Step struct {
	Src, Dest Square
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
	matches := setupPattern.FindStringSubmatch(s)
	if matches == nil {
		return Step{}, fmt.Errorf("input does not match /%s/", stepPattern)
	}
	src := ParseSquare(matches[2])
	dest := src.Translate(ParseDelta(matches[3]))
	if !dest.Valid() {
		return Step{}, fmt.Errorf("destination is invalid")
	}
	return Step{
		Src:  src,
		Dest: dest,
	}, nil
}
