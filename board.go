package zoo

type Bitboard uint64

const (
	AllBits  Bitboard = 0xFFFFFFFFFFFFFFFF
	NotFileA Bitboard = 0xFEFEFEFEFEFEFEFE
	NotFileH Bitboard = 0x7F7F7F7F7F7F7F7F
	NotRank1 Bitboard = 0xFFFFFFFFFFFFFF00
	NotRank8 Bitboard = 0x00FFFFFFFFFFFFFF
)

const (
	TrapC3          = Bitboard(1) << 18
	TrapF3          = Bitboard(1) << 21
	TrapC6          = Bitboard(1) << 42
	TrapF6          = Bitboard(1) << 45
	Traps  Bitboard = 0x0000240000240000
)

func (b Bitboard) Neighbors() Bitboard {
	bb := (b & NotFileA) >> 1
	bb |= (b & NotFileH) << 1
	bb |= (b & NotRank1) >> 8
	bb |= (b & NotRank8) << 8
	return bb
}

func (b Bitboard) ne0v() uint8 {
	if b == 0 {
		return 0
	}
	return 1
}

func (b Bitboard) Square() Square {
	x := (b & 0xAAAAAAAAAAAAAAAA).ne0v()
	x |= (b & 0xCCCCCCCCCCCCCCCC).ne0v() << 1
	x |= (b & 0xF0F0F0F0F0F0F0F0).ne0v() << 2
	x |= (b & 0xFF00FF00FF00FF00).ne0v() << 3
	x |= (b & 0xFFFF0000FFFF0000).ne0v() << 4
	x |= (b & 0xFFFFFFFF00000000).ne0v() << 5
	return Square(x)
}
