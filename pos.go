package zoo

import (
	"fmt"
)

type Pos struct {
	Bitboards []Bitboard
	Presence  []Bitboard
	Side      Color
	MoveNum   int
	// TODO(ajzaff): Add numSteps field for efficiency.
	Steps     []Step
	Push      bool
	LastPiece Piece
	LastFrom  Square
	ZHash     int64
}

func NewPos(
	bitboards []Bitboard,
	presence []Bitboard,
	side Color,
	moveNum int,
	steps []Step,
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
		zhash = ZHash(bitboards, side, CountSteps(steps))
	}
	return &Pos{
		Bitboards: bitboards,
		Presence:  presence,
		Side:      side,
		MoveNum:   moveNum,
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
	return p.atB(i.Bitboard())
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
	b := i.Bitboard()
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
		bs, ps, p.Side, p.MoveNum, p.Steps, p.Push,
		p.LastPiece, p.LastFrom, zhash,
	), nil
}

func (p *Pos) Remove(i Square) (*Pos, error) {
	b := i.Bitboard()
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
		bs, ps, p.Side, p.MoveNum, p.Steps, p.Push,
		p.LastPiece, p.LastFrom, zhash,
	), nil
}

func (p *Pos) Frozen(i Square) bool {
	return p.frozenB(i.Bitboard())
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
	if p.MoveNum == 1 {
		if !step.Setup() {
			return false, fmt.Errorf("expected setup move")
		}
		return true, nil
	}
	src := step.Src.Bitboard()
	dest := step.Dest.Bitboard()
	if src&p.Bitboards[Empty] != 0 {
		return false, fmt.Errorf("move from empty square")
	}
	if dest&p.Bitboards[Empty] == 0 {
		return false, fmt.Errorf("move to nonempty square")
	}
	piece := p.atB(src)
	dir := step.Src.Delta(step.Dest)
	srcNeighbors := src.Neighbors()
	destNeighbors := dest.Neighbors()

	if src&destNeighbors == 0 {
		return false, fmt.Errorf("move to nonadjacent square")
	}
	if piece.Color() == p.Side {
		if p.frozenB(src) {
			return false, fmt.Errorf("move from frozen piece")
		}
		if piece.SamePiece(GRabbit) && (piece.Color() == Gold && dir == -8 || piece.Color() == Silver && dir == 8) {
			return false, fmt.Errorf("backwards rabbit move")
		}
		if p.Push {
			if step.Dest != p.LastFrom {
				return false, fmt.Errorf("move must finish active push")
			}
			if !p.LastPiece.WeakerThan(piece) {
				return false, fmt.Errorf("piece is too weak to push")
			}
		}
		return true, nil
	}
	if p.Push {
		return false, fmt.Errorf("push started before active push finished")
	}
	if p.LastPiece == Empty ||
		p.LastFrom != step.Dest ||
		p.LastPiece.WeakerThan(piece) ||
		p.LastPiece.SamePiece(piece) {
		if CountSteps(p.Steps) == 3 {
			return false, fmt.Errorf("push started on last step")
		}
		foundPusher := false
		for s := piece.MakeColor(piece.Color().Opposite()) + 1; s <= GElephant.MakeColor(piece.Color().Opposite()); s++ {
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

func (p *Pos) Step(step Step) (rp *Pos, cap Step, err error) {
	if step.Setup() {
		p, err = p.Place(step.Piece, step.Dest)
		return p, Step{}, err
	}
	bs := make([]Bitboard, len(p.Bitboards))
	for i := range p.Bitboards {
		bs[i] = p.Bitboards[i]
	}
	ps := make([]Bitboard, 2)
	ps[0] = p.Presence[0]
	ps[1] = p.Presence[1]
	zhash := p.ZHash
	src := step.Src.Bitboard()
	dest := step.Dest.Bitboard()
	piece := p.atB(src)
	var push, pull bool
	if piece.Color() != p.Side {
		if p.LastPiece != Empty && p.LastFrom == step.Dest &&
			piece.WeakerThan(p.LastPiece) {
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
	trappedB := src.Neighbors() & Traps & ^ps[piece.Color()].Neighbors()
	if trappedB&ps[piece.Color()] != 0 {
		bs[Empty] |= trappedB
		ps[piece.Color()] &= ^trappedB
		for t := GRabbit.MakeColor(piece.Color()); t <= GElephant.MakeColor(piece.Color()); t++ {
			if bs[t]&trappedB != 0 {
				src := trappedB.Square()
				zhash ^= ZPieceKey(t, src)
				bs[t] &= ^trappedB
				cap = Step{
					Src:   src,
					Piece: t,
					Dir:   "x",
				}
				break // only one trapped piece possible
			}
		}
	}
	steps := make([]Step, len(p.Steps))
	copy(steps, p.Steps)
	steps = append(steps, step)
	if cap.Capture() {
		steps = append(steps, cap)
	}
	side := p.Side
	moveNum := p.MoveNum
	if CountSteps(steps) > 3 {
		p.Side = p.Side.Opposite()
		if p.Side == Gold {
			moveNum++
		}
		steps = steps[:0]
		piece = Empty
		src = 0
	}
	if p.Push || pull {
		piece = Empty
		src = 0
	}
	return NewPos(
		bs, ps, side, moveNum, steps, push,
		piece, src.Square(), zhash,
	), cap, nil
}
func (p *Pos) NullMove() *Pos {
	side := p.Side.Opposite()
	zhash := p.ZHash
	zhash ^= ZSilverKey()
	moveNum := p.MoveNum
	if side == Gold {
		moveNum++
	}
	return NewPos(
		p.Bitboards, p.Presence, side,
		moveNum, nil, false, Empty, 0, zhash,
	)
}

func (p *Pos) Move(steps []Step, check bool) (rp *Pos, out []Step, err error) {
	if p.MoveNum == 1 {
		if len(steps) != 16 {
			return nil, nil, fmt.Errorf("wrong number of setup moves")
		}
	} else { // Prune captures
		newSteps := make([]Step, 0, len(steps))
		for _, step := range steps {
			if !step.Capture() {
				newSteps = append(newSteps, step)
			}
		}
		if len(newSteps) == 0 || len(newSteps) > 4 {
			return nil, nil, fmt.Errorf("wrong number of steps")
		}
		steps = newSteps
	}
	initZHash := p.ZHash
	side := p.Side
	for i, step := range steps {
		if check {
			legal, err := p.CheckStep(step)
			if !legal {
				return nil, nil, fmt.Errorf("check %d: %v", i+1, err)
			}
			out = append(out, step)
			var cap Step
			p, cap, err = p.Step(step)
			if cap.Capture() {
				out = append(out, cap)
			}
			if err != nil {
				return nil, nil, fmt.Errorf("step %d: %v", i+1, err)
			}

		}
	}
	if p.Side == side {
		side = side.Opposite()
		moveNum := p.MoveNum
		if side == Gold {
			moveNum++
		}
		zhash := p.ZHash
		zhash ^= ZSilverKey()
		zhash ^= ZStepsKey(CountSteps(p.Steps))
		zhash ^= ZStepsKey(0)
		p = NewPos(
			p.Bitboards, p.Presence, side,
			moveNum, nil, false, Empty, 0, zhash,
		)
	}
	// TODO(ajzaff): This doesn't work yet.
	if initZHash == p.ZHash {
		return nil, nil, fmt.Errorf("recurring position is illegal")
	}
	return p, out, nil
}
