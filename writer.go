package zoo

import (
	"fmt"
	"os"
	"path/filepath"

	zoopb "github.com/ajzaff/bot_zoo/proto"
	"github.com/golang/protobuf/proto"
)

const gamesPerBatch = 1000

// BatchWriter implements a writer capable of outputting training data.
type BatchWriter struct {
	dir   string
	epoch int

	batchNumber int               // batch number
	inProgress  *zoopb.Match_Game // in progress game
	buffered    *zoopb.Match      // buffered games
	finished    *zoopb.Match      // finished games
}

// NewBatchWriter creates a new BatchWriter with the given batch size in number of games per Dataset file.
func NewBatchWriter(epoch int) *BatchWriter {
	return &BatchWriter{
		dir:        filepath.Join("data", "training"),
		epoch:      epoch,
		inProgress: &zoopb.Match_Game{Pgn: &zoopb.PGN{}},
		buffered:   &zoopb.Match{},
		finished:   &zoopb.Match{},
	}
}

// WriteExample writes the example trajectory to the buffer.
// To be called for each step in the game.
// Call finalize after the game is over to commit the final result.
func (w *BatchWriter) WriteExample(p *Pos, t *Tree) {
	w.inProgress.Pgn.Policy = t.Root().Policy()
	for _, s := range p.MoveList().Last() {
		w.inProgress.Pgn.Steps = append(w.inProgress.Pgn.Steps, uint32(s.Index()))
	}
}

// write writes the buffered examples to a Dataset file.
// Format of a single record:
//  uint64    length
//  uint32    masked crc of length
//  byte      data[length]
//  uint32    masked crc of data
func (w *BatchWriter) write() (err error) {
	f, err := os.OpenFile(filepath.Join(w.dir, fmt.Sprintf("games.%d.%d.pb", w.epoch, w.batchNumber)), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer func() {
		err1 := f.Close()
		if err == nil {
			err = err1
		}
	}()

	payload, err := proto.Marshal(w.finished)
	if err != nil {
		return err
	}

	_, err = f.Write(payload)
	if err != nil {
		return err
	}

	w.batchNumber++
	w.finished = &zoopb.Match{}
	return nil
}

// Finalize is called after the game has completed with the result for the given side.
// The method updates all examples in memory with the final score and commits them to
// the finished examples.
func (w *BatchWriter) Finalize(p *Pos, t Value) error {
	if p.Side() == Silver {
		t = -t
	}
	w.inProgress.Pgn.Result = int32(t)
	w.finished.Games = append(w.finished.Games, w.inProgress)
	w.inProgress = &zoopb.Match_Game{Pgn: &zoopb.PGN{}}
	if len(w.finished.Games) >= gamesPerBatch {
		if err := w.write(); err != nil {
			return err
		}
	}
	return nil
}

// Flush writes the remaining examples if any to a training file.
func (w *BatchWriter) Flush() error {
	if len(w.finished.Games) > 0 {
		if err := w.write(); err != nil {
			return err
		}
	}
	return nil
}
