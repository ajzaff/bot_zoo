package zoo

import "fmt"

type Pos struct {
	Bitboards []Bitboard
	Presence  []Bitboard
	Side      Color
	Steps     int
	Push      bool
	LastPiece Piece
	LastFrom  Square
	ZHash     int64
}

func NewPos(
	bitboards []Bitboard,
	presence []Bitboard,
	side Color,
	steps int,
	push bool,
	lastPiece Piece,
	lastFrom Square,
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
	return -1
}

func (p *Pos) RabbitLoss() Color {
	if p.Bitboards[GRabbit] == 0 {
		return Gold
	}
	if p.Bitboards[SRabbit] == 0 {
		return Silver
	}
	return -1
}

func (p *Pos) Terminal() bool {
	return p.Goal() != -1 || p.RabbitLoss() != -1
}

func (p *Pos) At(i Square) Piece {
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

func (p *Pos) Place(piece Piece, i Square) (*Pos, error) {
	if piece == Empty {
		return p.Remove(i)
	}
	b := Bitboard(1) << i
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
	zhash := p.ZHash ^ ZPieceKey(piece, i)
	return NewPos(
		bs, ps, p.Side, p.Steps, p.Push,
		p.LastPiece, p.LastFrom, zhash,
	), nil
}

func (p *Pos) Remove(i Square) (*Pos, error) {
	b := Bitboard(1) << i
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
	zhash := p.ZHash ^ ZPieceKey(piece, i)
	return NewPos(
		bs, ps, p.Side, p.Steps, p.Push,
		p.LastPiece, p.LastFrom, zhash,
	), nil
}

func (p *Pos) Frozen(i Square) bool {
	return p.frozenB(1 << i)
}

func (p *Pos) frozenB(b Bitboard) bool {
	neighbors := b.Neighbors()
	piece := p.atB(b)
	color := piece.Color()
	if neighbors&p.Presence[color] == 0 &&
		neighbors&p.Presence[color.Opposite()] != 0 {
		for s := piece.MakeColor(color.Opposite()) + 1; s <= GElephant.MakeColor(color.Opposite()); s++ {
			if neighbors&p.Bitboards[s] != 0 {
				return true
			}
		}
	}
	return false
}

func (p *Pos) FrozenNeighbors(b Bitboard) Bitboard {
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

func (p *Pos) CheckStep(step Step) (ok bool, err error) {
	src := Bitboard(1) << step.Src
	dest := Bitboard(1) << step.Dest
	if src&p.Bitboards[Empty] != 0 {
		return false, fmt.Errorf("move from empty square")
	}
	if dest&p.Bitboards[Empty] == 0 {
		return false, fmt.Errorf("move to nonempty square")
	}
	piece := p.atB(src)
	dir := step.Dest - step.Src
	srcNeighbors := src.Neighbors()
	destNeighbors := dest.Neighbors()

	if src&destNeighbors == 0 {
		return false, fmt.Errorf("move to nonadjacent square")
	}
	if piece.Color() == p.Side {
		if p.frozenB(src) {
			return false, fmt.Errorf("move from frozen piece")
		}
		if piece.SamePiece(GRabbit) && (piece.Color() == Gold && dir < 0 || piece.Color() == Silver && dir > 0) {
			return false, fmt.Errorf("backwards rabbit move")
		}
		if p.Push {
			if piece.WeakerThan(p.LastPiece) {
				return false, fmt.Errorf("piece is too weak to push")
			}
			if step.Dest != p.LastFrom {
				return false, fmt.Errorf("move must finish active push")
			}
		}
		return true, nil
	}
	if p.Push {
		return false, fmt.Errorf("push started before active push finished")
	}
	if p.Steps == 3 {
		return false, fmt.Errorf("push started on last step")
	}
	if p.LastPiece == Empty ||
		p.LastFrom != step.Dest ||
		p.LastPiece.WeakerThan(piece) ||
		p.LastPiece.SamePiece(piece) {
		foundPusher := false
		for s := piece.MakeColor(piece.Color().Opposite()) + 1; s <= GElephant.MakeColor(p.Side); s++ {
			if srcNeighbors&p.Bitboards[s]&^p.FrozenNeighbors(src) != 0 {
				foundPusher = true
				break
			}
		}
		if !foundPusher {
			return false, fmt.Errorf("move with no pusher")
		}
	}
	return true, nil
}

func (p *Pos) Step(step Step) *Pos {
	bs := make([]Bitboard, len(p.Bitboards))
	for i := range p.Bitboards {
		bs[i] = p.Bitboards[i]
	}
	ps := make([]Bitboard, 2)
	ps[0] = p.Presence[0]
	ps[1] = p.Presence[1]
	zhash := p.ZHash
	src := Bitboard(1) << step.Src
	dest := Bitboard(1) << step.Dest
	piece := p.atB(src)
	var push, pull bool
	if piece.Color() != p.Side {
		if p.LastPiece != Empty && p.LastFrom == step.Dest {
			pull = true
		} else {
			push = true
		}
	}
	stepB := src | dest
	bs[piece] ^= stepB
	bs[Empty] ^= stepB
	ps[piece.Color()] ^= stepB
	zhash ^= ZPieceKey(piece, step.Src)
	zhash ^= ZPieceKey(piece, step.Dest)
	trappedB := src.Neighbors() & Traps & ^ps[piece.Color()]
	if trappedB&ps[piece.Color()] != 0 {
		bs[Empty] |= trappedB
		ps[piece.Color()] &= ^trappedB
		for t := GRabbit.MakeColor(piece.Color()); t <= GElephant.MakeColor(piece.Color()); t++ {
			if bs[t]&trappedB != 0 {
				zhash ^= ZPieceKey(t, trappedB.Square())
				bs[t] &= ^trappedB
				break // only one trapped piece possible
			}
		}
	}
	steps := p.Steps + 1
	side := p.Side
	if steps < 1 {
		p.Side = p.Side.Opposite()
		steps = 0
		piece = Empty
		src = 0
	}
	if p.Push || pull {
		piece = Empty
		src = 0
	}
	return NewPos(
		bs, ps, side, steps, push,
		piece, src.Square(), zhash,
	)
}
func (p *Pos) NullMove() *Pos {
	side := p.Side.Opposite()
	zhash := p.ZHash
	zhash ^= ZSilverKey()
	return NewPos(
		p.Bitboards, p.Presence, side,
		0, false, Empty, 0, zhash,
	)
}

func (p *Pos) Move(steps []Step, check bool) (*Pos, error) {
	if p.Steps+len(steps) > 4 {
		return nil, fmt.Errorf("tried to take more than 4 steps")
	}
	initZHash := p.ZHash
	side := p.Side
	for i, step := range steps {
		if check {
			legal, err := p.CheckStep(step)
			if !legal {
				return nil, fmt.Errorf("move %d: %v", i+1, err)
			}
			p = p.Step(step)
		}
	}
	if p.Side == side {
		side = side.Opposite()
		zhash := p.ZHash
		zhash ^= ZSilverKey()
		zhash ^= ZStepsKey(p.Steps)
		zhash ^= ZStepsKey(0)
		p = NewPos(
			p.Bitboards, p.Presence, side,
			0, false, Empty, 0, zhash,
		)
	}
	// TODO(ajzaff): This doesn't work yet.
	if initZHash == p.ZHash^ZSilverKey() {
		return nil, fmt.Errorf("recurring position is illegal")
	}
	return p, nil
}
