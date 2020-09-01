package zoo

import (
	"fmt"
	"strings"
)

// ParseMove parses the move string into steps
// and checks for validity but not legality.
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

// MoveString outputs a legal move string by adding captures if missing.
func (p *Pos) MoveString(move []Step) string {
	var sb strings.Builder
	for i, step := range move {
		sb.WriteString(step.String())
		if i+1 < len(move) {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}

type Capture struct {
	Piece
	Src Square
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
		at1 := ParseSquare(s[1:3])
		if !at1.Valid() {
			return invalidStep, fmt.Errorf("invalid first square: %q", s)
		}
		dest1 := at1.Translate(ParseDelta(s[3:4]))
		p2, err := ParsePiece(s[4:5])
		if err != nil {
			return invalidStep, err
		}
		at2 := ParseSquare(s[6:7])
		if !at2.Valid() {
			return invalidStep, fmt.Errorf("invalid second square: %q", s)
		}
		dest2 := at2.Translate(ParseDelta(string(s[8])))
		if p1&decolorMask < p2&decolorMask {
			p1, p2 = p2, p1
			at1, at2 = at2, at1
			dest1, dest2 = dest2, dest1
		}
		return Step{
			Src:    at1,
			Dest:   dest1,
			Alt:    at2,
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
	if s.Alt.Valid() {
		if sv && dv {
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
		}
		if p1e {
			return KindInvalid
		}
		return KindSetup
	}
	if p1e {
		return KindInvalid
	}
	return KindDefault
}

func (s Step) Capture() bool {
	return s.Cap.Piece != Empty
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
	default:
		fmt.Fprintf(&sb, "InvalidStep(src=%s, dest=%s, alt=%s, piece1=%s, piece2=%s)", s.Src, s.Dest, s.Alt, s.Piece1, s.Piece2)
	}
	if s.Capture() {
		fmt.Fprintf(&sb, " %c%sx", s.Cap.Piece.Byte(), s.Cap.Src)
	}
	return sb.String()
}
