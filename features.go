package zoo

import expb "github.com/ajzaff/bot_zoo/proto"

func featureBitset(ex *expb.Example, c Color, p Piece) *expb.Example_Bitset {
	idx := uint32(p.RemoveColor())
	if c != p.Color() {
		idx += 6
	}
	b := ex.Bitsets[idx]
	if b == nil {
		b = &expb.Example_Bitset{}
		ex.Bitsets[idx] = b
	}
	return b
}

func featureIndex(c Color, i Square) uint32 {
	if c == Silver {
		i = i.Flip()
	}
	return uint32(i)
}

// precondition: p != Empty
func pushFeatureIndex(c Color, i Square, p Piece) int {
	idx := 768 + 64*int(p.RemoveColor()-1)
	if c == Silver {
		i = i.Flip()
	}
	return idx + int(i)
}

func clearFeatures(ex *expb.Example) {
	ex.Bitsets = make(map[uint32]*expb.Example_Bitset)
}

// Features fills in the flat features slice which has the shape (8 * 8 * 21,)
// with features extracted from p. This slice has all the necessary components
// for encoding an Arimaa position and can be reshaped to directly input into
// the network.
// Input planes are as follows:
//	My side pieces  (6 planes),
//	Opponent pieces (6 planes),
//	Push piece mask (6 planes),
//	In push?        (1 plane; all 0 or 1),
//	Last turn?      (1 plane; all 0 or 1),
//	Setup?          (1 plane; all 0 or 1).
// Positions are flipped as necessary to ensure the side to move is relative to
// Gold's perspective of the board (with home rank of A, and goal rank of H).
// lateralFeats is filled with board features mirrored laterally (for dataset
// augmentation).
func Features(p *Pos, ex *expb.Example) {
	c := p.Side()

	clearFeatures(ex)
	for i := A1; i <= H8; i++ {
		p := p.At(i)
		if p != Empty {
			b := featureBitset(ex, c, p)
			b.Ones = append(b.Ones, featureIndex(c, i))
		}
	}

	if src, piece, ok := p.Push(); src.Valid() {
		idx := 12 + uint32(piece.RemoveColor())
		b := ex.Bitsets[idx]
		if b == nil {
			b = &expb.Example_Bitset{}
			ex.Bitsets[idx] = b
		}
		b.Ones = append(b.Ones, uint32(src))
		if ok {
			ex.Bitsets[18] = &expb.Example_Bitset{AllOnes: true}
		}
	}

	if p.LastStep() {
		ex.Bitsets[19] = &expb.Example_Bitset{AllOnes: true}
	}

	if p.MoveNum() == 1 {
		ex.Bitsets[20] = &expb.Example_Bitset{AllOnes: true}
	}
}
