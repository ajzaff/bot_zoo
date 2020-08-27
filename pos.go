package zoo

import "fmt"

type Pos struct {
	Bitboards []Bitboard
	Presence  []Bitboard
	Side      Color
	Steps     int
	Push      bool
	Frozen    bool
	LastPiece Piece
	LastFrom  uint8
	ZHash     int64
}

func NewPos(
	bitboards []Bitboard,
	presence []Bitboard,
	side Color,
	steps int,
	push bool,
	lastPiece Piece,
	lastFrom uint8,
	zhash int64,
) *Pos {
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
			presence[Gold] |= bitboards[p.MakeSilver()]
		}
	}
	if zhash == 0 {
		zhash = ZHash(bitboards, side, steps)
	}
	return &Pos{
		Bitboards: bitboards,
		Presence:  presence,
		Side:      side,
		Steps:     steps,
		Push:      push,
		LastPiece: lastPiece,
		LastFrom:  lastFrom,
		ZHash:     zhash,
	}
}

func (p *Pos) Goal() Color {
	if p.Bitboards[GRabbit] & ^NotRank8 != 0 {
		return Gold
	}
	if p.Bitboards[SRabbit] & ^NotRank1 != 0 {
		return Silver
	}
	return Neither
}

func (p *Pos) RabbitLoss() Color {
	if p.Bitboards[GRabbit] == 0 {
		return Gold
	}
	if p.Bitboards[SRabbit] == 0 {
		return Silver
	}
	return Neither
}

func (p *Pos) Terminal() bool {
	return p.Goal() != Neither || p.RabbitLoss() != Neither
}

func (p *Pos) At(i uint8) Piece {
	return p.atB(1 << i)
}

func (p *Pos) atB(b Bitboard) Piece {
	for piece := GRabbit; piece <= SElephant && int(piece) < len(p.Bitboards); piece++ {
		if p.Bitboards[piece]&b != 0 {
			return piece
		}
	}
	return Empty
}

func (p *Pos) Place(piece Piece, idx uint8) (*Pos, error) {
	b := Bitboard(1) << idx
	if p.Bitboards[Empty]&b == 0 {
		return nil, fmt.Errorf("piece already present")
	}
	bs := make([]Bitboard, len(p.Bitboards))
	for i := range p.Bitboards {
		bs[i] = p.Bitboards[i]
	}
	bs[piece] |= b
	bs[Empty] &= ^b
	ps := make([]Bitboard, 2)
	ps[0] = p.Presence[0]
	ps[1] = p.Presence[1]
	ps[piece.Color()] |= b
	zhash := p.ZHash ^ zkeys[1+5+piece*64]
	return NewPos(
		bs, ps, p.Side, p.Steps, p.Push,
		p.LastPiece, p.LastFrom, zhash,
	), nil
}

func (p *Pos) Remove(idx uint8) (*Pos, error) {
	b := Bitboard(1) << idx
	if p.Bitboards[Empty]&b == 0 {
		return nil, fmt.Errorf("no piece present")
	}
	piece := GRabbit
	for piece <= SElephant && (p.Bitboards[piece] == 0 || p.Bitboards[piece]&b == 0) {
		piece++
	}
	if piece > SElephant {
		panic("inconsistent board state")
	}
	bs := make([]Bitboard, len(p.Bitboards))
	for i := range p.Bitboards {
		bs[i] = p.Bitboards[i]
	}
	bs[piece] &= ^b
	bs[Empty] |= b
	ps := make([]Bitboard, 2)
	ps[0] = p.Presence[0]
	ps[1] = p.Presence[1]
	ps[piece.Color()] &= ^b
	zhash := p.ZHash ^ zkeys[1+5+piece*64]
	return NewPos(
		bs, ps, p.Side, p.Steps, p.Push,
		p.LastPiece, p.LastFrom, zhash,
	), nil
}

func (p *Pos) Frozen(i uint8) bool {
	return p.frozenB(1 << i)
}

func (p *Pos) frozenB(b Bitboard) bool {
	neighbors := b.Neighbors()
	piece := p.atB(b)
	color := piece.Color()
	frozen := neighbors&p.Presence[color] == 0 &&
		neighbors&p.Presence[color^1] != 0
	if frozen {
		frozen := false
		for s := (piece ^ Piece.COLOR) + 1; s < (Piece.GELEPHANT|(pcbit^Piece.COLOR))+1; s++ {
			if neighbors & bitboards[s] {
				return true
			}
		}
	}
	return false
}

func (p *Pos) CheckStep(step Step) (ok bool, err error) {
	src := Bitboard(1) << step.Src
	dest := Bitboard(1) << step.Dest
	if src&p.Bitboards[Empty] != 0 {
		return false, fmt.Errorf("move from empty square")
	}
	if dest&p.Bitboards[Empty] != 0 {
		return false, fmt.Errorf("move to nonempty square")
	}
	piece := p.atB(src)
	dir := step.Dest - step.Src // fixme

	srcNeighbors := src.Neighbors()
	destNeighbors := dest.Neighbors()
	if srcNeighbors&destNeighbors == 0 {
		return false, fmt.Errorf("move to nonadjacent square")
	}

	// pcbit = piece & Piece.COLOR
	// pcolor = pcbit >> 3
	// pstrength = piece & Piece.DECOLOR
	// if pcolor == self.color:
	//     if self.is_frozen_at(from_bit):
	//         return BadStep("Tried to move a frozen piece")
	//     if pstrength == Piece.GRABBIT:
	//         if ((pcolor == Color.GOLD and direction == -8) or
	//             (pcolor == Color.SILVER and direction == 8)):
	//             return BadStep("Tried to move a rabbit back")
	//     if self.inpush:
	//         if pstrength <= self.last_piece & Piece.DECOLOR:
	//             return BadStep("Tried to push with too weak of a piece")
	//         if self.last_from != step[1]:
	//             return BadStep("Tried to neglect finishing a push")
	// else:
	//     if self.inpush:
	//         return BadStep(
	//             "Tried to move opponent piece while already in push")
	//     if (self.last_piece == Piece.EMPTY or self.last_from != step[1] or
	//         pstrength >= self.last_piece & Piece.DECOLOR):
	//         if self.stepsLeft == 1:
	//             return BadStep("Tried to start a push on the last step")
	//         stronger_and_unfrozen = False
	//         for s in xrange((piece ^ Piece.COLOR) + 1,
	//                         (Piece.GELEPHANT | (self.color << 3)) + 1):
	//             if from_neighbors & bitboards[s] & \
	//                     (~self.frozen_neighbors(from_bit)):
	//                 stronger_and_unfrozen = True
	//                 break
	//         if not stronger_and_unfrozen:
	//             return BadStep(
	//                 "Tried to push a piece with no pusher around")
	// return True
	return false, nil
}
