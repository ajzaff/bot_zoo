package zoo

import (
	"errors"
	"fmt"
)

type Pos struct {
	bitboards []Bitboard
	presence  []Bitboard
	side      Color
	moveNum   int
	// TODO(ajzaff): Add numSteps field for efficiency.
	steps []Step
	zhash int64
}

func newPos(
	bitboards []Bitboard,
	presence []Bitboard,
	side Color,
	moveNum int,
	steps []Step,
	zhash int64,
) *Pos {
	if bitboards == nil {
		bitboards = make([]Bitboard, 15)
		bitboards[Empty] = AllBits
	}
	if presence == nil {
		presence = make([]Bitboard, 2)
		for _, p := range []Piece{
			GRabbit,
			GCat,
			GDog,
			GHorse,
			GCamel,
			GElephant,
		} {
			presence[Gold] |= bitboards[p]
			presence[Silver] |= bitboards[p.MakeColor(Silver)]
		}
	}
	if zhash == 0 {
		zhash = ZHash(bitboards, side, len(steps))
	}
	return &Pos{
		bitboards: bitboards,
		presence:  presence,
		side:      side,
		moveNum:   moveNum,
		steps:     steps,
		zhash:     zhash,
	}
}

func (p *Pos) Terminal() bool {
	return p.terminalGoalValue() != 0 || p.terminalEliminationValue() != 0 || p.terminalImmobilizedValue() != 0
}

func (p *Pos) At(i Square) Piece {
	return p.atB(i.Bitboard())
}

func (p *Pos) atB(b Bitboard) Piece {
	for piece := GRabbit; piece <= SElephant && int(piece) < len(p.bitboards); piece++ {
		if p.bitboards[piece]&b != 0 {
			return piece
		}
	}
	return Empty
}

func (p *Pos) Place(piece Piece, i Square) error {
	if piece == Empty {
		return p.Remove(i)
	}
	assert("invalid piece", piece.Valid())
	b := i.Bitboard()
	if p.bitboards[Empty]&b == 0 {
		return fmt.Errorf("piece already present")
	}
	p.bitboards[piece] |= b
	p.bitboards[Empty] &= ^b
	p.presence[piece.Color()] |= b
	p.zhash ^= ZPieceKey(piece, i)
	return nil
}

func (p *Pos) Remove(i Square) error {
	b := i.Bitboard()
	if p.bitboards[Empty]&b != 0 {
		return fmt.Errorf("no piece present on %s", i)
	}
	piece := p.atB(b)
	assert("inconsistent board state", piece.Valid())
	p.bitboards[piece] &= ^b
	p.bitboards[Empty] |= b
	p.presence[piece.Color()] &= ^b
	p.zhash ^= ZPieceKey(piece, i)
	return nil
}

func (p *Pos) Step(step Step) error {
	switch step.Kind() {
	case KindSetup:
		if step.Capture() {
			return fmt.Errorf("setup step has capture")
		}
		if p.moveNum > 1 {
			return fmt.Errorf("setup move after first turn")
		}
		return p.Place(step.Piece1, step.Alt)
	case KindPush:
		if err := p.Remove(step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece2, step.Alt); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Remove(step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece1, step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
	case KindPull:
		if err := p.Remove(step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece1, step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Remove(step.Alt); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece2, step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
	case KindDefault:
		if step.Src.Valid() && step.Dest.Valid() { // Not lone capture:
			if err := p.Remove(step.Src); err != nil {
				return fmt.Errorf("%s: %v", step, err)
			}
			if err := p.Place(step.Piece1, step.Dest); err != nil {
				return fmt.Errorf("%s: %v", step, err)
			}
		}
	case KindInvalid:
		return fmt.Errorf("invalid step: %s", step)
	}
	p.steps = append(p.steps, step)
	if step.Capture() {
		if err := p.Remove(step.Cap.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
	}
	return nil
}

func (p *Pos) Unstep(step Step) error {
	if step.Capture() {
		if err := p.Place(step.Cap.Piece, step.Cap.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
	}
	switch step.Kind() {
	case KindSetup:
		if step.Capture() {
			return fmt.Errorf("setup step has capture")
		}
		if p.moveNum > 1 {
			return fmt.Errorf("setup move after first turn")
		}
		return p.Remove(step.Alt)
	case KindPush:
		if err := p.Remove(step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece1, step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Remove(step.Alt); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece2, step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
	case KindPull:
		if err := p.Remove(step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece2, step.Alt); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Remove(step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece1, step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
	case KindDefault:
		if err := p.Remove(step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece1, step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
	case KindInvalid:
		return fmt.Errorf("invalid step: %s", step)
	}
	p.steps = p.steps[:len(p.steps)-1]
	return nil
}

func (p *Pos) NullMove() {
	p.steps = nil
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum++
	}
}

func (p *Pos) NullUnmove(steps []Step) {
	p.steps = steps
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Silver {
		p.moveNum--
	}
}

var errRecurringPosition = errors.New("recurring position")

func (p *Pos) Move(steps []Step) error {
	if p.moveNum == 1 && len(steps) != 16 {
		return fmt.Errorf("move %s: wrong number of setup moves", MoveString(steps))
	}
	initZHash := p.zhash
	for _, step := range steps {
		if err := p.Step(step); err != nil {
			return fmt.Errorf("move %s: %v", MoveString(steps), err)
		}
	}
	// TODO(ajzaff): Movegen should filter moves that would result
	// in recurring positions.
	if initZHash == p.zhash {
		return errRecurringPosition
	}
	p.NullMove()
	return nil
}

func (p *Pos) Unmove(steps []Step) error {
	if p.moveNum == 1 && len(steps) != 16 {
		return fmt.Errorf("wrong number of setup moves")
	}
	p.NullUnmove(steps)
	for i := len(steps) - 1; i >= 0; i-- {
		if err := p.Unstep(steps[i]); err != nil {
			return fmt.Errorf("umove %s: %v", MoveString(steps), err)
		}
	}
	return nil
}
