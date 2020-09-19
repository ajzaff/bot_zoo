package zoo

import "fmt"

// Piece represents a game piece as a byte.
// Only 3 least-significant bits are used.
type Piece uint8

// Piece constants.
const (
	Empty Piece = iota
	GRabbit
	GCat
	GDog
	GHorse
	GCamel
	GElephant
	_
	_
	SRabbit
	SCat
	SDog
	SHorse
	SCamel
	SElephant
)

const pieceBytes = " RCDHMExxrcdhme"

// ParsePiece parses the byte to a Piece or returns an error.
func ParsePiece(b byte) (Piece, error) {
	for i, x := range []byte(pieceBytes) {
		if b == x {
			return Piece(i), nil
		}
	}
	return 0, fmt.Errorf("not valid: %c", b)
}

// Color returns the color of this piece.
func (p Piece) Color() Color {
	return Color((p & 0b1000) >> 3)
}

// WithColor returns the Piece p  with the Color c.
func (p Piece) WithColor(c Color) Piece {
	return Piece(uint8(p&0b011) | uint8(c<<3))
}

// RemoveColor returns the Piece p if it were Gold.
func (p Piece) RemoveColor() Piece {
	return p & 0b0111
}

// SameType returns true when p and piece have the same Piece type.
func (p Piece) SameType(piece Piece) bool {
	return p.RemoveColor() == piece.RemoveColor()
}

// SameColor returns true when p and piece have the same Color.
func (p Piece) SameColor(piece Piece) bool {
	return p&0b1000 == piece&0b1000
}

// WeakerThan returns true if p is strictly weaker than piece.
func (p Piece) WeakerThan(piece Piece) bool {
	return p.RemoveColor() < piece.RemoveColor()
}

// Valid returns whether p is a valid piece.
func (p Piece) Valid() bool {
	return p < 15 && p != 0 && p != 7 && p != 8
}

// Byte returns the printable representation of the Piece.
func (p Piece) Byte() byte {
	if p < 15 {
		return pieceBytes[p]
	}
	return 0
}

// Color defines the piece color enum.
type Color uint8

// Color constants.
const (
	Gold Color = iota
	Silver
)

// ParseColor parses a color byte or returns an error.
// It supports the legacy white and black colors.
func ParseColor(s byte) (Color, error) {
	switch s {
	case 'w', 'g':
		return Gold, nil
	case 'b', 's':
		return Silver, nil
	default:
		return 0, fmt.Errorf("failed to parse color: %c", s)
	}
}

// Valid returns true if c is either Gold or Silver.
func (c Color) Valid() bool {
	return c <= Silver
}

// Opposite returns the opposite color of c.
// This method assumes c is a Valid Color.
func (c Color) Opposite() Color {
	return c ^ 1
}

// Byte returns the byte for this Color assuming c is Valid.
func (c Color) Byte() byte {
	if c.Valid() {
		return "gs"[c]
	}
	return 0
}
