package zoo

// precondition: p != Empty
func featureIndex(c Color, i Square, p Piece, lateral bool) int {
	idx := 64 * int(p.RemoveColor()-1)
	if c != p.Color() {
		idx += 384
	}
	if c == Silver {
		i = i.Flip()
	}
	if lateral {
		i = i.MirrorLateral()
	}
	return idx + int(i)
}

// precondition: p != Empty
func pushFeatureIndex(c Color, i Square, p Piece, lateral bool) int {
	idx := 768 + 64*int(p.RemoveColor()-1)
	if c == Silver {
		i = i.Flip()
	}
	if lateral {
		i = i.MirrorLateral()
	}
	return idx + int(i)
}

func clearFeatures(feats, lateralFeats []float32) {
	for i := range feats {
		feats[i] = 0
		lateralFeats[i] = 0
	}
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
func Features(p *Pos, feats, lateralFeats []float32) {
	c := p.Side()

	clearFeatures(feats, lateralFeats)
	for i := A1; i <= H8; i++ {
		p := p.At(i)
		if p != Empty {
			feats[featureIndex(c, i, p, false)] = 1
			lateralFeats[featureIndex(c, i, p, true)] = 1
		}
	}

	if src, piece, ok := p.Push(); src.Valid() {
		feats[pushFeatureIndex(c, src, piece, false)] = 1
		lateralFeats[pushFeatureIndex(c, src, piece, true)] = 1
		if ok {
			for i := 1152; i < 1216; i++ {
				feats[i] = 1
				lateralFeats[i] = 1
			}
		}
	}

	if p.LastStep() {
		for i := 1216; i < 1280; i++ {
			feats[i] = 1
			lateralFeats[i] = 1
		}
	}

	if p.MoveNum() == 1 {
		for i := 1280; i < 1344; i++ {
			feats[i] = 1
			lateralFeats[i] = 1
		}
	}
}
