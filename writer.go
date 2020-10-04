package zoo

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	expb "github.com/tensorflow/tensorflow/tensorflow/go/core/example/example_protos_go_proto"
)

const examplesPerBatch = 50000

// BatchWriter implements a writer capable of outputting TFRecord training data.
type BatchWriter struct {
	dir   string
	epoch int

	batchNumber     int             // batch number
	inProgress      []*expb.Example // in progress game positions
	finished        []*expb.Example // finished game positions
	feats           []float32       // position features
	latFeats        []float32       // laterally mirrored position features
	policyLabels    []float32       // policy labels
	latPolicyLabels []float32       // laterally mirrored policy labels
	valueLabel      []float32       // value label
}

// NewBatchWriter creates a new BatchWriter with the given batch size in number of games per tfrecord file.
func NewBatchWriter(epoch int) *BatchWriter {
	return &BatchWriter{
		dir:             filepath.Join("data", "training"),
		inProgress:      make([]*expb.Example, 0, 1024),
		finished:        make([]*expb.Example, 0, examplesPerBatch),
		feats:           make([]float32, 8*8*21),
		latFeats:        make([]float32, 8*8*21),
		policyLabels:    make([]float32, 231),
		latPolicyLabels: make([]float32, 231),
		valueLabel:      make([]float32, 1),
	}
}

// WriteExample writes the example trajectory to the buffer.
// To be called for each step in the game.
// Call finalize after the game is over to commit the final result.
func (w *BatchWriter) WriteExample(p *Pos, t *Tree) {
	Features(p, w.feats, w.latFeats)
	PolicyLabels(t, w.policyLabels, w.latPolicyLabels)

	feats := make([]float32, 8*8*21)
	latFeats := make([]float32, 8*8*21)
	policyLabels := make([]float32, 231)
	latPolicyLabels := make([]float32, 231)

	copy(feats, w.feats)
	copy(latFeats, w.latFeats)
	copy(policyLabels, w.policyLabels)
	copy(latPolicyLabels, w.latPolicyLabels)

	w.inProgress = append(w.inProgress, &expb.Example{
		Features: &expb.Features{
			Feature: map[string]*expb.Feature{
				"b": {Kind: &expb.Feature_FloatList{FloatList: &expb.FloatList{Value: feats}}},
				"p": {Kind: &expb.Feature_FloatList{FloatList: &expb.FloatList{Value: policyLabels}}},
				"v": {Kind: &expb.Feature_FloatList{FloatList: &expb.FloatList{Value: []float32{0} /* placeholder */}}},
			},
		},
	}, &expb.Example{
		Features: &expb.Features{
			Feature: map[string]*expb.Feature{
				"b": {Kind: &expb.Feature_FloatList{FloatList: &expb.FloatList{Value: latFeats}}},
				"p": {Kind: &expb.Feature_FloatList{FloatList: &expb.FloatList{Value: latPolicyLabels}}},
				"v": {Kind: &expb.Feature_FloatList{FloatList: &expb.FloatList{Value: []float32{0} /* placeholder */}}},
			},
		},
	})
}

var crc32c = crc32.MakeTable(crc32.Castagnoli)

// checksum returns the masked checksum of the data length.
func (w *BatchWriter) checksum(data []byte) uint32 {
	crc := crc32.Checksum(data, crc32c)
	return ((crc >> 15) | (crc << 17)) + 0xa282ead8
}

// write writes the buffered examples to a tfrecord file.
// Format of a single record:
//  uint64    length
//  uint32    masked crc of length
//  byte      data[length]
//  uint32    masked crc of data
func (w *BatchWriter) write() (err error) {
	f, err := os.OpenFile(filepath.Join(w.dir, fmt.Sprintf("games-%d.%d.tfrecord.gz", w.epoch, w.batchNumber)), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer func() {
		err1 := f.Close()
		if err == nil {
			err = err1
		}
	}()

	gz := gzip.NewWriter(f)
	defer func() {
		err1 := gz.Flush()
		if err == nil {
			err = err1
		}
	}()

	header := make([]byte, 12)
	footer := make([]byte, 4)

	for _, e := range w.finished {
		payload, err := proto.Marshal(e)
		if err != nil {
			return err
		}

		binary.LittleEndian.PutUint64(header[0:8], uint64(len(payload)))
		binary.LittleEndian.PutUint32(header[8:12], w.checksum(header[0:8]))
		binary.LittleEndian.PutUint32(footer[0:4], w.checksum(payload))

		_, err = gz.Write(header)
		if err != nil {
			return err
		}
		_, err = gz.Write(payload)
		if err != nil {
			return err
		}
		_, err = gz.Write(footer)
		if err != nil {
			return err
		}
	}
	w.batchNumber++
	w.finished = w.finished[:0]
	return nil
}

// Finalize is called after the game has completed with the result for the given side.
// The method updates all examples in memory with the final score and commits them to
// the finished examples.
func (w *BatchWriter) Finalize(t Value) error {
	for i := len(w.inProgress) - 1; i >= 0; i-- {
		w.inProgress[i].Features.Feature["v"].GetFloatList().Value[0] = float32(t)
		t = -t
	}
	w.finished = append(w.finished, w.inProgress...)
	w.inProgress = w.inProgress[:0]
	if len(w.finished) >= examplesPerBatch {
		if err := w.write(); err != nil {
			return err
		}
	}
	return nil
}
