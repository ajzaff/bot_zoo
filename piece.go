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

type Piece int

const PChars = " RCDHMExxrcdhme"

func ParsePiece(s string) (Piece, error) {
	for i, r := range PChars {
		if s == string(r) {
			return Piece(i), nil
		}
	}
	return Empty, fmt.Errorf("input does not match /^[%s]$/", PChars)
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
	return p | c.PieceMask()
}

func (p Piece) SamePiece(piece Piece) bool {
	return p&decolorMask == piece&decolorMask
}

func (p Piece) WeakerThan(piece Piece) bool {
	return p&decolorMask < piece&decolorMask
}
