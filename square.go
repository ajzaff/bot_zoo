package zoo

import "fmt"

// Square represents a square on the Arimaa board as an ordinal index in the range 0-63.
type Square uint8

// Square constants.
const (
	A1 Square = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
)

// ParseSquare parses the 2-byte square string or returns an error.
func ParseSquare(s string) (Square, error) {
	if len(s) != 2 {
		return 64, fmt.Errorf("wrong number of bytes: %v", s)
	}
	r, f := s[1]-'1', s[0]-'a'
	if r < 0 || r > 7 || f < 0 || f > 7 {
		return 64, fmt.Errorf("bad rank or file: %v", s)
	}
	return Square(8*r + f), nil
}

// Rank returns the rank of i assuming i is Valid.
func (i Square) Rank() uint8 {
	return uint8(i >> 3)
}

// File returns the file of i assuming i is Valid.
func (i Square) File() uint8 {
	return uint8(i & 0b111)
}

// Add returns the Square at `i + d` if valid.
// Adding DirNone results in an invalid Square.
func (i Square) Add(d Direction) Square {
	if i.Valid() {
		if v := i + Square(d.Value()); v.Valid() && (i.Rank() == v.Rank()) != (i.File() == v.File()) {
			return v
		}
	}
	return 64
}

// Sub returns the result of `i - j` or DirNone.
// It returns DirNone if either i or j is invalid.
func (i Square) Sub(j Square) Direction {
	if i.Valid() && j.Valid() {
		v := int8(i) - int8(j)
		for d, x := range dirValues {
			if v == x {
				return Direction(d)
			}
		}
	}
	return DirNone
}

// Valid returns true if i is a valid square on the board.
func (i Square) Valid() bool {
	return i < 64
}

// Bitboard returns a bitboard mask for i.
func (i Square) Bitboard() Bitboard {
	return Bitboard(1) << i
}

var (
	neighbors        [64]Bitboard
	forwardNeighbors [128]Bitboard
)

func init() {
	for i := Square(0); i < 64; i++ {
		b := i.Bitboard()
		steps := b.Neighbors()
		neighbors[i] = steps
		grSteps := steps
		if b&NotRank1 != 0 { // rabbits can't move backwards.
			grSteps ^= b >> 8
		}
		srSteps := steps
		if b&NotRank8 != 0 {
			srSteps ^= b << 8
		}
		forwardNeighbors[i] = grSteps
		forwardNeighbors[64+i] = srSteps
	}
}

// Neighbors returns the neighbors of the square.
// It panics if i is invalid.
func (i Square) Neighbors() Bitboard {
	return neighbors[i]
}

// ForwardNeighbors returns the forward neighbors of the square for the given color c.
// It panics if i is invalid.
func (i Square) ForwardNeighbors(c Color) Bitboard {
	return forwardNeighbors[64*uint8(c)+uint8(i)]
}

// AdjacentTrap returns the Trap Square adjacent to i if present.
func (i Square) AdjacentTrap() Square {
	if i.Valid() {
		b := i.Neighbors() & Traps
		if b != 0 {
			return b.Square()
		}
	}
	return 64
}

// Trap returns true if i is a trap Square.
func (i Square) Trap() bool {
	return i == C3 || i == F6 || i == C6 || i == F3
}

// Mirror returns the square mirrored acoss the rank and file axes (e.g. c1 <=> f8)
func (i Square) Mirror() Square {
	if i.Valid() {
		return 63 - i
	}
	return 64
}

// Flip returns the square mirrored acoss the rank axis (e.g. c1 <=> c8)
func (i Square) Flip() Square {
	if i.Valid() {
		return Square(8*(7-i.Rank()) + i.File())
	}
	return 64
}

const squareNames = "a1b1c1d1e1f1g1h1a2b2c2d2e2f2g2h2a3b3c3d3e3f3g3h3a4b4c4d4e4f4g4h4a5b5c5d5e5f5g5h5a6b6c6d6e6f6g6h6a7b7c7d7e7f7g7h7a8b8c8d8e8f8g8h8"

// String returns the string representation of this square if valid.
func (i Square) String() string {
	if i.Valid() {
		return squareNames[2*i : 2*i+2]
	}
	return ""
}

// Direction represents a move direction in Arimaa.
type Direction int8

// Direction constants.
const (
	DirNone Direction = iota
	North
	East
	South
	West
)

const dirBytes = "xnesw"

// ParseDir parses the Direction from print representation or returns DirNone.
func ParseDir(b byte) Direction {
	for i, x := range []byte(dirBytes) {
		if b == x {
			return Direction(i)
		}
	}
	return DirNone
}

// Valid returns true if d is a valid cardinal direction.
func (d Direction) Valid() bool {
	return d < 5 && d != 0
}

var dirValues = []int8{0, 8, 1, -8, -1}

// Value returns the offset value of this direction.
func (d Direction) Value() int8 {
	if d.Valid() {
		return dirValues[d]
	}
	return 0
}

// Byte returns the printable byte for d.
func (d Direction) Byte() byte {
	if d.Valid() {
		return dirBytes[d]
	}
	return 0
}
