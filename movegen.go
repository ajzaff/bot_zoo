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
		var piece Piece
		for t := GRabbit; t <= SElephant; t++ {
			if p.bitboards[t]&srcB != 0 {
				piece = t
				break
			}
		}
		// Capture of piece pushed from srcB to destB.
		return Capture{
			Piece: piece,
			Src:   destB.Square(),
		}
	}
	if b := trappedB(newPresence&srcB.Neighbors(), newPresence); b != 0 {
		var piece Piece
		for t := GRabbit; t <= SElephant; t++ {
			if p.bitboards[t]&b != 0 {
				piece = t
				break
			}
		}
		// Capture of piece left next to src.
		// TODO(ajzaff): It would be more clear to have a adjacentTrap helper for this.
		return Capture{
			Piece: piece,
			Src:   b.Square(),
		}
	}
	return Capture{}
}

// Steps is a low level helper for getting steps in a current position
// without allocating a new array backed slice every time.
// This (and scoreSteps) was caught in profiling as the biggest cost to search.
func (p *Pos) Steps() (a []Step) {
	if p.stepsLeft == 0 {
		return []Step{{Pass: true}}
	}
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
			if p.frozenB(c1, sb) {
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
						for r := GRabbit.MakeColor(c2); r <= GElephant.MakeColor(c2); r++ {
							if p.bitboards[r]&db != 0 {
								if r.WeakerThan(t) {
									dbn := db.Neighbors()
									(dbn & p.bitboards[Empty]).Each(func(ab Bitboard) {
										step := Step{
											Src:    src,
											Dest:   db.Square(),
											Alt:    ab.Square(),
											Piece1: t,
											Piece2: r,
										}
										if cap := p.capture(p.presence[c1], sb, db); cap.Valid() {
											step.Cap = cap
										} else if cap := p.capture(p.presence[c2], db, ab); cap.Valid() {
											step.Cap = cap
										}
										a = append(a, step)
									})
								}
							}
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
				a = append(a, step)

				// Generate all pulls to this dest (if canPush):
				for i := range pullPieces {
					step.Alt = pullAlts[i].Square()
					step.Piece2 = pullPieces[i]
					if !step.Capture() {
						if cap := p.capture(p.presence[c2], pullAlts[i], sb); cap.Valid() {
							step.Cap = cap
						}
					}
					a = append(a, step)
				}
			})
		})
	}
	if p.stepsLeft < 4 {
		a = append(a, Step{Pass: true})
	}
	return a
}

func (p *Pos) getRootMovesLenInternal(set map[int64]bool, prefix []Step, moves *[][]Step, stepsLeft int) {
	if stepsLeft == 0 {
		if !prefix[len(prefix)-1].Pass {
			prefix = append(prefix, Step{Pass: true})
		}
		if !set[p.zhash] {
			set[p.zhash] = true
			*moves = append(*moves, prefix)
		}
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
		if err := p.Step(step); err != nil {
			panic(err)
		}
		if step.Pass {
			if !set[p.zhash] {
				set[p.zhash] = true
				*moves = append(*moves, newPrefix)
			}
		} else {
			p.getRootMovesLenInternal(set, newPrefix, moves, stepsLeft-step.Len())
		}
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
	set := map[int64]bool{p.zhash: true}
	for i := 1; i <= depth; i++ {
		p.getRootMovesLenInternal(set, nil, &moves, i)
	}
	return moves
}
