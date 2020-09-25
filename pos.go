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
	stepsLeft  int        // steps remaining in the current move
	stack      []pushInfo // information allocated per step thats needs to be restored on unstep.
	hash       Hash       // hash of the current position
	turnHash   []Hash     // hash at the beginning of the turn used to detect repetition
}

type pushInfo struct {
	push  bool
	piece Piece
	src   Square
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
		stepsLeft:  16,
		stack:      make([]pushInfo, 1, 4*maxPly),
	}
	p.stack[0].src = 64
	p.hash = computeHash(p.bitboards, p.side, p.stepsLeft)
	p.turnHash = append(p.turnHash, p.hash)
	return p
}

var shortPosPattern = regexp.MustCompile(`^([wbgs]) \[([ RCDHMErcdhme]{64})\]$`)

// ParseShortPosition parses the position in short notation.
// The turn number is set to 2 with the provided color to move.
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
	if p.side != side {
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
		p.Place(piece, square)
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
	hashes := make([]Hash, len(p.turnHash))
	stack := make([]pushInfo, len(p.stack), cap(p.stack))
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
	copy(stack, p.stack)
	copy(hashes, p.turnHash)
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
		stepsLeft:  p.stepsLeft,
		stack:      stack,
		hash:       p.hash,
		turnHash:   hashes,
	}
}

func (p *Pos) currentMove() *Move {
	if len(p.moves) > 0 {
		return &p.moves[len(p.moves)-1]
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

// Terminal tests whether this position is a terminal node from the perspecitve of player B to move.
// Only checks for goal or elimination.
func (p *Pos) Terminal() Value {

	// Still setting up?
	if p.moveNum == 1 {
		return 0
	}

	c := p.Side()

	// Goal:
	goalA := p.bitboards[GRabbit] & ^NotRank8 != 0
	goalB := p.bitboards[SRabbit] & ^NotRank1 != 0
	if c == Gold {
		goalA, goalB = goalB, goalA
	}

	// Has a rabbit of player A reached goal? If so player A wins.
	if goalA {
		return Loss
	}

	// Has a rabbit of player B reached goal? If so player B wins.
	if goalB {
		return Win
	}

	// Elimination:
	elimA := p.bitboards[GRabbit] == 0
	elimB := p.bitboards[SRabbit] == 0
	if c == Gold {
		elimA, elimB = elimB, elimA
	}

	// Has player B lost all rabbits? If so player A wins.
	if elimB {
		return Loss
	}

	// Has player A lost all rabbits? If so player B wins.
	if elimA {
		return Win
	}

	return 0
}

// Place places piece on i. If piece is Empty it instead removes the piece.
// If a piece is already present it removes the piece first.
func (p *Pos) Place(piece Piece, i Square) {
	if piece == Empty {
		p.Remove(piece, i)
		return
	}
	b := i.Bitboard()
	if p.bitboards[Empty]&b == 0 {
		// Remove piece before placing a new one.
		p.Remove(p.At(i), i)
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
}

// Remove removes the piece from i. If no piece is present it does nothing.
func (p *Pos) Remove(piece Piece, i Square) {
	b := i.Bitboard()
	if p.bitboards[Empty]&b != 0 {
		return
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
}

// HashAfter returns the hash that would result after playing the Step s.
func (p *Pos) HashAfter(s Step) Hash {
	hash := p.Hash()
	switch {
	case s.Capture():
		hash ^= pieceHashKey(s.Piece(), s.Src())
	case s.Setup():
		hash ^= pieceHashKey(s.Piece(), s.Dest())
	default:
		t := s.Piece()
		src, dest := s.Src(), s.Dest()
		hash ^= pieceHashKey(t, s.Src())
		hash ^= pieceHashKey(t, s.Dest())
		p1 := p.Presence(t.Color()) ^ (src.Bitboard() | dest.Bitboard())
		p2 := p.Presence(t.Color().Opposite())
		if cap := p.completeCapture(p1, p2); cap != 0 {
			hash ^= pieceHashKey(cap.Piece(), cap.Src())
		}
	}
	if p.stepsLeft == 1 {
		hash ^= silverHashKey()
	}
	return hash
}

// Push returns a Square, Piece, and bool of the ongoing push if any.
func (p *Pos) Push() (src Square, piece Piece, ok bool) {
	// return isPushMove(p.Side(), p.currentMove())
	e := p.stack[len(p.stack)-1]
	return e.src, e.piece, e.push
}

var setupCounts = []uint8{0, 8, 2, 2, 2, 1, 1}

// Legal checks the legality of a step in the context of an ongoing move
// and returns ok and an error if any.
// Legal is meant to be called before playing s.
func (p *Pos) Legal(s Step) bool {

	// We don't try to validate captures.
	if s.Capture() {
		return false
	}

	// Steps left?
	if p.stepsLeft == 0 {
		return false
	}

	// Is piece valid?
	piece := s.Piece()
	if !piece.Valid() {
		return false
	}

	dest := s.Dest()

	if s.Setup() {
		// Setup move after move 1?
		if p.moveNum != 1 {
			return false
		}

		// Check setup piece:
		if piece.Color() != p.Side() {
			return false
		}
		if t := piece.RemoveColor(); p.Bitboard(piece).Count() >= setupCounts[t] {
			return false
		}

		// Check setup square:
		if c := p.Side(); c == Gold && dest > H2 || c == Silver && dest < A6 {
			return false
		}

		return true
	}

	// Is src valid?
	src := s.Src()
	if t := p.At(src); !t.Valid() {
		return false
	}

	// Is dest empty?
	if t := p.At(dest); t != Empty {
		return false
	}

	// Move to non-adjacent square?
	if src.Neighbors()&dest.Bitboard() == 0 {
		return false
	}

	if piece.Color() == p.Side() {

		// Is src frozen?
		if piece.Color() == p.Side() && p.Frozen(src) {
			return false
		}

		// Backwards rabbit move?
		direction := dest.Sub(src)
		if piece.SameType(GRabbit) &&
			(piece.Color() == Gold && direction == South ||
				piece.Color() == Silver && direction == North) {
			return false
		}

		// Check that this step completes the last push if any.
		if src, t, ok := p.Push(); ok && (dest != src || !t.WeakerThan(piece)) {
			return false
		}
	} else {
		// Step abandons ongoing push.
		lastSrc, lastPiece, ok := p.Push()
		if ok {
			return false
		}

		if lastPiece == Empty || lastSrc != dest || !piece.WeakerThan(lastPiece) {
			// Push on last step.
			if p.stepsLeft == 1 {
				return false
			}

			// Find a valid pusher.
			var strongerUnfrozen bool
			for b := src.Neighbors() & p.Stronger(piece) & p.Presence(p.Side()); b > 0; b &= b - 1 {
				if !p.Frozen(b.Square()) {
					strongerUnfrozen = true
					break
				}
			}
			if !strongerUnfrozen {
				return false
			}
		}
	}

	if p.stepsLeft == 1 {
		hashAfter := p.HashAfter(s)

		// Does this step end the turn and repeat a position for the third time?
		if p.threefold.Lookup(hashAfter) >= 3 {
			return false
		}

		// Does this step repeat the position?
		if hashAfter == p.turnHash[len(p.turnHash)-1]^silverHashKey() {
			return false
		}
	}

	return true
}

// CanPass returns true when passing the turn would be a legal move.
// It considers the number of steps taken so far, push progress,
// and threefold repetitions.
func (p *Pos) CanPass() bool {
	// Never pass during setup.
	if p.moveNum == 1 {
		return false
	}

	// We need to make at least one step.
	if p.stepsLeft == 4 {
		return false
	}

	// Are we in the middle of a push?
	if _, _, ok := p.Push(); ok {
		return false
	}

	// Would the position would repeat for a third time if we passed?
	if p.threefold.Lookup(p.Hash()^silverHashKey()) >= 3 {
		return false
	}

	return true
}

// Pass the turn and reset step variables.
func (p *Pos) Pass() {
	p.moves = append(p.moves, nil)
	p.hash ^= silverHashKey()
	p.turnHash = append(p.turnHash, p.hash)
	if p.side = p.side.Opposite(); p.side == Gold {
		p.moveNum++
	}
	p.stepsLeft = 4
	if p.moveNum == 1 {
		p.stepsLeft = 16
	}
}

// Unpass the turn and restore step variables.
func (p *Pos) Unpass() {
	if len(p.moves) < 2 {
		// No move to unpass
		return
	}
	p.moves = p.moves[:len(p.moves)-1]
	p.hash ^= silverHashKey()
	p.turnHash = p.turnHash[:len(p.turnHash)-1]
	if p.side = p.side.Opposite(); p.side == Silver {
		p.moveNum--
	}
	move := p.currentMove()
	if p.moveNum == 1 {
		p.stepsLeft = 16 - move.Len()
	} else {
		p.stepsLeft = 4 - move.Len()
	}
}

// completeCapture returns a capture step resulting from an undefended piece on a trap.
func (p *Pos) completeCapture(p1, p2 Bitboard) Step {
	nonEmpty := ^p.Empty()

	// Capture any undefended piece.
	if b := Traps&nonEmpty&^nonEmpty.Neighbors() | Traps&p1 & ^p1.Neighbors() | Traps&p2 & ^p2.Neighbors(); b != 0 {
		i := b.Square()
		return MakeCapture(p.At(i), i)
	}

	return 0
}

func (p *Pos) Step(step Step) (capture Step) {
	// captures are generated internally and we can skip them.
	if step.Capture() {
		return
	}

	p.moves[len(p.moves)-1] = append(p.moves[len(p.moves)-1], step)

	// Execute the step:
	piece, src, dest := step.Piece(), step.Src(), step.Dest()
	switch {
	case step.Setup():
		p.Place(piece, dest)
	default:
		p.Remove(piece, src)
		p.Place(piece, dest)
		// Check if any capture results and execute it:
		if capture = p.completeCapture(p.Presence(Gold), p.Presence(Silver)); capture != 0 {
			p.Remove(capture.Piece(), capture.Src())
			p.moves[len(p.moves)-1] = append(p.moves[len(p.moves)-1], capture)
		}
		// Update push stack:
		l := len(p.stack)
		if l < cap(p.stack) {
			p.stack = p.stack[:l+1]
		} else {
			p.stack = append(p.stack, pushInfo{})
		}
		var pull bool
		if piece.Color() != p.Side() {
			// Set push flag to whether this is a push.
			pull = p.stack[l-1].piece.Valid() &&
				p.stack[l-1].src == dest &&
				piece.WeakerThan(p.stack[l-1].piece)
			p.stack[l].push = !pull
		}
		if p.stack[l-1].push || pull {
			piece = 0
			dest = 64
		}
		p.stack[l].piece = piece
		p.stack[l].src = src
	}
	p.stepsLeft--
	if p.stepsLeft == 0 {
		p.Pass()
	}
	return capture
}

func (p *Pos) Unstep() {
	step, cap := p.currentMove().Pop()
	if step == 0 {
		p.Unpass()
		step, cap = p.currentMove().Pop()
		if step == 0 {
			return
		}
	}
	if cap.Capture() {
		p.Place(cap.Piece(), cap.Src())
	}
	switch {
	case step.Setup():
		p.Remove(step.Piece(), step.Dest())
	default:
		p.Remove(step.Piece(), step.Dest())
		p.Place(step.Piece(), step.Src())
	}
	p.stack = p.stack[:len(p.stack)-1]
	p.stepsLeft++
}

var errRecurringPosition = errors.New("recurring position")

func (p *Pos) Move(m Move) {
	initSide := p.Side()
	for _, step := range m {
		p.Step(step)
	}
	if p.Side() == initSide {
		p.Pass()
	}
	p.threefold.Increment(p.Hash())
}

func (p *Pos) Unmove() {
	p.threefold.Decrement(p.Hash())
	p.Unpass()
	move := p.currentMove()
	for i := len(*move) - 1; i >= 0; i-- {
		p.Unstep()
	}
}

func (p *Pos) appendShortString(sb *strings.Builder) {
	fmt.Fprintf(sb, "%c [", p.side.Byte())
	for i := 7; i >= 0; i-- {
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			sb.WriteByte(p.board[at].Byte())
		}
	}
	sb.WriteByte(']')
}

func (p *Pos) ShortString() string {
	var sb strings.Builder
	p.appendShortString(&sb)
	return sb.String()
}

func (p *Pos) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "%d%c", p.moveNum, p.side.Byte())
	if move := p.currentMove(); move != nil {
		fmt.Fprintf(sb, " %s", move.String())
	}
	sb.WriteString("\n +-----------------+\n")
	for i := 7; i >= 0; i-- {
		fmt.Fprintf(sb, "%d| ", i+1)
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
}

func (p *Pos) String() string {
	var sb strings.Builder
	p.appendString(&sb)
	return sb.String()
}
