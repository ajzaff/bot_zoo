package zoo

import (
	"errors"
	"fmt"
)

type Pos struct {
	board      []Piece    // board information; captures are negated such that they can be undone later
	bitboards  []Bitboard // bitboard data
	presence   []Bitboard // board presence for each side
	stronger   []Bitboard // stronger pieces by piece&decolorMask
	weaker     []Bitboard // weaker pieces by piece&decolorMask
	touching   []Bitboard // squares touched for each side
	dominating []Bitboard // squares dominated by each side (touched by a nonrabbit)
	threefold  *Threefold // threefold repetition store
	frozen     []Bitboard // frozen squares for each (dominating) side
	side       Color      // side to play
	depth      int        // number of steps to arrive at the position
	moveNum    int        // number of moves left for this turn
	moves      [][]Step   // moves to arrive at this position after appending steps
	steps      []Step     // steps of the current move
	stepsLeft  int        // steps remaining in the current move
	zhash      uint64     // zhash of the current position
}

func newPos(
	board []Piece,
	bitboards,
	presence,
	stronger,
	weaker,
	touching,
	dominating,
	frozen []Bitboard,
	threefold *Threefold,
	side Color,
	depth int,
	moveNum int,
	moves [][]Step,
	steps []Step,
	stepsLeft int,
	zhash uint64,
) *Pos {
	if board == nil {
		board = make([]Piece, 64)
	}
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
			presence[Silver] |= bitboards[p.WithColor(Silver)]
		}
	}
	if stronger == nil {
		stronger = make([]Bitboard, 8)
	}
	if weaker == nil {
		weaker = make([]Bitboard, 8)
	}
	if touching == nil {
		touching = make([]Bitboard, 2)
	}
	if dominating == nil {
		dominating = make([]Bitboard, 2)
	}
	if frozen == nil {
		frozen = make([]Bitboard, 2)
	}
	if threefold == nil {
		threefold = NewThreefold()
	}
	if zhash == 0 {
		zhash = ZHash(bitboards, side, len(steps))
	}
	return &Pos{
		board:      board,
		bitboards:  bitboards,
		presence:   presence,
		stronger:   stronger,
		weaker:     weaker,
		touching:   touching,
		dominating: dominating,
		frozen:     frozen,
		threefold:  threefold,
		side:       side,
		stepsLeft:  stepsLeft,
		depth:      depth,
		moveNum:    moveNum,
		steps:      steps,
		zhash:      zhash,
	}
}

func (p *Pos) Clone() *Pos {
	board := make([]Piece, 64)
	bs := make([]Bitboard, 15)
	ps := make([]Bitboard, 2)
	sb := make([]Bitboard, 8)
	wb := make([]Bitboard, 8)
	tb := make([]Bitboard, 2)
	db := make([]Bitboard, 2)
	fb := make([]Bitboard, 2)
	threefold := p.threefold.Clone()
	steps := make([]Step, len(p.steps))
	moves := make([][]Step, len(p.moves))
	copy(board, p.board)
	copy(bs, p.bitboards)
	copy(ps, p.presence)
	copy(sb, p.stronger)
	copy(wb, p.weaker)
	copy(tb, p.touching)
	copy(db, p.dominating)
	copy(fb, p.frozen)
	copy(steps, p.steps)
	for i := range moves {
		moves[i] = make([]Step, len(p.moves[i]))
		copy(moves[i], p.moves[i])
	}
	return newPos(
		board, bs, ps, sb, wb, tb, db, fb, threefold,
		p.side, p.depth, p.moveNum, moves, steps, p.stepsLeft, p.zhash,
	)
}

func (p *Pos) ZHash() uint64 {
	return p.zhash
}

func (p *Pos) Depth() int {
	return p.depth
}

func (p *Pos) Side() Color {
	return p.side
}

func (p *Pos) updateFrozen() {
	p.frozen[Gold] = p.dominating[Silver] & ^p.touching[Gold]
	p.frozen[Silver] = p.dominating[Gold] & ^p.touching[Silver]
}

func (p *Pos) frozenB(t Piece, b Bitboard) bool {
	return !t.SameType(GElephant) && p.frozen[t.Color()]&(p.presence[t.Color().Opposite()]&p.stronger[t.WithColor(Gold)]).Neighbors()&b != 0
}

func (p *Pos) Frozen(i Square) bool {
	return p.frozenB(p.board[i], i.Bitboard())
}

func (p *Pos) Nonempty() Bitboard {
	return ^p.bitboards[Empty]
}

func (p *Pos) Empty() Bitboard {
	return p.bitboards[Empty]
}

func (p *Pos) Presence(c Color) Bitboard {
	return p.presence[c]
}

func (p *Pos) Bitboard(t Piece) Bitboard {
	return p.bitboards[t]
}

func (p *Pos) At(i Square) Piece {
	return p.board[i]
}

func (p *Pos) Stronger(t Piece) Bitboard {
	return p.stronger[t.RemoveColor()]
}

func (p *Pos) Weaker(t Piece) Bitboard {
	return p.weaker[t.RemoveColor()]
}

func (p *Pos) Place(piece Piece, i Square) error {
	if piece == Empty {
		return p.Remove(piece, i)
	}
	if !piece.Valid() {
		return fmt.Errorf("invalid piece: %s", piece)
	}
	b := i.Bitboard()
	if p.bitboards[Empty]&b == 0 {
		return fmt.Errorf("piece already present on %s", i)
	}
	c := piece.Color()
	p.board[i] = piece
	p.bitboards[piece] |= b
	p.bitboards[Empty] &= ^b
	p.presence[c] |= b
	p.touching[c] = p.presence[c].Neighbors()
	p.dominating[c] = (p.presence[c] & ^p.bitboards[GRabbit.WithColor(c)]).Neighbors()
	p.updateFrozen()
	for r := GRabbit; r < piece.RemoveColor(); r++ {
		p.stronger[r] |= b
	}
	for r := piece.RemoveColor() + 1; r <= GElephant; r++ {
		p.weaker[r] |= b
	}
	p.zhash ^= ZPieceKey(piece, i)
	return nil
}

func (p *Pos) Remove(piece Piece, i Square) error {
	b := i.Bitboard()
	if p.bitboards[Empty]&b != 0 {
		return fmt.Errorf("no piece present on %s", i)
	}
	c := piece.Color()
	notB := ^b
	p.board[i] = Empty
	p.bitboards[piece] &= notB
	p.bitboards[Empty] |= b
	p.presence[c] &= notB
	p.touching[c] = p.presence[c].Neighbors()
	p.dominating[c] = (p.presence[c] & ^p.bitboards[GRabbit.WithColor(c)]).Neighbors()
	p.updateFrozen()
	for r := GRabbit; r < piece.RemoveColor(); r++ {
		p.stronger[r] &= notB
	}
	for r := piece.RemoveColor() + 1; r <= GElephant; r++ {
		p.weaker[r] &= notB
	}
	p.zhash ^= ZPieceKey(piece, i)
	return nil
}

func (p *Pos) Pass() {
	p.depth++
	p.moves = append(p.moves, p.steps)
	p.steps = nil
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum++
	}
	p.stepsLeft = 4
	if p.moveNum == 1 {
		p.stepsLeft = 16
	}
}

func (p *Pos) Unpass() error {
	if len(p.moves) == 0 {
		return fmt.Errorf("no move to unpass")
	}
	if len(p.steps) != 0 {
		return fmt.Errorf("steps were made since passing")
	}
	p.depth--
	p.steps = p.moves[len(p.moves)-1]
	p.moves = p.moves[:len(p.moves)-1]
	p.zhash ^= ZSilverKey()
	if p.side = p.side.Opposite(); p.side == Silver {
		p.moveNum--
	}
	p.stepsLeft = 4 - Move(p.steps).Len()
	if p.moveNum == 1 {
		p.stepsLeft = 16
	}
	return nil
}

func (p *Pos) Step(step Step) error {
	n := 1
	if n > p.stepsLeft {
		return fmt.Errorf("%s: not enough steps left", step)
	}
	switch {
	case step.Capture():
		if err := p.Remove(step.Piece(), step.Src()); err != nil {
			return fmt.Errorf("%s (%s): %v", step.GoString(), step, err)
		}
	case step.Setup():
		if err := p.Place(step.Piece(), step.Dest()); err != nil {
			return fmt.Errorf("%s (%s): %v", step.GoString(), step, err)
		}
	default:
		if err := p.Remove(step.Piece(), step.Src()); err != nil {
			return fmt.Errorf("%s (%s): %v", step.GoString(), step, err)
		}
		if err := p.Place(step.Piece(), step.Dest()); err != nil {
			return fmt.Errorf("%s (%s): %v", step.GoString(), step, err)
		}
	}
	p.depth += n
	p.stepsLeft -= n
	p.steps = append(p.steps, step)
	return nil
}

func (p *Pos) Unstep() error {
	if len(p.steps) == 0 {
		return p.Unpass()
	}
	step := p.steps[len(p.steps)-1]
	p.steps = p.steps[:len(p.steps)-1]
	n := 1
	p.depth -= n
	p.stepsLeft += n
	switch {
	case step.Capture():
		return p.Place(step.Piece(), step.Src())
	case step.Setup():
		if err := p.Remove(step.Piece(), step.Dest()); err != nil {
			return err
		}
	default:
		if err := p.Remove(step.Piece(), step.Dest()); err != nil {
			return fmt.Errorf("%s (%s): %v", step.GoString(), step, err)
		}
		if err := p.Place(step.Piece(), step.Src()); err != nil {
			return fmt.Errorf("%s (%s): %v", step.GoString(), step, err)
		}
	}
	return nil
}

var errRecurringPosition = errors.New("recurring position")

func (p *Pos) Move(m Move) error {
	if p.moveNum == 1 && m.Len() != 16 {
		return fmt.Errorf("move %s: wrong number of setup moves", m.String())
	}
	initZHash := p.zhash
	for _, step := range m {
		if err := p.Step(step); err != nil {
			return fmt.Errorf("move %s: step %s: %v", m.String(), step.String(), err)
		}
	}
	p.Pass()
	// TODO(ajzaff): Movegen should filter moves that would result
	// in recurring positions.
	if initZHash == p.zhash^ZSilverKey() {
		return errRecurringPosition
	}
	// Check threefold repetition.
	if p.threefold.Lookup(p.zhash) >= 3 {
		return errRecurringPosition
	}
	return nil
}

func (p *Pos) Unmove() error {
	if err := p.Unpass(); err != nil {
		return fmt.Errorf("unmove %s: unpass: %v", err)
	}
	for i := len(p.steps) - 1; i >= 0; i-- {
		step := p.steps[i]
		if err := p.Unstep(); err != nil {
			return fmt.Errorf("unmove %s: unstep %s: %v", Move(p.steps).String(), step.String(), err)
		}
	}
	return nil
}
