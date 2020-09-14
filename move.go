package zoo

import (
	"bufio"
	"fmt"
	"strings"
)

func splitMove(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// A maximum of 2 split indices for a sequence with 2 steps and a capture.
	// A third split corresponds to end of the advance return value.
	var indices []int

	for i, p := 0, 0; i < 3; i++ {
		for ; p < len(data) && data[p] == ' '; p++ {
		}
		indices = append(indices, p)
		if p == len(data) {
			break
		}
		for ; p < len(data) && data[p] != ' '; p++ {
		}
		indices = append(indices, p)
		if p == len(data) {
			break
		}
	}
	// Check if there's no step to return:
	if len(indices) < 2 {
		if atEOF {
			return len(data), nil, nil
		}
		// We need more data:
		return 0, nil, nil
	}
	// Check if first move is a setup move and if so return it:
	if stepLen := indices[1] - indices[0]; stepLen <= 3 {
		return indices[1], data[indices[0]:indices[1]], nil
	}
	// Handle a single step left:
	if len(indices) < 4 {
		if atEOF {
			return indices[1], data[indices[0]:indices[1]], nil
		}
		// We need more data:
		return 0, nil, nil
	}
	// Check the next steps to see if they go together.
	// Steps go together if they are a related push, pull or capture.
	// The following patterns are possible:
	//	push PUSH
	//	PULL pull
	//	push PUSH CAP
	//	PULL pull CAP
	p1, _ := ParsePiece(data[indices[0]])
	p2, _ := ParsePiece(data[indices[2]])
	s1 := ParseSquare(string(data[indices[0]+1 : indices[0]+3]))
	s2 := ParseSquare(string(data[indices[2]+1 : indices[2]+3]))
	d1 := s1.Translate(MakeDeltaFromByte(data[indices[0]+3]))
	d2 := s2.Translate(MakeDeltaFromByte(data[indices[2]+3]))
	cap := data[indices[2]+3] == 'x'

	// The do not match the pattern and should not go together:
	if !cap && (p1.SameColor(p2) || p1.SameType(p2) || (p1.WeakerThan(p2) && s2 != d1) || (!p1.WeakerThan(p2) && s1 != d2)) {
		return indices[1], data[indices[0]:indices[1]], nil
	}

	if len(indices) < 6 {
		// There cannot be two captures in a row so we know the sequence ends.
		if cap || atEOF {
			return indices[3], data[indices[0]:indices[3]], nil
		}
		// The push/pull may lead to a capture and we need more data:
		return 0, nil, nil
	}

	// The push/pull leads to a capture:
	if cap2 := data[indices[4]+3] == 'x'; cap2 {
		return indices[5], data[indices[0]:indices[5]], nil
	}

	// Return the push/pull sequence
	return indices[3], data[indices[0]:indices[3]], nil
}

// ParseMove parses the move string into steps
// and checks for validity but not legality.
func ParseMove(s string) ([]Step, error) {
	var (
		sc  = bufio.NewScanner(strings.NewReader(s))
		res []Step
	)
	sc.Split(splitMove)
	for sc.Scan() {
		step, err := ParseStep(sc.Text())
		if err != nil {
			return nil, fmt.Errorf("%s: %v", sc.Text(), err)
		}
		res = append(res, step)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%s: %v", s, err)
	}
	res = append(res, Pass)
	return res, nil
}

// MoveString outputs a legal move string.
func MoveString(move []Step) string {
	var sb strings.Builder
	var join bool
	for _, step := range move {
		if step.Pass() {
			continue
		}
		if join {
			sb.WriteByte(' ')
		}
		join = true
		sb.WriteString(step.String())
	}
	return sb.String()
}

func MoveLen(move []Step) int {
	n := 0
	for _, step := range move {
		n += step.Len()
	}
	return n
}

func MoveEqual(a, b []Step) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Step represents a compact step as used in the engine.
// It uses the following layout (little endian):
//	src (8 bits)
//	dest delta (3 bits)
//	alt (8 bits)
//	piece1 (4 bits)
//	piece2 (4 bits)
//	capture piece (4 bits)
// A Pass step has all the capture bits set, which would
// correspond to an illegal piece. The canonical invalid
// step has all bits set.
type Step uint32

// MakeStep makes a step with all given parameters.
func MakeStep(src, dest, alt Square, p1, p2, cap Piece) Step {
	return Step(src&0xff) |
		Step(src.Delta(dest).Packed())<<8 |
		Step(alt&0xff)<<11 |
		Step(p1&0xf)<<19 |
		Step(p2&0xf)<<24 |
		Step(cap&0xf)<<28
}

// MakeSetup makes a setup step.
func MakeSetup(piece Piece, alt Square) Step {
	v := MakeStep(0xff, 0xff, alt, piece, 0xf, 0xf)
	v |= Step(piece)
	return v
}

// MakeDefaultCapture makes a default step with a capture.
func MakeDefaultCapture(src, dest Square, p1, cap Piece) Step {
	return MakeStep(src, dest, 0xff, p1, 0xf, cap)
}

// MakeDefault makes a default step.
func MakeDefault(src, dest Square, p1 Piece) Step {
	return MakeDefaultCapture(src, dest, p1, 0xf)
}

// MakeAlternateCapture makes a push/pull step with a capture.
func MakeAlternateCapture(src, dest, alt Square, p1, p2, cap Piece) Step {
	return MakeStep(src, dest, alt, p1, p2, cap)
}

// MakeAlternate makes a push/pull step.
func MakeAlternate(src, dest, alt Square, p1, p2 Piece) Step {
	return MakeAlternateCapture(src, dest, alt, p1, p2, 0xf)
}

// Pass is a Step representing the pas step which delimits all moves.
const Pass Step = 0xf0000000

const invalidStep Step = 0xffffffff

// Src square for the originating piece.
func (s Step) Src() Square {
	return Square(s & 0xff)
}

// Dest square for the originating piece.
func (s Step) Dest() Square {
	return s.Src().Translate(MakeDeltaFromPacked(uint8((s >> 8) & 3)))
}

// Alt square for the alternate piece.
func (s Step) Alt() Square {
	return Square((s >> 11) & 0xff)
}

// Piece1 returns the primary piece square.
func (s Step) Piece1() Piece {
	return Piece((s >> 19) & 0xf)
}

// Piece2 returns the alternate piece.
func (s Step) Piece2() Piece {
	return Piece((s >> 24) & 0xf)
}

// Cap returns the capture piece.
func (s Step) Cap() Piece {
	return Piece((s >> 28) & 0xf)
}

// Capture returns whether the step has a capture.
func (s Step) Capture() bool {
	return s&0xf0000000 < 0xf0000000
}

// CapSquare computes the capture square given the position p as context.
func (s Step) CapSquare() Square {
	if !s.Capture() {
		return invalidSquare
	}
	c := s.Piece1().Color()
	cap := s.Cap()
	switch s.Kind() {
	case KindDefault:
		return s.Dest()
	case KindPush:
		alt := s.Alt()
		if alt.Trap() {
			return alt
		}
		if dest := s.Dest(); dest.Trap() {
			return dest
		}
		if v := s.Src().AdjacentTrap(); c == cap.Color() && v.Valid() {
			return v
		}
		return s.Dest().AdjacentTrap()
	case KindPull:
		src := s.Src()
		if v := src.AdjacentTrap(); c == cap.Color() && v.Valid() {
			return v
		}
		dest := s.Dest()
		if dest.Trap() {
			return dest
		}
		return s.Alt().AdjacentTrap()
	default:
		return invalidSquare
	}
}

// Pass returns whether the step passes the turn.
func (s Step) Pass() bool {
	return s == Pass
}

// StepKind defines the granular kind of step such as
// default, push, or pull, used in String.
type StepKind uint8

// StepKind constants.
const (
	KindInvalid StepKind = iota
	KindDefault
	KindSetup
	KindPush
	KindPull
)

// ParseStep parses a single step and optional capture.
func ParseStep(s string) (Step, error) {
	// Check if the step is too short:
	if len(s) < 3 {
		return invalidStep, fmt.Errorf("too short step: %s", s)
	}
	piece1, err := ParsePiece(s[0])
	if err != nil {
		return invalidStep, fmt.Errorf("invalid first piece: %v", err)
	}
	src1 := ParseSquare(s[1:3])
	if !src1.Valid() {
		return invalidStep, fmt.Errorf("invalid first square: %q", s[1:3])
	}
	// Return the setup step:
	if len(s) == 3 {
		return MakeSetup(piece1, src1), nil
	}
	delta1 := MakeDeltaFromByte(s[3])
	if delta1 == 0 {
		return invalidStep, fmt.Errorf("invalid first delta: %c", s[3])
	}
	dest1 := src1.Translate(delta1)
	// Return single step:
	if len(s) == 4 {
		return MakeDefault(src1, dest1, piece1), nil
	}
	if len(s) < 9 {
		return invalidStep, fmt.Errorf("too short step sequence: %q", s)
	}
	piece2, err := ParsePiece(s[5])
	if err != nil {
		return invalidStep, fmt.Errorf("invalid second piece: %v", err)
	}
	src2 := ParseSquare(s[6:8])
	if !src2.Valid() {
		return invalidStep, fmt.Errorf("invalid second square: %q", s[6:8])
	}
	cap := Piece(0xf)
	delta2 := MakeDeltaFromByte(s[8])
	if s[8] == 'x' {
		if !piece1.SameColor(piece2) {
			return invalidStep, fmt.Errorf("invalid first capture color: %q", s)
		}
		cap = piece2
	} else if delta2 == 0 {
		return invalidStep, fmt.Errorf("invalid second delta: %c", s[8])
	}
	dest2 := src2.Translate(delta2)

	// Return default self capture:
	if cap.Valid() && len(s) == 9 {
		return MakeDefaultCapture(src1, dest1, piece1, piece2), nil
	}
	if piece1.SameColor(piece2) {
		return invalidStep, fmt.Errorf("invalid push or pull color: %q", s)
	}
	// Step sequence is a push or pull with a possible capture:
	if piece1.WeakerThan(piece2) {
		// Swap step order in a push (when the opponents piece comes first):
		piece1, piece2 = piece2, piece1
		src1, src2 = src2, src1
		dest1, dest2 = dest2, dest1
	}
	// Alt for the push/pull becomes src2:
	// No capture:
	if len(s) == 9 {
		return MakeAlternate(src1, dest1, src2, piece1, piece2), nil
	}

	if cap, err = ParsePiece(s[10]); err != nil {
		return invalidStep, err
	}
	if v := ParseSquare(s[11:13]); !v.Valid() {
		return invalidStep, fmt.Errorf("invalid capture square: %q", s[11:13])
	}
	if s[13] != 'x' {
		return invalidStep, fmt.Errorf("invalid capture delta: %c", s[13])
	}
	return MakeAlternateCapture(src1, dest1, src2, piece1, piece2, cap), nil
}

func (s Step) Kind() StepKind {
	switch src, dest, alt := s.Src(), s.Dest(), s.Alt(); {
	case alt.Valid():
		switch {
		case src.Valid() && dest.Valid():
			switch {
			case dest.AdjacentTo(alt):
				return KindPush
			case src.AdjacentTo(alt):
				return KindPull
			default:
				return KindInvalid
			}
		case !s.Cap().Valid():
			return KindSetup
		default:
			return KindInvalid
		}
	case src.Valid() && dest.Valid():
		return KindDefault
	default:
		return KindInvalid
	}
}

func (s Step) Len() int {
	switch kind := s.Kind(); {
	case s.Pass():
		return 0
	case kind == KindDefault, kind == KindSetup:
		return 1
	case kind == KindPush, kind == KindPull:
		return 2
	default:
		return 0
	}
}

// String returns the formatted step given the position p.
func (s Step) String() string {
	var sb strings.Builder
	switch kind := s.Kind(); {
	case s.Pass():
		fmt.Fprint(&sb, "(pass)")
	case kind == KindSetup:
		fmt.Fprintf(&sb, "%c%s", s.Piece1().Byte(), s.Alt())
	case kind == KindPush:
		fmt.Fprintf(&sb, "%c%s%s", s.Piece2().Byte(), s.Dest(), s.Dest().Delta(s.Alt()).String())
		if s.Cap().SameColor(s.Piece2()) {
			fmt.Fprintf(&sb, " %c%sx", s.Cap().Byte(), s.CapSquare())
		}
		fmt.Fprintf(&sb, " %c%s%s", s.Piece1().Byte(), s.Src(), s.Src().Delta(s.Dest()).String())
		if s.Cap().SameColor(s.Piece1()) {
			fmt.Fprintf(&sb, " %c%sx", s.Cap().Byte(), s.CapSquare())
		}
	case kind == KindPull:
		fmt.Fprintf(&sb, "%c%s%s", s.Piece1().Byte(), s.Src(), s.Src().Delta(s.Dest()).String())
		if s.Cap().SameColor(s.Piece1()) {
			fmt.Fprintf(&sb, " %c%sx", s.Cap().Byte(), s.CapSquare())
		}
		fmt.Fprintf(&sb, " %c%s%s", s.Piece2().Byte(), s.Alt(), s.Alt().Delta(s.Src()).String())
		if s.Cap().SameColor(s.Piece2()) {
			fmt.Fprintf(&sb, " %c%sx", s.Cap().Byte(), s.CapSquare())
		}
	case kind == KindDefault:
		fmt.Fprintf(&sb, "%c%s%s", s.Piece1().Byte(), s.Src(), s.Src().Delta(s.Dest()).String())
		if s.Capture() {
			fmt.Fprintf(&sb, " %c%sx", s.Cap().Byte(), s.CapSquare())
		}
	default: // Invalid
		return s.GoString()
	}
	return sb.String()
}

// GoString returns the formatted GoString for this step.
func (s Step) GoString() string {
	return fmt.Sprintf("Step(src=%s, dest=%s, alt=%s, p1=%s, p2=%s, cap=%s)",
		s.Src(), s.Dest(), s.Alt(), s.Piece1(), s.Piece2(), s.Cap())
}
