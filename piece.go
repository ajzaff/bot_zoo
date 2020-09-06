package zoo

import "fmt"

type Color int

const (
	Gold Color = iota
	Silver
)
const (
	colorMask   = 8
	decolorMask = ^colorMask
)

func ParseColor(s string) Color {
	switch s {
	case "w", "g":
		return Gold
	case "b", "s":
		return Silver
	default:
		return -1
	}
}

func (c Color) Valid() bool {
	return c == Gold || c == Silver
}

func (c Color) Opposite() Color {
	return c ^ 1
}

func (c Color) PieceMask() Piece {
	return Piece(c << 3)
}

func (c Color) Byte() byte {
	if c == Gold {
		return 'g'
	}
	return 's'
}

func (c Color) String() string {
	return string(c.Byte())
}

type Piece int

const pchars = " RCDHMExxrcdhme"

func ParsePiece(b byte) (Piece, error) {
	for i, x := range []byte(pchars) {
		if b == x {
			return Piece(i), nil
		}
	}
	return Empty, fmt.Errorf("input does not match /^[%s]$/", pchars)
}

const (
	Empty Piece = iota
	GRabbit
	GCat
	GDog
	GHorse
	GCamel
	GElephant
)

const (
	SRabbit Piece = iota + 9
	SCat
	SDog
	SHorse
	SCamel
	SElephant
)

func (p Piece) Color() Color {
	if p&colorMask == 0 {
		return Gold
	}
	return Silver
}

func (p Piece) MakeColor(c Color) Piece {
	return p&decolorMask | c.PieceMask()
}

func (p Piece) SameType(piece Piece) bool {
	return p&decolorMask == piece&decolorMask
}

func (p Piece) HasColor() bool {
	return p > 0 && p < 15 && p != 7 && p != 8
}

func (p Piece) SameColor(piece Piece) bool {
	return p.HasColor() && piece.HasColor() && p&colorMask == piece&colorMask
}

func (p Piece) WeakerThan(piece Piece) bool {
	return p&decolorMask < piece&decolorMask
}

func (p Piece) Valid() bool {
	return p < 15 && p != 7 && p != 8
}

func (p Piece) validForPrint() bool {
	return p < 15
}

func (p Piece) Byte() byte {
	if p.validForPrint() {
		return pchars[p]
	}
	return '?'
}

func (p Piece) String() string {
	if p.validForPrint() {
		return string(p.Byte())
	}
	return fmt.Sprintf("Piece(%d)", p)
}
