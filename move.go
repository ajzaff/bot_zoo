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
			return nil, fmt.Errorf("failed to parse move %q: at %q: %v", s, s[i:], err)
		}
		i += step.strLen()
		move = append(move, step)
	}
	return move, nil
}

// Len returns the length of the Move m in number of steps.
func (m Move) Len() int {
	return len(m)
}

// strLen returns the length of the String representation of the Move.
func (m Move) strLen() int {
	var l int
	for _, step := range m {
		l += step.strLen()
	}
	return l
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
//	dest delta      3 bits
//	capture delta   3 bits
// The zero value is the canonical invalid Step value.
// Capture delta is relative to src.
// Src is always legal in the packed encoding.
type Step uint16

func MakeCaptureStep(piece Piece, src Square, delta, cap SquareDelta) Step {
	piece.Packed()
	return 0
}

func MakeStep(piece Piece, src Square, delta SquareDelta) Step {
	return MakeCaptureStep(piece, src, delta, 7)
}

var stepPattern = regexp.MustCompile(`([rcdhmeRCDHME])([a-f][1-8])([nsew])(?: ([rcdhmeRCDHME])([a-f][1-8])x)?`)

// ParseStep parses the step and possible capture.
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
		return err
	}
	src, err := ParseSquare(s[matches[3]:matches[4]])
	if err != nil {
		return err
	}
}

// appendString appends the Step string to the builder.
func (s Step) appendString(sb *strings.Builder) {

}

// strLen returns the length of the String representation of the Step.
func (s Step) strLen() int {
	if s.Capture() {
		return 9
	}
	return 3
}

// String returns the String representation of the Step and possible capture.
func (s Step) String() string {
	var sb strings.Builder
	s.appendString(&sb)
	return sb.String()
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

// Default returns true when this is a simple step (with no push or pull) but possible capture.
func (s Step) Default() bool {
	return !s.Alt().Valid()
}

// Push returns true when this step is a push.
func (s Step) Push() bool {
	dest, alt := s.Src(), s.Dest()
	return s.Alt().Valid() && dest.AdjacentTo(alt)
}

// Pull returns true when this step is a pull.
func (s Step) Pull() bool {
	src, alt := s.Src(), s.Alt()
	return s.Alt().Valid() && src.AdjacentTo(alt)
}

// SacrificesMaterial returns true when this step captures a piece of our own color.
// This is possible via a self capture or removing ourselves as a defender.
func (s Step) SacrificesMaterial() bool {
	return s.Capture() && s.Cap().SameColor(s.Piece1())
}

// SelfCapture returns true when this step captures the first piece.
func (s Step) SelfCapture() bool {
	return s.Capture() && s.Dest().Trap() && s.Cap() == s.Piece1()
}

// DirectCapture returns true when this step pushes or pulls a piece onto a trap square.
func (s Step) DirectCapture() bool {
	return s.Capture() && s.Dest().Trap() && (s.Src().Trap() || s.Alt().Trap())
}

// IndirectCapture returns true when this step captures a piece by removing a defender.
// This includes removing ourself as a defender.
func (s Step) IndirectCapture() bool {
	ours := s.SacrificesMaterial()
	return s.Capture() && (!ours && s.Alt().AdjacentTrap().Valid() || ours && s.Src().AdjacentTrap().Valid())
}

// CapSquare computes the capture square given the position p as context.
func (s Step) CapSquare() Square {
	// Confusingly the step notation does not store the square that the
	// captured piece was captured on. We need to compute it every time
	// based on this table:
	//	         push  pull  default
	//	direct   alt   src   n/a
	//	indirect dstA  altA  src
	//	self     dest  dest  dest
	if !s.Capture() {
		return invalidSquare
	}
	if s.SelfCapture() {
		return s.Dest()
	}
	if s.SacrificesMaterial() {
		if v := s.Src().AdjacentTrap(); v.Valid() {
			return v
		}
		panic(fmt.Sprintf("bad push: %v", s.GoString()))
	}
	switch s.Kind() {
	case KindPush:
		if s.DirectCapture() {
			return s.Alt()
		}
		// Indirect capture:
		if v := s.Dest().AdjacentTrap(); v.Valid() {
			return v
		}
		panic(fmt.Sprintf("bad push: %v", s.GoString()))
	case KindPull:
		if s.DirectCapture() {
			return s.Src()
		}
		// Indirect capture:
		if v := s.Alt().AdjacentTrap(); v.Valid() {
			return v
		}
		panic(fmt.Sprintf("bad pull: %v", s.GoString()))
	default:
		if s.IndirectCapture() {
			if v := s.Src().AdjacentTrap(); v.Valid() {
				return v
			}
		}
		panic(fmt.Sprintf("bad default: %v", s.GoString()))
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
