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
	138, 26, 1178, 1066, 1034, 2218, 2106, 2074, 3258, 3146, 3114, 4298, 4186, 4154, 5338, 5226, 5194,
	6378, 6266, 6234, 7418, 7274, 8458, 8346, 8202, 9498, 9386, 9354, 9242, 10538, 10426, 10394, 10282,
	11578, 11466, 11434, 11322, 12618, 12506, 12474, 12362, 13658, 13546, 13514, 13402, 14698, 14586, 14554, 14442,
	15738, 15594, 15482, 16778, 16666, 16522, 17818, 17706, 17674, 17562, 18858, 18746, 18714, 18602, 19898, 19786,
	19754, 19642, 20938, 20826, 20794, 20682, 21978, 21866, 21834, 21722, 23018, 22906, 22874, 22762, 24058, 23914,
	23802, 25098, 24986, 24842, 26138, 26026, 25994, 25882, 27178, 27066, 27034, 26922, 28218, 28106, 28074, 27962,
	29258, 29146, 29114, 29002, 30298, 30186, 30154, 30042, 31338, 31226, 31194, 31082, 32378, 32234, 32122, 33418,
	33306, 33162, 34458, 34346, 34314, 34202, 35498, 35386, 35354, 35242, 36538, 36426, 36394, 36282, 37578, 37466,
	37434, 37322, 38618, 38506, 38474, 38362, 39658, 39546, 39514, 39402, 40698, 40554, 40442, 41738, 41626, 41482,
	42778, 42666, 42634, 42522, 43818, 43706, 43674, 43562, 44858, 44746, 44714, 44602, 45898, 45786, 45754, 45642,
	46938, 46826, 46794, 46682, 47978, 47866, 47834, 47722, 49018, 48874, 48762, 50058, 49946, 49802, 51098, 50986,
	50954, 50842, 52138, 52026, 51994, 51882, 53178, 53066, 53034, 52922, 54218, 54106, 54074, 53962, 55258, 55146,
	55114, 55002, 56298, 56186, 56154, 56042, 57338, 57194, 57082, 58266, 58122, 59306, 59274, 59162, 60346, 60314,
	60202, 61386, 61354, 61242, 62426, 62394, 62282, 63466, 63434, 63322, 64506, 64474, 64362, 65514, 65402, 449,
	450, 451, 452, 453, 454,
}

// MakeStepFromIndex returns the step from the compact index or ok = false.
// If i is the pass value, pass will be true.
// Captures are unmapped.
func MakeStepFromIndex(i uint8) (s Step, pass, ok bool) {
	if i < passIndex {
		return stepTable[i], false, true
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
