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
	d1 := s1.Translate(ParseDelta(data[indices[0]+3]))
	d2 := s2.Translate(ParseDelta(data[indices[2]+3]))
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
	res = append(res, Step{Pass: true})
	return res, nil
}

// MoveString outputs a legal move string.
func MoveString(move []Step) string {
	var sb strings.Builder
	for i, step := range move {
		if step.Pass {
			continue
		}
		sb.WriteString(step.String())
		if i+1 < len(move) {
			sb.WriteByte(' ')
		}
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

type Capture struct {
	Piece
	Src Square
}

func (c Capture) Valid() bool {
	return c.Piece != Empty
}

type StepKind uint8

const (
	KindInvalid StepKind = iota
	KindDefault
	KindSetup
	KindPush
	KindPull
)

type Step struct {
	Src, Dest, Alt Square
	Piece1, Piece2 Piece
	Cap            Capture
	Pass           bool
}

var invalidStep = Step{
	Src:  invalidSquare,
	Dest: invalidSquare,
	Alt:  invalidSquare,
}

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
		return Step{
			Src:    invalidSquare,
			Dest:   invalidSquare,
			Alt:    src1,
			Piece1: piece1,
		}, nil
	}
	delta1 := ParseDelta(s[3])
	if delta1 == 0 {
		return invalidStep, fmt.Errorf("invalid first delta: %c", s[3])
	}
	dest1 := src1.Translate(delta1)
	// Return single step:
	if len(s) == 4 {
		return Step{
			Src:    src1,
			Dest:   dest1,
			Alt:    invalidSquare,
			Piece1: piece1,
		}, nil
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
	cap1 := Capture{}
	delta2 := ParseDelta(s[8])
	if s[8] == 'x' {
		if !piece1.SameColor(piece2) {
			return invalidStep, fmt.Errorf("invalid first capture color: %q", s)
		}
		cap1.Piece = piece2
		cap1.Src = src2
	} else if delta2 == 0 {
		return invalidStep, fmt.Errorf("invalid second delta: %c", s[8])
	}
	dest2 := src2.Translate(delta2)

	// Return default self capture:
	if cap1.Valid() && len(s) == 9 {
		return Step{
			Src:    src1,
			Dest:   dest1,
			Alt:    invalidSquare,
			Piece1: piece1,
			Cap:    cap1,
		}, nil
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
	alt := src2
	step := Step{
		Src:    src1,
		Dest:   dest1,
		Alt:    alt,
		Piece1: piece1,
		Piece2: piece2,
	}

	// No capture:
	if len(s) == 9 {
		return step, nil
	}

	piece3, err := ParsePiece(s[10])
	if err != nil {
		return invalidStep, err
	}
	src3 := ParseSquare(s[11:13])
	if !src3.Valid() {
		return invalidStep, fmt.Errorf("invalid capture square: %q", s[11:13])
	}
	if s[13] != 'x' {
		return invalidStep, fmt.Errorf("invalid capture delta: %c", s[13])
	}
	step.Cap = Capture{
		Piece: piece3,
		Src:   src3,
	}
	return step, nil
}

func (s Step) Kind() StepKind {
	sv, dv := s.Src.Valid(), s.Dest.Valid()
	p1e, p2e := s.Piece1 == Empty, s.Piece2 == Empty
	switch {
	case s.Alt.Valid():
		switch {
		case sv && dv:
			switch {
			case p1e || p2e:
				return KindInvalid
			case s.Dest.AdjacentTo(s.Alt):
				return KindPush
			case s.Src.AdjacentTo(s.Alt):
				return KindPull
			default:
				return KindInvalid
			}
		case !p1e:
			return KindSetup
		default:
			return KindInvalid
		}
	case !p1e && sv && dv:
		return KindDefault
	default:
		return KindInvalid
	}
}

func (s Step) Capture() bool {
	return s.Cap.Piece != Empty
}

func (s Step) Len() int {
	switch kind := s.Kind(); {
	case s.Pass:
		return 0
	case kind == KindDefault, kind == KindSetup:
		return 1
	case kind == KindPush, kind == KindPull:
		return 2
	default:
		return 0
	}
}

func (s Step) String() string {
	var sb strings.Builder
	switch s.Kind() {
	case KindSetup:
		fmt.Fprintf(&sb, "%c%s", s.Piece1.Byte(), s.Alt)
		if s.Capture() {
			fmt.Fprintf(&sb, " %c?%sx", s.Cap.Piece.Byte(), s.Cap.Src)
		}
	case KindPush:
		fmt.Fprintf(&sb, "%c%s%s", s.Piece2.Byte(), s.Dest, NewDelta(s.Dest.Delta(s.Alt)))
		if s.Cap.Piece == s.Piece2 {
			fmt.Fprintf(&sb, " %c%sx ", s.Cap.Piece.Byte(), s.Cap.Src)
		}
		fmt.Fprintf(&sb, "%c%s%s", s.Piece1.Byte(), s.Src, NewDelta(s.Src.Delta(s.Dest)))
		if s.Cap.Piece == s.Piece1 {
			fmt.Fprintf(&sb, " %c%sx", s.Cap.Piece.Byte(), s.Cap.Src)
		}
		if s.Capture() && s.Cap.Piece != s.Piece1 && s.Cap.Piece != s.Piece2 {
			fmt.Fprintf(&sb, " %c?%sx", s.Cap.Piece.Byte(), s.Cap.Src)
		}
	case KindPull:
		fmt.Fprintf(&sb, "%c%s%s", s.Piece1.Byte(), s.Src, NewDelta(s.Src.Delta(s.Dest)))
		if s.Cap.Piece == s.Piece1 {
			fmt.Fprintf(&sb, " %c%sx", s.Cap.Piece.Byte(), s.Cap.Src)
		}
		fmt.Fprintf(&sb, " %c%s%s", s.Piece2.Byte(), s.Alt, NewDelta(s.Alt.Delta(s.Src)))
		if s.Cap.Piece == s.Piece2 {
			fmt.Fprintf(&sb, " %c%sx", s.Cap.Piece.Byte(), s.Cap.Src)
		}
		if s.Capture() && s.Cap.Piece != s.Piece1 && s.Cap.Piece != s.Piece2 {
			fmt.Fprintf(&sb, " %c?%sx", s.Cap.Piece.Byte(), s.Cap.Src)
		}
	case KindDefault:
		fmt.Fprintf(&sb, "%c%s%s", s.Piece1.Byte(), s.Src, NewDelta(s.Src.Delta(s.Dest)))
		if s.Capture() {
			fmt.Fprintf(&sb, " %c%sx", s.Cap.Piece.Byte(), s.Cap.Src)
		}
	default: // Invalid
		if s.Pass {
			fmt.Fprint(&sb, "(pass)")
			if s.Capture() {
				fmt.Fprintf(&sb, " %c?%sx", s.Cap.Piece.Byte(), s.Cap.Src)
			}
		}
		return s.GoString()
	}
	if s.Capture() {
		fmt.Fprintf(&sb, " %c%sx", s.Cap.Piece.Byte(), s.Cap.Src)
	}
	return sb.String()
}

func (s Step) GoString() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Step(src=%s, dest=%s, alt=%s, piece1=%s, piece2=%s)", s.Src, s.Dest, s.Alt, s.Piece1, s.Piece2)
	if s.Capture() {
		fmt.Fprintf(&sb, " %c%sx", s.Cap.Piece.Byte(), s.Cap.Src)
	}
	return sb.String()
}
