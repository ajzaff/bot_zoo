package zoo

import (
	"math/rand"
	"sync"
	"time"
)

type searchState struct {
	tree *Tree
	tt   *TranspositionTable

	model       ModelInterface
	batchWriter BatchWriterInterface

	wg       sync.WaitGroup
	bestMove Move

	// semi-atomic
	stopping int32
	running  int32
}

// ModelInterface defines an interface for a model.
type ModelInterface interface {
	EvaluatePosition(p *Pos)
	SetSeed(seed int64)
	Value() float32
	Policy(policy []float32)
}

// BatchWriterInterface defines an interface for writing TFRecord batch data.
type BatchWriterInterface interface {
	WriteExample(p *Pos, t *Tree)
	Finalize(Value) error
}

func (s *searchState) Reset() {
	if s.tt == nil {
		s.tt = &TranspositionTable{}
	}
	s.tt.Resize(50)
	if s.tree == nil {
		s.tree = NewEmptyTree(s.tt)
	}
	s.model = NewDummyModel()
	s.wg = sync.WaitGroup{}
	s.stopping = 0
	s.running = 0
}

func (e *Engine) searchRoot(ponder bool) {
	defer e.Stop()

	if e.UseTranspositionTable {
		e.tt.NewSearch()
	}

	e.wg.Add(1)
	defer e.wg.Done()

	p := e.Pos.Clone()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	e.tree.UpdateRoot(p)
	e.tree.SetSample(e.SampleBestMove)

	for i := 0; e.tree.Len() > 0 && i < 100; i++ {
		n := e.tree.Select()
		n.Expand()
		v := n.Evaluate(e.model)
		n.Backprop(v, 1)
	}

	m, value, node, ok := e.tree.BestMove(r)

	if e.UseTFRecordWriter {
		e.batchWriter.WriteExample(p, e.tree)
		if value.Terminal() {
			e.batchWriter.Finalize(value)
		}
	}

	e.tree.RetainOptimalSubtree(node)

	if !ok {
		e.Logf("no moves")
		return
	}
	e.Logf("info score %f", value)
	if ponder {
		e.Outputf("info pv %s", m)
	} else {
		e.Outputf("bestmove %s", m)
	}

	e.bestMove = m
}
