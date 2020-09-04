package zoo

import (
	"bufio"
	"fmt"
	"log"
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
	switch n := len(indices); {
	case n <= 1: // EOF
		if atEOF {
			return len(data), nil, nil
		}
		// Request more
		return 0, nil, nil
	case n == 2:
		if atEOF {
			return len(data), data[indices[0]:indices[1]], nil
		}
		if indices[1]-indices[0] <= 3 {
			return len(data), data[indices[0]:indices[1]], nil
		}
		// Request more
		return 0, nil, nil
	case n == 4:
		if indices[1]-indices[0] <= 3 {
			return indices[1], data[indices[0]:indices[1]], nil
		}
		p1, _ := ParsePiece(string(data[indices[0]]))
		p2, _ := ParsePiece(string(data[indices[2]]))
		s1 := ParseSquare(string(data[indices[0]+1 : indices[0]+3]))
		s2 := ParseSquare(string(data[indices[2]+1 : indices[2]+3]))
		d1 := ParseSquare(string(data[indices[0]+1 : indices[0]+3])).Translate(ParseDelta(string(indices[0] + 3)))
		d2 := ParseSquare(string(data[indices[2]+1 : indices[2]+3])).Translate(ParseDelta(string(indices[2] + 3)))
		cap := data[indices[2]+3] == 'x'
		if cap {
			return len(data), data[indices[0]:indices[3]], nil
		}
		if p1.Color() != p2.Color() && p1&decolorMask < p2&decolorMask && (d1 == s2 || d2 == s1) {
			return len(data), data[indices[0]:indices[3]], nil
		}
		return indices[1], data[indices[0]:indices[1]], nil
	case n >= 6:
		if indices[1]-indices[0] <= 3 {
			return indices[1], data[indices[0]:indices[1]], nil
		}
		p1, _ := ParsePiece(string(data[indices[0]]))
		p2, _ := ParsePiece(string(data[indices[2]]))
		s1 := ParseSquare(string(data[indices[0]+1 : indices[0]+3]))
		s2 := ParseSquare(string(data[indices[2]+1 : indices[2]+3]))
		d1 := ParseSquare(string(data[indices[0]+1 : indices[0]+3])).Translate(ParseDelta(string(indices[0] + 3)))
		d2 := ParseSquare(string(data[indices[2]+1 : indices[2]+3])).Translate(ParseDelta(string(indices[2] + 3)))
		cap := data[indices[2]+3] == 'x'
		if cap {
			return indices[3], data[indices[0]:indices[3]], nil
		}
		if p1.Color() == p2.Color() {
			return indices[1], data[indices[0]:indices[1]], nil
		}
		if p1&decolorMask < p2&decolorMask && (d1 == s2 || d2 == s1) {
			return indices[1], data[indices[0]:indices[1]], nil
		}
		cap2 := data[indices[4]+3] == 'x'
		if !cap2 {
			return indices[3], data[indices[0]:indices[3]], nil
		}
		return indices[5], data[indices[0]:indices[5]], nil
	default:
		return 0, nil, fmt.Errorf("unexpected input: %q", string(data))
	}
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
		log.Printf("%q\n", sc.Text())
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
// TODO(ajzaff): Clean this up.
func ParseStep(s string) (Step, error) {
	switch {
	case len(s) == 3: // Setup:
		piece, err := ParsePiece(s[0:1])
		if err != nil {
			return invalidStep, err
		}
		alt := ParseSquare(s[1:3])
		if !alt.Valid() {
			return invalidStep, fmt.Errorf("invalid setup square: %q", s)
		}
		return Step{
			Src:    invalidSquare,
			Dest:   invalidSquare,
			Alt:    alt,
			Piece1: piece,
		}, nil
	case len(s) == 4: // Default:
		piece, err := ParsePiece(string(s[0]))
		if err != nil {
			return invalidStep, err
		}
		src := ParseSquare(s[1:3])
		if !src.Valid() {
			return invalidStep, fmt.Errorf("invalid step square: %q", s)
		}
		if s[3:4] == "x" { // Lone capture:
			return Step{
				Src:  invalidSquare,
				Dest: invalidSquare,
				Alt:  invalidSquare,
				Cap: Capture{
					Piece: piece,
					Src:   src,
				},
			}, nil
		}
		dest := src.Translate(ParseDelta(string(s[3])))
		return Step{
			Src:    src,
			Dest:   dest,
			Alt:    invalidSquare,
			Piece1: piece,
		}, nil
	case len(s) == 9: // Push/Pull or Default capture:
		if strings.HasSuffix(s, "x") { // Default capture:
			piece, err := ParsePiece(s[0:1])
			if err != nil {
				return invalidStep, err
			}
			src := ParseSquare(s[1:3])
			if !src.Valid() {
				return invalidStep, fmt.Errorf("invalid step square: %q", s)
			}
			dest := src.Translate(ParseDelta(s[3:4]))
			step := Step{
				Src:    src,
				Dest:   dest,
				Alt:    invalidSquare,
				Piece1: piece,
			}
			capPiece, err := ParsePiece(s[5:6])
			if err != nil {
				return invalidStep, err
			}
			capSrc := ParseSquare(s[6:8])
			if !src.Valid() {
				return invalidStep, fmt.Errorf("invalid capture square: %q", s)
			}
			step.Cap = Capture{
				Piece: capPiece,
				Src:   capSrc,
			}
			return step, nil
		}
		p1, err := ParsePiece(s[0:1])
		if err != nil {
			return invalidStep, err
		}
		src1 := ParseSquare(s[1:3])
		if !src1.Valid() {
			return invalidStep, fmt.Errorf("invalid first square: %q", s)
		}
		dest1 := src1.Translate(ParseDelta(s[3:4]))
		p2, err := ParsePiece(s[5:6])
		if err != nil {
			return invalidStep, err
		}
		src2 := ParseSquare(s[6:8])
		if !src2.Valid() {
			return invalidStep, fmt.Errorf("invalid second square: %q", s)
		}
		dest2 := src2.Translate(ParseDelta(string(s[8])))
		if p1&decolorMask < p2&decolorMask {
			p1, p2 = p2, p1
			src1 = src2
			dest1, dest2 = dest2, dest1
		}
		return Step{
			Src:    src1,
			Dest:   dest1,
			Alt:    dest2,
			Piece1: p1,
			Piece2: p2,
		}, nil
	case len(s) == 14: // Push/Pull with capture:
		p1, err := ParsePiece(s[0:1])
		if err != nil {
			return invalidStep, err
		}
		at1 := ParseSquare(s[1:3])
		if !at1.Valid() {
			return invalidStep, fmt.Errorf("invalid first square: %q", s)
		}
		dest1 := at1.Translate(ParseDelta(s[3:4]))
		p2, err := ParsePiece(s[5:6])
		if err != nil {
			return invalidStep, err
		}
		at2 := ParseSquare(s[6:8])
		if !at2.Valid() {
			return invalidStep, fmt.Errorf("invalid second square: %q", s)
		}
		dest2 := at2.Translate(ParseDelta(s[8:9]))
		if p1&decolorMask < p2&decolorMask {
			p1, p2 = p2, p1
			at1, at2 = at2, at1
			dest1, dest2 = dest2, dest1
		}
		step := Step{
			Src:    at1,
			Dest:   dest1,
			Alt:    at2,
			Piece1: p1,
			Piece2: p2,
		}
		capPiece, err := ParsePiece(s[9:10])
		if err != nil {
			return invalidStep, err
		}
		capSrc := ParseSquare(s[10:12])
		if !capSrc.Valid() {
			return invalidStep, fmt.Errorf("invalid capture square: %q", s)
		}
		if s[12] != 'x' {
			return invalidStep, fmt.Errorf("invalid capture: %q", s)
		}
		step.Cap = Capture{
			Piece: capPiece,
			Src:   capSrc,
		}
		return step, nil
	default:
		return invalidStep, fmt.Errorf("malformed step: %q", s)
	}
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
			case s.Piece1&decolorMask < s.Piece2&decolorMask:
				return KindPull
			case s.Piece1&decolorMask > s.Piece2&decolorMask:
				return KindPush
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
	case KindPush:
		fmt.Fprintf(&sb, "%c%s%s %c%s%s",
			s.Piece1.Byte(), s.Dest, NewDelta(s.Dest.Delta(s.Alt)),
			s.Piece2.Byte(), s.Src, NewDelta(s.Src.Delta(s.Dest)),
		)
	case KindPull:
		fmt.Fprintf(&sb, "%c%s%s %c%s%s",
			s.Piece1.Byte(), s.Src, NewDelta(s.Src.Delta(s.Dest)),
			s.Piece2.Byte(), s.Alt, NewDelta(s.Alt.Delta(s.Src)),
		)
	case KindDefault:
		fmt.Fprintf(&sb, "%c%s%s", s.Piece1.Byte(), s.Src, NewDelta(s.Src.Delta(s.Dest)))
	default: // Invalid
		if s.Pass {
			return "(pass)"
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
