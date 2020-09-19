package zoo

import (
	"fmt"
	"math/rand"
)

// Hash implements a Zobrist hash on Arimaa positions.
type Hash uint64

var hashKeys [1 + 17 + 15*64]Hash

func newHashKey(r *rand.Rand, usedKeys map[Hash]bool) Hash {
	var candidate Hash
	for usedKeys[candidate] {
		candidate = Hash(r.Uint64())
	}
	usedKeys[candidate] = true
	return candidate
}

const hashSeed = 1337

func init() {
	r := rand.New(rand.NewSource(hashSeed))
	usedKeys := map[Hash]bool{0: true}
	for i := range hashKeys {
		hashKeys[i] = newHashKey(r, usedKeys)
	}
}

func silverHashKey() Hash {
	return hashKeys[0]
}

func stepsHashKey(steps int) Hash {
	if steps < 0 || steps > 16 {
		panic(fmt.Sprintf("invalid steps: %d", steps))
	}
	return hashKeys[1+steps]
}

func pieceHashKey(p Piece, i Square) Hash {
	return hashKeys[1+17+int(p)*64+int(i)]
}

func computeHash(bitboards []Bitboard, side Color, steps int) Hash {
	zhash := stepsHashKey(steps)
	if side != Gold {
		zhash ^= silverHashKey()
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
			zhash ^= pieceHashKey(p, b.Square())
		}
	}
	return zhash
}
