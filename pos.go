package zoo

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Pos represents an Arimaa position.
type Pos struct {
	board      []Piece    // board information; captures are negated such that they can be undone later
	bitboards  []Bitboard // bitboard data
	presence   []Bitboard // board presence for each side
	stronger   []Bitboard // stronger pieces by piece&decolorMask
	weaker     []Bitboard // weaker pieces by piece&decolorMask
	touching   []Bitboard // squares touched for each side
	dominating []Bitboard // squares dominated by each side (touched by a nonrabbit)
	frozen     []Bitboard // frozen squares for each (dominating) side
	threefold  *Threefold // threefold repetition store
	side       Color      // side to play
	moveNum    int        // number of moves left for this turn
	moves      MoveList   // moves to arrive at this position including the current in progress move
	lastSrc    Square     // last source square or an invalid square used for validating pulls
	stepsLeft  int        // steps remaining in the current move
	hash       Hash       // hash of the current position
}

// NewEmptyPosition creates a new initial position with no pieces and turn number 1g.
func NewEmptyPosition() *Pos {
	p := &Pos{
		board:      make([]Piece, 64),
		bitboards:  []Bitboard{AllBits, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		presence:   make([]Bitboard, 2),
		stronger:   make([]Bitboard, 8),
		weaker:     make([]Bitboard, 8),
		touching:   make([]Bitboard, 2),
		dominating: make([]Bitboard, 2),
		frozen:     make([]Bitboard, 2),
		threefold:  NewThreefold(),
		side:       Gold,
		moveNum:    1,
		moves:      []Move{nil},
		lastSrc:    64,
		stepsLeft:  16,
	}
	p.hash = computeHash(p.bitboards, p.side, p.stepsLeft)
	return p
}

var shortPosPattern = regexp.MustCompile(`^([wbgs]) \[([ RCDHMErcdhme]{64})\]$`)

// ParseShortPosition parses the position in short notation.
// The turn number is set to 2 with the provided color to move.
// Pass moves are inserted to represent previous moves.
func ParseShortPosition(s string) (*Pos, error) {
	matches := shortPosPattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("input does not match /%s/", shortPosPattern)
	}
	side, err := ParseColor(matches[1][0])
	if err != nil {
		return nil, err
	}
	p := NewEmptyPosition()
	p.Pass()
	p.Pass()
	if side != p.side {
		p.Pass()
	}
	for i, b := range []byte(matches[2]) {
		square := Square(8*(7-i/8) + i%8)
		piece, err := ParsePiece(b)
		if err != nil {
			return nil, fmt.Errorf("at %s: %v", square.String(), err)
		}
		if piece == Empty {
			continue
		}
		if err := p.Place(piece, square); err != nil {
			return nil, fmt.Errorf("at %s: %v", square.String(), err)
		}
	}
	return p, nil
}

// Clone returns a deep copy of the position p.
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
	moves := make([]Move, len(p.moves))
	copy(board, p.board)
	copy(bs, p.bitboards)
	copy(ps, p.presence)
	copy(sb, p.stronger)
	copy(wb, p.weaker)
	copy(tb, p.touching)
	copy(db, p.dominating)
	copy(fb, p.frozen)
	for i := range moves {
		moves[i] = make(Move, len(p.moves[i]))
		copy(moves[i], p.moves[i])
	}
	return &Pos{
		board:      board,
		bitboards:  bs,
		presence:   ps,
		stronger:   sb,
		weaker:     wb,
		touching:   tb,
		dominating: db,
		frozen:     fb,
		threefold:  threefold,
		side:       p.Side(),
		moveNum:    p.moveNum,
		moves:      moves,
		lastSrc:    p.lastSrc,
		stepsLeft:  p.stepsLeft,
		hash:       p.hash,
	}
}

func (p *Pos) CurrentMove() Move {
	if len(p.moves) > 0 {
		return p.moves[len(p.moves)-1]
	}
	return nil
}

func (p *Pos) Hash() Hash {
	return p.hash
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
		return fmt.Errorf("invalid piece: %c", piece.Byte())
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
	p.hash ^= pieceHashKey(piece, i)
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
	p.hash ^= pieceHashKey(piece, i)
	return nil
}

func (p *Pos) Pass() {
	p.moves = append(p.moves, nil)
	p.hash ^= silverHashKey()
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum++
	}
	p.stepsLeft = 4
	if p.moveNum == 1 {
		p.stepsLeft = 16
	}
}

func (p *Pos) Unpass() error {
	if len(p.moves) < 2 {
		return fmt.Errorf("no move to unpass")
	}
	p.moves = p.moves[:len(p.moves)-1]
	p.hash ^= silverHashKey()
	if p.side = p.side.Opposite(); p.side == Silver {
		p.moveNum--
	}
	p.stepsLeft = 4 - p.moves[len(p.moves)-1].Len()
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
	p.stepsLeft -= n
	p.moves[len(p.moves)-1] = append(p.moves[len(p.moves)-1], step)
	return nil
}

func (p *Pos) Unstep() error {
	if p.CurrentMove().Len() == 0 {
		return p.Unpass()
	}
	move := p.moves[len(p.moves)-1]
	step := move[len(move)-1]
	p.moves[len(p.moves)-1] = p.moves[len(p.moves)-1][:len(move)-1]
	n := 1
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
	initHash := p.hash
	for _, step := range m {
		if err := p.Step(step); err != nil {
			return fmt.Errorf("move %s: step %s: %v", m.String(), step.String(), err)
		}
	}
	p.Pass()
	// TODO(ajzaff): Movegen should filter moves that would result
	// in recurring positions.
	if initHash == p.Hash()^silverHashKey() {
		return errRecurringPosition
	}
	// Check threefold repetition.
	if p.threefold.Lookup(p.Hash()) >= 3 {
		return errRecurringPosition
	}
	return nil
}

func (p *Pos) Unmove() error {
	move := p.CurrentMove()
	if err := p.Unpass(); err != nil {
		return fmt.Errorf("unmove %s: unpass: %v", move, err)
	}
	for i := len(move) - 1; i >= 0; i-- {
		step := move[i]
		if err := p.Unstep(); err != nil {
			return fmt.Errorf("unmove %s: unstep %s: %v", move.String(), step.String(), err)
		}
	}
	return nil
}

const (
	posEmpty     = `g [                                                                ]`
	posStandard  = `g [rrrrrrrrhdcemcdh                                HDCMECDHRRRRRRRR]`
	posStandardG = `s [rrrrrrrrhdcemcdh                                                ]`
)

func (p *Pos) ShortString() string {
	if p == nil {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%c [", p.side.Byte())
	for i := 7; i >= 0; i-- {
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			sb.WriteByte(p.board[at].Byte())
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (p *Pos) String() string {
	if p == nil {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d%c", p.moveNum, p.side.Byte())
	if move := p.CurrentMove(); move != nil {
		fmt.Fprintf(&sb, " %s", move.String())
	}
	sb.WriteString("\n +-----------------+\n")
	for i := 7; i >= 0; i-- {
		fmt.Fprintf(&sb, "%d| ", i+1)
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			if piece := p.board[at]; piece == Empty {
				if at.Bitboard()&Traps != 0 {
					sb.WriteByte('x')
				} else {
					sb.WriteByte('.')
				}
			} else {
				sb.WriteByte(piece.Byte())
			}
			sb.WriteByte(' ')
		}
		sb.WriteString("|\n")
	}
	sb.WriteString(" +-----------------+\n   a b c d e f g h")
	return sb.String()
}
