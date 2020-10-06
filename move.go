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
		for ; i < len(s) && s[i] == ' '; i++ {
		}
		step, err := ParseStep(s[i:])
		if err != nil {
			return nil, fmt.Errorf("invalid step at %q: %v", s[i:], err)
		}
		move = append(move, step)
		if !step.Setup() {
			i++
		}
		i += 3
	}
	return move, nil
}

// Len returns the length of the Move m in number of steps not including captures.
func (m Move) Len() (n int) {
	for _, step := range m {
		if !step.Capture() {
			n++
		}
	}
	return n
}

// Equals returns true if m and move contain the same Steps.
func (m Move) Equals(move Move) bool {
	if len(m) != len(move) {
		return false
	}
	for i := range m {
		if m[i] != move[i] {
			return false
		}
	}
	return true
}

// Recurring extends Step.Recurring and returns whether appending s would make m recurring.
// By the Arimaa rules, a move that would result in no change to the position is not allowed.
// However, it is ok to have intermediate recurring positions, such as when pulling a piece.
func (m Move) Recurring(s Step) bool {
	if s.Capture() {
		return false
	}
	for _, x := range m {
		if x.Capture() {
			continue
		}
		if x.Recurring(s) {
			return true
		}
	}
	return false
}

// LastIndex returns the index of the last step that is not a capture or -1.
func (m Move) LastIndex() int {
	for i := len(m) - 1; i >= 0; i-- {
		if !m[i].Capture() {
			return i
		}
	}
	return -1
}

// Last returns the last step that is not a capture or 0.
func (m Move) Last() Step {
	if i := m.LastIndex(); i >= 0 {
		return m[i]
	}
	return 0
}

// Pop removes and returns the last step with possible capture.
func (m *Move) Pop() (s, cap Step) {
	n := len(*m) - 1
	if n >= 0 && (*m)[n].Capture() {
		cap = (*m)[n]
		n--
	}
	if n >= 0 {
		s = (*m)[n]
		*m = (*m)[:n]
	}
	return s, cap
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
	return Step(uint16(piece)&0b1111 |
		uint16(src&0b111111)<<4 |
		uint16(dest&0b111111)<<10)
}

// MakeCapture creates a step with a src, dest and capture Square.
func MakeCapture(piece Piece, cap Square) Step {
	return MakeStep(piece, cap, E4)
}

// MakeSetup makes a setup step in which src and dest are the same Square.
func MakeSetup(piece Piece, setup Square) Step {
	return MakeStep(piece, E4, setup)
}

var stepPattern = regexp.MustCompile(`^([rcdhmeRCDHME])([a-h][1-8])([xnsew])?`)

// ParseStep parses the Arimaa step.
// It handles steps of the form:
//	Rd2
//	me3n
//	Df6x
// It does not attempt to validate the legality of the step.
func ParseStep(s string) (Step, error) {
	matches := stepPattern.FindStringSubmatchIndex(s)
	if matches == nil {
		return 0, fmt.Errorf("does not match /%s/", stepPattern.String())
	}
	piece, err := ParsePiece(s[matches[2]])
	if err != nil {
		return 0, fmt.Errorf("bad piece: %v", err)
	}
	src, err := ParseSquare(s[matches[4]:matches[5]])
	if err != nil {
		return 0, err
	}
	if matches[6] == -1 {
		return MakeSetup(piece, src), nil
	}
	d := s[matches[6]]
	if d == 'x' {
		return MakeCapture(piece, src), nil
	}
	dest := ParseDir(d)
	if dest == DirNone {
		return 0, fmt.Errorf("bad step direction: %c", d)
	}
	return MakeStep(piece, src, src.Add(dest)), nil
}

var stepTable = [...]Step{
	128, 16, 1168, 1056, 1024, 2208, 2096, 2064, 3248, 3136, 3104, 4288, 4176, 4144, 5328, 5216, 5184,
	6368, 6256, 6224, 7408, 7264, 8448, 8336, 8192, 9488, 9376, 9344, 9232, 10528, 10416, 10384, 10272,
	11568, 11456, 11424, 11312, 12608, 12496, 12464, 12352, 13648, 13536, 13504, 13392, 14688, 14576, 14544, 14432,
	15728, 15584, 15472, 16768, 16656, 16512, 17808, 17696, 17664, 17552, 18848, 18736, 18704, 18592, 19888, 19776,
	19744, 19632, 20928, 20816, 20784, 20672, 21968, 21856, 21824, 21712, 23008, 22896, 22864, 22752, 24048, 23904,
	23792, 25088, 24976, 24832, 26128, 26016, 25984, 25872, 27168, 27056, 27024, 26912, 28208, 28096, 28064, 27952,
	29248, 29136, 29104, 28992, 30288, 30176, 30144, 30032, 31328, 31216, 31184, 31072, 32368, 32224, 32112, 33408,
	33296, 33152, 34448, 34336, 34304, 34192, 35488, 35376, 35344, 35232, 36528, 36416, 36384, 36272, 37568, 37456,
	37424, 37312, 38608, 38496, 38464, 38352, 39648, 39536, 39504, 39392, 40688, 40544, 40432, 41728, 41616, 41472,
	42768, 42656, 42624, 42512, 43808, 43696, 43664, 43552, 44848, 44736, 44704, 44592, 45888, 45776, 45744, 45632,
	46928, 46816, 46784, 46672, 47968, 47856, 47824, 47712, 49008, 48864, 48752, 50048, 49936, 49792, 51088, 50976,
	50944, 50832, 52128, 52016, 51984, 51872, 53168, 53056, 53024, 52912, 54208, 54096, 54064, 53952, 55248, 55136,
	55104, 54992, 56288, 56176, 56144, 56032, 57328, 57184, 57072, 58256, 58112, 59296, 59264, 59152, 60336, 60304,
	60192, 61376, 61344, 61232, 62416, 62384, 62272, 63456, 63424, 63312, 64496, 64464, 64352, 65504, 65392, 449,
	450, 451, 452, 453, 454,
}

// MakeStepFromIndex returns the step from the compact index or ok = false.
// If i is the pass value, pass will be true.
// Captures are unmapped.
func MakeStepFromIndex(p *Pos, i uint8) (s Step, pass, ok bool) {
	if i < passIndex {
		s := stepTable[i]
		if !s.Setup() {
			s |= Step(p.At(s.Src()))
		}
		return s, false, true
	}
	if i == passIndex {
		return 0, true, true
	}
	return 0, false, false
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
	return s&0b1111110000000000 == 0b0111000000000000 && // dest == E4
		s&0b1111110000 != 0b0101000000 && // src != E3
		s&0b1111110000 != 0b1001000000 && // src != E5
		s&0b1111110000 != 0b0110110000 && // src != D4
		s&0b1111110000 != 0b0111010000 // src != F4
}

// Setup returns true if the Step is a setup Step.
func (s Step) Setup() bool {
	return s&0b1111110000 == 0b0111000000 && // src == E4
		s&0b1111110000000000 != 0b0101000000000000 && // dest != E3
		s&0b1111110000000000 != 0b1001000000000000 && // dest != E5
		s&0b1111110000000000 != 0b0110110000000000 && // dest != D4
		s&0b1111110000000000 != 0b0111010000000000 // dest != F4
}

// Recurring returns true if playing s and step lead to a recurring position (i.e. step "undoes" s).
func (s Step) Recurring(step Step) bool {
	return s.Piece() == step.Piece() && s.Src() == step.Dest() && s.Dest() == step.Src()
}

// Flip flips this step across rank and file axes.
func (s Step) Flip() Step {
	switch {
	case s.Setup():
		return MakeSetup(s.Piece(), s.Dest().Flip())
	case s.Capture():
		return MakeCapture(s.Piece(), s.Src().Flip())
	default:
		return MakeStep(s.Piece(), s.Src().Flip(), s.Dest().Flip())
	}
}

// MirrorLateral mirrors this step across file axis.
// Used for dataset augmentation.
func (s Step) MirrorLateral() Step {
	switch {
	case s.Setup():
		return MakeSetup(s.Piece(), s.Dest().MirrorLateral())
	case s.Capture():
		return MakeCapture(s.Piece(), s.Src().MirrorLateral())
	default:
		return MakeStep(s.Piece(), s.Src().MirrorLateral(), s.Dest().MirrorLateral())
	}
}

// RemoveColor returns this step with color data removed (same as gold pieces).
func (s Step) RemoveColor() Step {
	return s & 0b1111111111110111
}

const passIndex = 231

// Index returns the computed compact index for the step.
// Capture indices are undefined.
func (s Step) Index() uint8 {
	var i uint8
	if s.Setup() {
		return 224 + uint8(s.Piece().RemoveColor())
	}
	src, dest := s.Src(), s.Dest()
	for j := North; j > DirNone; j-- {
		v := dest.Add(j)
		if v == src {
			break
		}
		if v.Valid() {
			i++
		}
	}
	switch {
	case dest <= H1:
		switch {
		case dest == A1:
		default:
			i += 2 + 3*(uint8(dest)-1)
		}
	case dest < A8:
		i += 22 + 30*uint8(dest.Rank()-1)
		switch f := dest.File(); {
		case f == 0:
		default:
			i += 3 + 4*(f-1)
		}
	case dest < H8:
		i += 202
		switch f := dest.File(); {
		case f == 0:
		default:
			i += 2 + 3*(f-1)
		}
	default:
		i += 222
	}
	return i
}

// DebugCaptureContext returns the capture resulting from playing s.
// Intended for debugging only.
func (s Step) DebugCaptureContext(p *Pos) (cap Step) {
	p = p.Clone()
	if !s.Capture() {
		cap = p.Step(s)
	}
	return cap
}

// appendString appends the Step string to the builder.
func (s Step) appendString(sb *strings.Builder) {
	if !s.Piece().Valid() {
		s.appendGoString("Invalid", sb)
		return
	}
	sb.WriteByte(s.Piece().Byte())
	if s.Setup() {
		sb.WriteString(s.Dest().String())
		return
	}
	src := s.Src()
	sb.WriteString(src.String())
	if s.Capture() {
		sb.WriteByte('x')
		return
	}
	sb.WriteByte(s.Dest().Sub(src).Byte())
}

// String returns the String representation of the Step and possible capture.
func (s Step) String() string {
	var sb strings.Builder
	s.appendString(&sb)
	return sb.String()
}

func (s Step) appendGoString(prefix string, sb *strings.Builder) {
	fmt.Fprintf(sb, "%s(piece=%c src=%s, dest=%s)", prefix, s.Piece().Byte(), s.Src(), s.Dest())
}

// GoString returns the formatted GoString for this step.
func (s Step) GoString() string {
	var sb strings.Builder
	s.appendGoString("Step", &sb)
	return sb.String()
}
