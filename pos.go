package zoo

import (
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

func (p *Pos) immobilized(c Color) bool {
	b := p.presence[c]
	for b > 0 {
		atB := b & -b
		if !p.frozenB(atB) {
			return false
		}
		b &= ^atB
	}
	return true
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
	b := i.Bitboard()
	if p.bitboards[Empty]&b == 0 {
		return fmt.Errorf("piece already present")
	}
	bs := make([]Bitboard, len(p.bitboards))
	for i := range p.bitboards {
		bs[i] = p.bitboards[i]
	}
	p.bitboards[piece] |= b
	p.bitboards[Empty] &= ^b
	p.presence[piece.Color()] |= b
	p.zhash ^= ZPieceKey(piece, i)
	return nil
}

func (p *Pos) Remove(i Square) error {
	b := i.Bitboard()
	if p.bitboards[Empty]&b == 0 {
		return fmt.Errorf("no piece present")
	}
	piece := GRabbit
	for piece <= SElephant && (p.bitboards[piece] == 0 || p.bitboards[piece]&b == 0) {
		piece++
	}
	assert("inconsistent board state", piece <= SElephant)
	bs := make([]Bitboard, len(p.bitboards))
	for i := range p.bitboards {
		bs[i] = p.bitboards[i]
	}
	p.bitboards[piece] &= ^b
	p.bitboards[Empty] |= b
	p.presence[piece.Color()] &= ^b
	p.zhash ^= ZPieceKey(piece, i)
	return nil
}

func (p *Pos) frozenB(b Bitboard) bool {
	neighbors := b.Neighbors()
	piece := p.atB(b)
	color := piece.Color()
	if neighbors&p.presence[color] == 0 &&
		neighbors&p.presence[color.Opposite()] != 0 {
		for s := piece.MakeColor(color.Opposite()) + 1; s <= GElephant.MakeColor(color.Opposite()); s++ {
			if neighbors&p.bitboards[s] != 0 {
				return true
			}
		}
	}
	return false
}

func (p *Pos) frozenNeighbors(b Bitboard) Bitboard {
	var res Bitboard
	b = b.Neighbors()
	for b > 0 {
		lsb := b & -b
		b &= ^lsb
		if p.frozenB(lsb) {
			res |= lsb
		}
	}
	return res
}

func (p *Pos) Step(step Step) error {
	if step.Setup() {
		return p.Place(step.Piece, step.Dest)
	}
	if step.Capture() {
		return p.Remove(step.Src)
	}
	if err := p.Remove(step.Src); err != nil {
		return err
	}
	piece := step.Piece
	if err := p.Place(piece, step.Dest); err != nil {
		return err
	}
	stepB := step.Src.Bitboard() | step.Dest.Bitboard()
	p.bitboards[piece] ^= stepB
	p.bitboards[Empty] ^= stepB
	p.presence[piece.Color()] ^= stepB
	p.zhash ^= ZPieceKey(piece, step.Src)
	p.zhash ^= ZPieceKey(piece, step.Dest)
	if c := p.capture(step); c.Capture() {
		return p.Step(c)
	}
	return nil
}

func (p *Pos) Unstep(step Step) error {
	if step.Setup() {
		return p.Remove(step.Dest)
	}
	piece := step.Piece
	if step.Capture() {
		return p.Place(piece, step.Src)
	}
	if c := p.capture(step); c.Capture() {
		if err := p.Unstep(c); err != nil {
			return err
		}
	}
	if err := p.Remove(step.Dest); err != nil {
		return err
	}
	if err := p.Place(piece, step.Src); err != nil {
		return err
	}
	stepB := step.Src.Bitboard() | step.Dest.Bitboard()
	p.bitboards[piece] ^= stepB
	p.bitboards[Empty] ^= stepB
	p.presence[piece.Color()] ^= stepB
	p.zhash ^= ZPieceKey(piece, step.Src)
	p.zhash ^= ZPieceKey(piece, step.Dest)
	return nil
}

func (p *Pos) NullMove() {
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum++
	}
}

func (p *Pos) NullUnmove() {
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum--
	}
}

func (p *Pos) Move(steps []Step) error {
	if p.moveNum == 1 && len(steps) != 16 {
		return fmt.Errorf("wrong number of setup moves")
	}
	initZHash := p.zhash
	for _, step := range steps {
		if err := p.Step(step); err != nil {
			return fmt.Errorf("%s: %v", step.String(), err)
		}
	}
	p.NullMove()
	// TODO(ajzaff): Movegen should filter moves that would result
	// in recurring positions.
	if initZHash == p.zhash {
		return fmt.Errorf("recurring position is illegal")
	}
	return nil
}

func (p *Pos) Unmove(steps []Step) error {
	if p.moveNum == 1 && len(steps) != 16 {
		return fmt.Errorf("wrong number of setup moves")
	}
	p.NullUnmove()
	for _, step := range steps {
		if err := p.Unstep(step); err != nil {
			return fmt.Errorf("%s: %v", step.String(), err)
		}
	}
	return nil
}
