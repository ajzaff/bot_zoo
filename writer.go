package zoo

import (
	"fmt"
	"os"
	"path/filepath"

	expb "github.com/ajzaff/bot_zoo/proto"
	"github.com/golang/protobuf/proto"
)

const examplesPerBatch = 100000

// BatchWriter implements a writer capable of outputting training data.
type BatchWriter struct {
	dir   string
	epoch int

	batchNumber int            // batch number
	inProgress  *expb.Examples // in progress game positions
	finished    *expb.Examples // finished game positions
}

// NewBatchWriter creates a new BatchWriter with the given batch size in number of games per Dataset file.
func NewBatchWriter(epoch int) *BatchWriter {
	return &BatchWriter{
		dir:        filepath.Join("data", "training"),
		epoch:      epoch,
		inProgress: &expb.Examples{},
		finished:   &expb.Examples{},
	}
}

// WriteExample writes the example trajectory to the buffer.
// To be called for each step in the game.
// Call finalize after the game is over to commit the final result.
func (w *BatchWriter) WriteExample(p *Pos, t *Tree) {
	var ex expb.Example
	Features(p, &ex)
	PolicyLabels(t, &ex)
	w.inProgress.Examples = append(w.inProgress.Examples, &ex)
}

// write writes the buffered examples to a Dataset file.
// Format of a single record:
//  uint64    length
//  uint32    masked crc of length
//  byte      data[length]
//  uint32    masked crc of data
func (w *BatchWriter) write() (err error) {
	f, err := os.OpenFile(filepath.Join(w.dir, fmt.Sprintf("examples.%d.%d.pb", w.epoch, w.batchNumber)), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
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
	w.finished = &expb.Examples{}
	return nil
}

// Finalize is called after the game has completed with the result for the given side.
// The method updates all examples in memory with the final score and commits them to
// the finished examples.
func (w *BatchWriter) Finalize(p *Pos, t Value) error {
	p = p.Clone()
	initSide := p.Side()
	for i := len(w.inProgress.Examples) - 1; i >= 0; i-- {
		w.inProgress.Examples[i].Value = float32(t)
		p.Unstep()
		if c := p.Side(); c != initSide {
			t = -t
			initSide = c
		}
	}
	w.finished.Examples = append(w.finished.Examples, w.inProgress.Examples...)
	w.inProgress = &expb.Examples{}
	if len(w.finished.Examples) >= examplesPerBatch {
		if err := w.write(); err != nil {
			return err
		}
	}
	return nil
}

// Flush writes the remaining examples if any to a training file.
func (w *BatchWriter) Flush() error {
	if len(w.finished.Examples) > 0 {
		if err := w.write(); err != nil {
			return err
		}
	}
	return nil
}
