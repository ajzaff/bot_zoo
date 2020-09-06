package zoo

var (
	stepsB       [64]Bitboard
	rabbitStepsB [2][64]Bitboard
)

func init() {
	for i := Square(0); i < 64; i++ {
		b := i.Bitboard()
		steps := b.Neighbors()
		stepsB[i] = steps
		grSteps := steps
		if b&NotRank1 != 0 { // rabbits can't move backwards.
			grSteps ^= b >> 8
		}
		srSteps := steps
		if b&NotRank8 != 0 {
			srSteps ^= b << 8
		}
		rabbitStepsB[Gold][i] = grSteps
		rabbitStepsB[Silver][i] = srSteps
	}
}

// func (p *Pos) getPushes(steps *[]Step) {
// 	s1, s2 := p.side, p.side.Opposite()
// 	s2n := p.presence[s2].Neighbors()
// 	for p1 := GCat.MakeColor(s1); p1 <= GElephant.MakeColor(s1); p1++ {
// 		p1b := p.bitboards[p1]
// 		if p1b == 0 || p1b&s2n == 0 {
// 			continue
// 		}
// 		p1n := p1b.Neighbors()
// 		for p2 := GRabbit.MakeColor(s2); p2 < p1.MakeColor(s2); p2++ {
// 			p2b := p.bitboards[p2]
// 			if p2b == 0 {
// 				continue
// 			}
// 			(p1n & p2b).Each(func(b Bitboard) {
// 				if p.frozenB(b) {
// 					return
// 				}
// 				src := b.Square()
// 				sB := stepsB
// 				if p2.SameType(GRabbit) {
// 					sB = rabbitStepsB[p2.Color()]
// 				} else {
// 					sB = stepsB
// 				}
// 				sB[src].Each(func(d Bitboard) {
// 					if p.bitboards[Empty]&d == 0 {
// 						return
// 					}
// 					dest := d.Square()
// 					assert("src == dest", src != dest)
// 					*steps = append(*steps, Step{
// 						Src:    src,
// 						Dest:   dest,
// 						Piece1: p2,
// 					})
// 				})
// 			})
// 		}
// 	}
// }

func unguardedB(b, presence Bitboard) Bitboard {
	return b & ^presence.Neighbors()
}

func trappedB(b, presence Bitboard) Bitboard {
	return b & unguardedB(Traps, presence)
}

// captureB determines if a move from srcB to destB would result in a capture
// without making the move. If so, the piece and square of the captured piece is returned.
func (p *Pos) capture(presence, srcB, destB Bitboard) Capture {
	newPresence := presence ^ (srcB | destB)
	if b := unguardedB(destB&Traps, newPresence); b != 0 {
		return Capture{
			Piece: p.atB(srcB),
			Src:   destB.Square(),
		}
	}
	if b := trappedB(newPresence&srcB.Neighbors(), newPresence); b != 0 {
		return Capture{
			Piece: p.atB(b),
			Src:   b.Square(),
		}
	}
	return Capture{}
}

// Steps generates steps including sliding moves, pushes and pulls.
// Captures are completed internally to Step, but you may check whether
// a step results in a capture by calling captures with the step as an
// argument.
func (p *Pos) Steps() []Step {
	if p.stepsLeft == 0 {
		return []Step{{Pass: true}}
	}
	var steps []Step
	c1, c2 := p.side, p.side.Opposite()
	canPush := p.stepsLeft > 1
	for t := GRabbit.MakeColor(c1); t <= GElephant.MakeColor(c1); t++ {
		ts := p.bitboards[t]
		if ts == 0 {
			continue
		}
		sB := stepsB

		// My rabbits can only step forward.
		if t.SameType(GRabbit) {
			sB = rabbitStepsB[p.side]
		}
		ts.Each(func(sb Bitboard) {
			if p.frozenB(sb) {
				return
			}
			src := sb.Square()

			// Find pullable pieces next to src (if canPush):
			var pullPieces []Piece
			var pullAlts []Bitboard
			if canPush && GRabbit.WeakerThan(t) {
				(sb.Neighbors() & p.presence[c2]).Each(func(nb Bitboard) {
					for r := GRabbit.MakeColor(c2); r < t.MakeColor(c2); r++ {
						if p.bitboards[r]&nb == 0 {
							continue
						}
						pullPieces = append(pullPieces, r)
						pullAlts = append(pullAlts, nb)
					}
				})
			}

			sB[src].Each(func(db Bitboard) {
				if p.bitboards[Empty]&db == 0 {
					if p.presence[c2]&db == 0 {
						return
					}
					// Generate all pushes out of this dest (if canPush):
					// Check that there's an empty place to push them to.
					if canPush {
						if r := p.atB(db); r.WeakerThan(t) {
							dbn := db.Neighbors()
							(dbn & p.bitboards[Empty]).Each(func(ab Bitboard) {
								step := Step{
									Src:    src,
									Dest:   db.Square(),
									Alt:    ab.Square(),
									Piece1: t,
									Piece2: r,
								}
								if cap := p.capture(p.presence[c2], db, ab); cap.Valid() {
									step.Cap = cap
								}
								steps = append(steps, step)
							})
						}
					}
					return
				}
				dest := db.Square()
				step := Step{
					Src:    src,
					Dest:   dest,
					Alt:    invalidSquare,
					Piece1: t,
				}
				if cap := p.capture(p.presence[c1], sb, db); cap.Valid() {
					step.Cap = cap
				}
				steps = append(steps, step)

				// Generate all pulls to this dest (if canPush):
				for i := range pullPieces {
					step.Alt = pullAlts[i].Square()
					step.Piece2 = pullPieces[i]
					if !step.Capture() {
						if cap := p.capture(p.presence[c2], pullAlts[i], sb); cap.Valid() {
							step.Cap = cap
						}
					}
					steps = append(steps, step)
				}
			})
		})
	}
	if p.stepsLeft < 4 {
		steps = append(steps, Step{Pass: true})
	}
	return steps
}

func (p *Pos) getRootMoves(prefix []Step, moves *[][]Step, stepsLeft int) {
	if stepsLeft == 0 {
		if !prefix[len(prefix)-1].Pass {
			prefix = append(prefix, Step{Pass: true})
		}
		*moves = append(*moves, prefix)
		return
	}
	assert("movesLeft < 0", stepsLeft > 0)
	for _, step := range p.Steps() {
		if step.Len() > stepsLeft {
			continue
		}
		newPrefix := make([]Step, 1+len(prefix))
		copy(newPrefix, prefix)
		newPrefix[len(newPrefix)-1] = step
		if step.Pass {
			*moves = append(*moves, newPrefix)
			continue
		}
		if err := p.Step(step); err != nil {
			panic(err)
		}
		p.getRootMoves(newPrefix, moves, stepsLeft-step.Len())
		if err := p.Unstep(); err != nil {
			panic(err)
		}
	}
}

func (e *Engine) getRootMovesLen(p *Pos, depth int) [][]Step {
	if depth <= 0 || depth > 4 {
		panic("depth <= 0 || depth > 4")
	}
	if e.p.stepsLeft-depth < 0 {
		panic("stepsLeft < depth")
	}
	var moves [][]Step
	for i := 1; i <= depth; i++ {
		p.getRootMoves(nil, &moves, i)
	}
	return moves
}
