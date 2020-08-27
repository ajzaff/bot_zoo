package zoo

type Color int

const (
	Gold Color = iota
	Silver
	Neither Color = -1
)

const (
	colorMask   = 8
	decolorMask = ^colorMask
)

func (c Color) PieceMask() Piece {
	if c == Gold {
		return 0
	}
	return colorMask
}

type Piece int

const PChars = " RCDHMExxrcdhme"

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

func (p Piece) MakeGold() Piece {
	return p & ^colorMask
}

func (p Piece) MakeSilver() Piece {
	return p | colorMask
}
