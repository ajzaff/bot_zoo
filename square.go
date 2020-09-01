package zoo

import "strconv"

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

func (src Square) Delta(dest Square) int8 {
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

func (i Square) String() string {
	if i.Valid() {
		return string([]byte{
			files[i%8],
			ranks[i/8],
		})
	}
	return strconv.Itoa(int(i))
}

const invalidSquare = 255

func NewDelta(d int8) string {
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
		return ""
	}
}

func ParseDelta(s string) int8 {
	switch s {
	case "n":
		return +8
	case "s":
		return -8
	case "e":
		return 1
	case "w":
		return -1
	default:
		return 0
	}
}

func (i Square) Translate(d int8) Square {
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
