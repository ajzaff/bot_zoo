package zoo

import (
	"fmt"
	"strconv"
)

const (
	ranks = "12345678"
	files = "abcdefgh"
)

type Square uint8

func ParseSquare(s string) Square {
	return Square(s[0] - 'a' + 8*(s[1]-'1'))
}

func (i Square) Valid() bool {
	return i < 64
}

func (i Square) Bitboard() Bitboard {
	return Bitboard(1) << i
}

func (src Square) Delta(dest Square) Delta {
	if src == dest {
		return 0
	}
	sb, db := src.Bitboard(), dest.Bitboard()
	switch {
	case sb&NotFileA>>1 == db:
		return -1
	case sb&NotRank1>>8 == db:
		return -8
	case sb&NotFileH<<1 == db:
		return 1
	case sb&NotRank8<<8 == db:
		return 8
	default:
		return 0
	}
}

func (i Square) AdjacentTrap() Square {
	if i.Valid() {
		b := stepsB[i] & Traps
		if b != 0 {
			return b.Square()
		}
	}
	return invalidSquare
}

func (i Square) Trap() bool {
	return i == 18 || i == 21 || i == 42 || i == 45
}

func (i Square) AdjacentTo(j Square) bool {
	return i.Valid() && j.Valid() && i.Delta(j) != 0
}

func (i Square) String() string {
	if i.Valid() {
		return string([]byte{
			files[i%8],
			ranks[i/8],
		})
	}
	return strconv.Itoa(int(i))
}

const invalidSquare Square = 255

type Delta int8

const (
	invalidDelta       = 0
	invalidPackedDelta = 7
)

func MakeDeltaFromPacked(p uint8) Delta {
	switch p {
	case 0:
		return +8
	case 1:
		return -8
	case 2:
		return +1
	case 3:
		return -1
	default:
		return invalidDelta
	}
}

func MakeDeltaFromByte(b byte) Delta {
	switch b {
	case 'n':
		return +8
	case 's':
		return -8
	case 'e':
		return +1
	case 'w':
		return -1
	default:
		return 0
	}
}

func (d Delta) Packed() uint8 {
	switch d {
	case +8:
		return 0
	case -8:
		return 1
	case 1:
		return 2
	case -1:
		return 3
	default:
		return invalidPackedDelta
	}
}

func (d Delta) String() string {
	switch d {
	case +8:
		return "n"
	case -8:
		return "s"
	case 1:
		return "e"
	case -1:
		return "w"
	default:
		return fmt.Sprintf("?(%d)", d)
	}
}

func (i Square) Translate(d Delta) Square {
	if !i.Valid() {
		return invalidSquare
	}
	switch d {
	case 0:
		return i
	case -1:
		if i%8 == 0 {
			return invalidSquare
		}
		return i - 1
	case 1:
		if i%8 == 7 {
			return invalidSquare
		}
		return i + 1
	case -8:
		if i/8 == 0 {
			return invalidSquare
		}
		return i - 8
	case 8:
		if i/8 == 7 {
			return invalidSquare
		}
		return i + 8
	default:
		return invalidSquare
	}
}
