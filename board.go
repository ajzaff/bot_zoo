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
	TrapC3        = Bitboard(1) << 18
	TrapF3        = Bitboard(1) << 21
	TrapC6        = Bitboard(1) << 42
	TrapF6        = Bitboard(1) << 45
	Traps         = Bitboard(0x0000240000240000)
	TrapNeighbors = Traps>>1 | Traps<<1 | Traps>>8 | Traps<<8
)

func (b Bitboard) Neighbors() Bitboard {
	bb := (b & NotFileA) >> 1
	bb |= (b & NotFileH) << 1
	bb |= (b & NotRank1) >> 8
	bb |= (b & NotRank8) << 8
	return bb
}

const debruijn64 Bitboard = 0x03f79d71b4cb0a89

var bsfIndex64 = [64]Square{
	0, 1, 48, 2, 57, 49, 28, 3,
	61, 58, 50, 42, 38, 29, 17, 4,
	62, 55, 59, 36, 53, 51, 43, 22,
	45, 39, 33, 30, 24, 18, 12, 5,
	63, 47, 56, 27, 60, 41, 37, 16,
	54, 35, 52, 21, 44, 32, 23, 11,
	46, 26, 40, 15, 34, 20, 31, 10,
	25, 14, 19, 9, 13, 8, 7, 6,
}

func (b Bitboard) Square() Square {
	if b == 0 {
		return invalidSquare
	}
	return bsfIndex64[((b&-b)*debruijn64)>>58]
}

func (b Bitboard) Each(f func(e Bitboard)) {
	for b > 0 {
		e := b & -b
		b &= ^e
		f(e)
	}
}

var countTable [256]int

func init() {
	for i := 1; i < 256; i++ {
		countTable[i] = countTable[i/2] + (i & 1)
	}
}

func (b Bitboard) Count() int {
	return countTable[b&0xff] +
		countTable[(b>>8)&0xff] +
		countTable[(b>>16)&0xff] +
		countTable[(b>>24)&0xff] +
		countTable[(b>>32)&0xff] +
		countTable[(b>>40)&0xff] +
		countTable[(b>>48)&0xff] +
		countTable[b>>56]
}
