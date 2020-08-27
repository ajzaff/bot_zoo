package zoo

import (
	"fmt"
	"math/rand"
)

const zseed = 0xF00F

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
	zkeys[0] = newZKey(r, usedKeys) // side to move is silver
	for i := 0; i < 5; i++ {        // number of moves remaining is 0-4.
		zkeys[1+i] = newZKey(r, usedKeys)
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
	} { // every piece X every square
		for j := 0; j < 64; j++ {
			zkeys[1+5+p*64] = newZKey(r, usedKeys)
		}
	}
}

func ZHash(bitboards []Bitboard, side Color, steps int) int64 {
	if steps < 0 || steps > 4 {
		panic(fmt.Sprintf("invalid steps: %d", steps))
	}
	zhash := zkeys[1+steps]
	if side != Gold {
		zhash ^= zkeys[0]
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
			zhash ^= zkeys[1+5+b.Index()*64]
		}
	}
	return zhash
}
