package zoo

import (
	"errors"
	"fmt"
)

type Pos struct {
	bitboards []Bitboard // bitboard data
	presence  []Bitboard // board presence for each side
	side      Color      // side to play
	moveNum   int        // number of moves left for this turn
	moves     [][]Step   // moves to arrive at this position after appending steps
	steps     []Step     // steps of the current move
	stepsLeft int        // steps remaining in the current move
	zhash     int64      // zhash of the current position
}

func newPos(
	bitboards []Bitboard,
	presence []Bitboard,
	side Color,
	moveNum int,
	moves [][]Step,
	steps []Step,
	stepsLeft int,
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
		stepsLeft: stepsLeft,
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
		return fmt.Errorf("piece already present on %s", i)
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

func (p *Pos) Pass() {
	if len(p.steps) > 0 {
		p.moves = append(p.moves, p.steps)
		p.steps = nil
	}
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum++
	}
	p.stepsLeft = 4
	if p.moveNum == 1 {
		p.stepsLeft = 16
	}
}

func (p *Pos) Unpass() {
	assert("len(moves) == 0", len(p.moves) > 0)
	assert("len(steps) != 0", len(p.steps) == 0)

	p.steps = p.moves[len(p.moves)-1]
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum--
	}
	p.stepsLeft = 4
	if p.moveNum == 1 {
		p.stepsLeft = 16
	}
}

func (p *Pos) Step(step Step) error {
	if step.Pass {
		p.Pass()
		return nil
	}
	{
		n := step.Len()
		if n > p.stepsLeft {
			return fmt.Errorf("%s: not enough steps left", step)
		}
		p.stepsLeft -= n
	}
	switch step.Kind() {
	case KindSetup:
		if step.Capture() {
			return fmt.Errorf("%s: setup step has capture", step)
		}
		if p.moveNum > 1 {
			return fmt.Errorf("%s: setup move after first turn", step)
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
		if err := p.Remove(step.Src); err != nil {
			return fmt.Errorf("%s: %v", step, err)
		}
		if err := p.Place(step.Piece1, step.Dest); err != nil {
			return fmt.Errorf("%s: %v", step, err)
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

func (p *Pos) Unstep() error {
	var step Step
	if len(p.steps) == 0 {
		assert("len(moves) == 0", len(p.moves) > 0)
		p.steps = p.moves[len(p.moves)-1]
		step = p.steps[len(p.steps)-1]
		p.moves = p.moves[:len(p.moves)-1]
	} else {
		step = p.steps[len(p.steps)-1]
		p.steps = p.steps[:len(p.steps)-1]
	}
	p.stepsLeft += step.Len()
	if step.Pass {
		p.Unpass()
		return nil
	}
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
	return nil
}

var errRecurringPosition = errors.New("recurring position")

func (p *Pos) Move(steps []Step) error {
	if p.moveNum == 1 && len(steps) != 16 {
		return fmt.Errorf("move %s: wrong number of setup moves", MoveString(steps))
	}
	initSide := p.side
	initZHash := p.zhash
	for i, step := range steps {
		if step.Pass && i < len(steps)-1 {
			return fmt.Errorf("move %s: pass before last step", MoveString(steps))
		}
		if err := p.Step(step); err != nil {
			return fmt.Errorf("move %s: %v", MoveString(steps), err)
		}
	}
	if p.side == initSide {
		p.Pass()
	}
	// TODO(ajzaff): Movegen should filter moves that would result
	// in recurring positions.
	if initZHash == p.zhash^ZSilverKey() {
		return errRecurringPosition
	}
	return nil
}

func (p *Pos) Unmove() error {
	assert("len(moves) == 0", len(p.moves) > 0)

	n := len(p.moves) - 1
	move := p.moves[n]
	if p.moveNum == 1 && len(move) != 16 {
		return fmt.Errorf("unmove: %s: wrong number of setup moves", MoveString(move))
	}
	for i := n; i >= 0; i-- {
		step := move[i]
		if step.Pass && i < n-1 {
			return fmt.Errorf("unmove %s: pass before last step", MoveString(move))
		}
		if err := p.Unstep(); err != nil {
			return fmt.Errorf("unmove %s: %v", MoveString(move), err)
		}
	}
	return nil
}
