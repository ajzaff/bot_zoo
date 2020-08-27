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
	if d == 0 {
		return i
	}
	if !i.Valid() {
		return invalidSquare
	}
	if d < 0 {
		if i%8 == 0 || d < -64 || i < Square(-d) {
			return invalidSquare
		}
		v := Square(-d)
		if i < v {
			return invalidSquare
		}
		return i - v
	}
	if i%8 == 7 || d > 64 {
		return invalidSquare
	}
	v := i + Square(d)
	if !v.Valid() {
		return invalidSquare
	}
	return v
}
