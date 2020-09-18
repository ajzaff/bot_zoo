package zoo

import (
	"fmt"
	"regexp"
	"strings"
)

// Move represents a sequence of up to 4 steps comprising of a single turn.
type Move []Step

// ParseMove parses the move and returns any errors.
// It does not attempt to validate the legality of the move.
func ParseMove(s string) (Move, error) {
	var move Move
	for i := 0; i < len(s); {
		step, err := ParseStep(s[i:])
		if err != nil {
			return nil, fmt.Errorf("invalid step at %q: %v", s[i:], err)
		}
		move = append(move, step)
	}
	return move, nil
}

// Len returns the length of the Move m in number of steps.
func (m Move) Len() int {
	return len(m)
}

// Equals returns true if m and move contain the same Steps.
func (m Move) Equals(move Move) bool {
	if m.Len() != move.Len() {
		return false
	}
	for i := range m {
		if m[i] != move[i] {
			return false
		}
	}
	return true
}

// appendString appends the move to the builder.
// This is useful for formatting PV lines.
func (m Move) appendString(sb *strings.Builder) {
	for i, step := range m {
		if i > 0 {
			sb.WriteByte(' ')
		}
		step.appendString(sb)
	}
}

// String returns a string representation of the Move.
func (m Move) String() string {
	var sb strings.Builder
	m.appendString(&sb)
	return sb.String()
}

// Step represents a compact step as used in the engine.
// It uses the following 16-bit layout (little endian):
//	piece           4 bits
//	src (packed)    6 bits
//	dest (packed)   6 bits
// The zero value is the canonical invalid Step value.
// Src and Dest are always legal in the packed encoding.
// Step is capable of encoding setup steps and captures.
// Setup steps assign the setup square to dest and E4 to
// src (an illegal step otherwise since dest will always
// be on the back two ranks).
// Similarly, captures assign the capture square to src
// and use E4 as dest.
type Step uint16

// MakeStep creates a step with a src and dest but no capture Square.
func MakeStep(piece Piece, src, dest Square) Step {
	return Step(uint8(piece)&0b1111 |
		uint8(src&0b111111)<<4 |
		uint8(dest&0b111111)<<10)
}

// MakeCaptureStep creates a step with a src, dest and capture Square.
func MakeCapture(piece Piece, cap Square) Step {
	return MakeStep(piece, cap, E4)
}

// MakeSetup makes a setup step in which src and dest are the same Square.
func MakeSetup(piece Piece, setup Square) Step {
	return MakeStep(piece, E4, setup)
}

var stepPattern = regexp.MustCompile(`([rcdhmeRCDHME])([a-f][1-8])([xnsew])?`)

// ParseStep parses the Arimaa step.
// It handles steps of the form:
//	Rd2
//	me3n
//	Df6x
// It does not attempt to validate the legality of the step.
func ParseStep(s string) (Step, error) {
	matches := stepPattern.FindStringSubmatchIndex(s)
	if matches == nil {
		return 0, fmt.Errorf("does not match /%s/")
	}
	if matches[0] != 0 {
		return 0, fmt.Errorf("unexpected string at start: %q", s[:matches[0]])
	}
	piece, err := ParsePiece(s[matches[2]])
	if err != nil {
		return 0, fmt.Errorf("bad piece: %v", err)
	}
	src, err := ParseSquare(s[matches[3]:matches[4]])
	if err != nil {
		return 0, err
	}
	if len(matches) < 6 {
		return MakeSetup(piece, src), nil
	}
	d := s[matches[5]]
	if d == 'x' {
		return MakeCapture(piece, src), nil
	}
	dest := ParseDir(s[matches[5]])
	if dest == DirNone {
		return 0, fmt.Errorf("bad step direction: %c", s[matches[5]])
	}
	return MakeStep(piece, src, src.Add(dest)), nil
}

// Piece returns the primary Piece for this Step.
func (s Step) Piece() Piece {
	return Piece(s & 0b1111)
}

// Src returns the originating Square for the Step.
func (s Step) Src() Square {
	return Square((s & 0b1111110000) >> 4)
}

// Dest returns the Direction of the destination relative to Src.
func (s Step) Dest() Square {
	return Square((s & 0b1111110000000000) >> 10)
}

// Capture returns true if the Step is a capture.
func (s Step) Capture() bool {
	return s&0b1111110000000000 == 0b111000000000000 &&
		s&0b1111110000 != 0b1001000000 &&
		s&0b1111110000 != 0b111010000 &&
		s&0b1111110000 != 0b101000000 &&
		s&0b1111110000 != 0b110110000
}

// Setup returns true if the Step is a setup Step.
func (s Step) Setup() bool {
	return s&0b1111110000 == 0b111000000 &&
		s&0b1111110000000000 != 0b1001000000000000 &&
		s&0b1111110000000000 != 0b111010000000000 &&
		s&0b1111110000000000 != 0b101000000000000 &&
		s&0b1111110000000000 != 0b110110000000000
}

// appendString appends the Step string to the builder.
func (s Step) appendString(sb *strings.Builder) {
	sb.WriteByte(s.Piece().Byte())
	src := s.Src()
	sb.WriteString(src.String())
	if !s.Setup() {
		if s.Capture() {
			sb.WriteByte('x')
			return
		}
		sb.WriteByte(s.Dest().Sub(src).Byte())
	}
}

// String returns the String representation of the Step and possible capture.
func (s Step) String() string {
	var sb strings.Builder
	s.appendString(&sb)
	return sb.String()
}

// GoString returns the formatted GoString for this step.
func (s Step) GoString() string {
	return fmt.Sprintf("Step(piece=%s src=%s, dest=%s)", s.Piece(), s.Src(), s.Dest())
}
