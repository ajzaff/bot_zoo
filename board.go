package zoo

// Bitboard represents
type Bitboard uint64

// Useful masks.
const (
	AllBits  Bitboard = 0xFFFFFFFFFFFFFFFF
	NotFileA Bitboard = 0xFEFEFEFEFEFEFEFE
	NotFileH Bitboard = 0x7F7F7F7F7F7F7F7F
	NotRank1 Bitboard = 0xFFFFFFFFFFFFFF00
	NotRank8 Bitboard = 0x00FFFFFFFFFFFFFF
)

// Trap masks.
const (
	Traps         Bitboard = 0x0000240000240000
	TrapNeighbors          = Traps>>1 | Traps<<1 | Traps>>8 | Traps<<8
)

// Neighbors computes the neighbors of the set bits in b.
func (b Bitboard) Neighbors() Bitboard {
	return (b&NotFileA)>>1 |
		(b&NotFileH)<<1 |
		(b&NotRank1)>>8 |
		(b&NotRank8)<<8
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

// Square returns the LSB Square in b.
// It is an error to pass 0 for b.
func (b Bitboard) Square() Square {
	return bsfIndex64[((b&-b)*debruijn64)>>58]
}

var countTable [256]uint8

func init() {
	for i := uint8(1); i < 255; i++ {
		countTable[i] = countTable[i/2] + (i & 1)
	}
	countTable[255] = countTable[127] + 1
}

// Count returns a count of the number of set bits in b.
func (b Bitboard) Count() uint8 {
	return countTable[b&0xff] +
		countTable[(b>>8)&0xff] +
		countTable[(b>>16)&0xff] +
		countTable[(b>>24)&0xff] +
		countTable[(b>>32)&0xff] +
		countTable[(b>>40)&0xff] +
		countTable[(b>>48)&0xff] +
		countTable[b>>56]
}
