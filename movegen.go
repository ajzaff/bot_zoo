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
func (p *Pos) capture(presence, srcB, destB Bitboard, src, dest Square) Capture {
	newPresence := presence ^ (srcB | destB)
	if b := unguardedB(destB&Traps, newPresence); b != 0 {
		// Capture of piece pushed from srcB to destB.
		return Capture{
			Piece: p.board[src],
			Src:   dest,
		}
	}
	if b := trappedB(newPresence&srcB.Neighbors(), newPresence); b != 0 {
		// Capture of piece left next to src.
		// TODO(ajzaff): It would be more clear to have a adjacentTrap helper for this.
		return Capture{
			Piece: p.board[b.Square()],
			Src:   b.Square(),
		}
	}
	return Capture{}
}

// Steps is a low level helper for getting steps in a current position
// without allocating a new array backed slice every time.
// This (and scoreSteps) was caught in profiling as the biggest cost to search.
func (p *Pos) Steps(a *[]Step) {
	if p.stepsLeft == 0 {
		*a = append(*a, Step{Pass: true})
		return
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
			src := sb.Square()

			if p.frozen(c1, src) {
				return
			}

			// Find pullable pieces next to src (if canPush):
			var pullPieces []Piece
			var pullAlts []Bitboard
			if canPush && GRabbit.WeakerThan(t) {
				(sb.Neighbors() & p.presence[c2]).Each(func(nb Bitboard) {
					r := p.board[nb.Square()]
					if r.Color() == c2 && r.WeakerThan(t) {
						pullPieces = append(pullPieces, r)
						pullAlts = append(pullAlts, nb)
					}
				})
			}

			sB[src].Each(func(db Bitboard) {
				dest := db.Square()

				if p.bitboards[Empty]&db == 0 {
					if p.presence[c2]&db == 0 {
						return
					}

					// Generate all pushes out of this dest (if canPush):
					// Check that there's an empty place to push them to.
					if canPush {
						r := p.board[dest]
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
								if cap := p.capture(p.presence[c1], sb, db, src, dest); cap.Valid() {
									step.Cap = cap
								} else if cap := p.capture(p.presence[c2], db, ab, dest, ab.Square()); cap.Valid() {
									step.Cap = cap
								}
								*a = append(*a, step)
							})
						}
					}
					return
				}
				step := Step{
					Src:    src,
					Dest:   dest,
					Alt:    invalidSquare,
					Piece1: t,
				}
				if cap := p.capture(p.presence[c1], sb, db, src, dest); cap.Valid() {
					step.Cap = cap
				}
				*a = append(*a, step)

				// Generate all pulls to this dest (if canPush):
				for i := range pullPieces {
					step.Alt = pullAlts[i].Square()
					step.Piece2 = pullPieces[i]
					if !step.Capture() {
						if cap := p.capture(p.presence[c2], pullAlts[i], sb, step.Alt, src); cap.Valid() {
							step.Cap = cap
						}
					}
					*a = append(*a, step)
				}
			})
		})
	}
	if p.stepsLeft < 4 {
		*a = append(*a, Step{Pass: true})
	}
}

func (p *Pos) getRootMovesLenInternal(set map[int64]bool, prefix []Step, moves *[][]Step, stepsLeft int) {
	if stepsLeft == 0 {
		move := make([]Step, len(prefix), len(prefix)+1)
		copy(move, prefix)
		if !prefix[len(prefix)-1].Pass {
			move = append(move, Step{Pass: true})
		}
		if !set[p.zhash] {
			set[p.zhash] = true
			*moves = append(*moves, move)
		}
		return
	}
	assert("movesLeft < 0", stepsLeft > 0)
	n := len(prefix)
	var stepList StepList
	stepList.Generate(p)
	for i := 0; i < stepList.Len(); i++ {
		step := stepList.At(i)
		if step.Len() > stepsLeft {
			continue
		}
		prefix = append(prefix, step)
		if err := p.Step(step); err != nil {
			panic(err)
		}
		if step.Pass {
			if !set[p.zhash] {
				set[p.zhash] = true
				move := make([]Step, len(prefix))
				copy(move, prefix)
				*moves = append(*moves, move)
			}
		} else {
			p.getRootMovesLenInternal(set, prefix, moves, stepsLeft-step.Len())
		}
		if err := p.Unstep(); err != nil {
			panic(err)
		}
		prefix = prefix[:n]
	}
}

func (e *Engine) getRootMovesLen(p *Pos, depth int) [][]Step {
	if depth <= 0 || depth > 4 {
		panic("depth <= 0 || depth > 4")
	}
	if p.stepsLeft-depth < 0 {
		panic("stepsLeft < depth")
	}
	var moves [][]Step
	var prefix []Step
	set := map[int64]bool{p.zhash: true}
	for i := 1; i <= depth; i++ {
		p.getRootMovesLenInternal(set, prefix, &moves, i)
	}
	return moves
}
