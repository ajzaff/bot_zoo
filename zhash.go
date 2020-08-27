package zoo

import (
	"fmt"
	"math/rand"
)

const zseed = 1337

var zkeys [1 + 5 + 15*64]int64

func newZKey(r *rand.Rand, usedKeys map[int64]bool) int64 {
	candidate := int64(0)
	for usedKeys[candidate] {
		candidate = r.Int63()
	}
	usedKeys[candidate] = true
	return candidate
}

func init() {
	r := rand.New(rand.NewSource(zseed))
	usedKeys := map[int64]bool{0: true}
	for i := range zkeys {
		zkeys[i] = newZKey(r, usedKeys)
	}
}

func ZSilverKey() int64 {
	return zkeys[0]
}

func ZStepsKey(steps int) int64 {
	if steps < 0 || steps > 4 {
		panic(fmt.Sprintf("invalid steps: %d", steps))
	}
	return zkeys[1+steps]
}

func ZPieceKey(p Piece, i Square) int64 {
	return zkeys[1+5+int(p)*64+int(i)]
}

func ZHash(bitboards []Bitboard, side Color, steps int) int64 {
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
			zhash ^= zkeys[1+5+b.Square()*64]
		}
	}
	return zhash
}
