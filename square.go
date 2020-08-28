package zoo

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

func (i Square) String() string {
	return string([]byte{
		files[i%8],
		ranks[i/8],
	})
}

const invalidSquare = 255

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
