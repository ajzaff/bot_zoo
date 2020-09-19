package zoo

import (
	"fmt"
	"math/rand"
)

// Hash implements a Zobrist hash on Arimaa positions.
type Hash uint64

var zkeys [1 + 5 + 15*64]uint64

func newZKey(r *rand.Rand, usedKeys map[uint64]bool) uint64 {
	candidate := uint64(0)
	for usedKeys[candidate] {
		candidate = r.Uint64()
	}
	usedKeys[candidate] = true
	return candidate
}

const zseed = 1337

func init() {
	r := rand.New(rand.NewSource(zseed))
	usedKeys := map[uint64]bool{0: true}
	for i := range zkeys {
		zkeys[i] = newZKey(r, usedKeys)
	}
}

func ZSilverKey() uint64 {
	return zkeys[0]
}

func ZStepsKey(steps int) uint64 {
	if steps < 0 || steps > 4 {
		panic(fmt.Sprintf("invalid steps: %d", steps))
	}
	return zkeys[1+steps]
}

func ZPieceKey(p Piece, i Square) uint64 {
	return zkeys[1+5+int(p)*64+int(i)]
}

func ZHash(bitboards []Bitboard, side Color, steps int) uint64 {
	zhash := ZStepsKey(steps)
	if side != Gold {
		zhash ^= ZSilverKey()
	}
	for _, p := range []Piece{
		GRabbit,
		GCat,
		GDog,
		GHorse,
		GCamel,
		GElephant,
		SRabbit,
		SCat,
		SDog,
		SHorse,
		SCamel,
		SElephant,
	} {
		pieces := bitboards[p]
		for pieces > 0 {
			b := pieces & -pieces
			pieces ^= b
			zhash ^= ZPieceKey(p, b.Square())
		}
	}
	return zhash
}
