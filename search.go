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
	Close() error
}

// BatchWriterInterface defines an interface for writing Dataset batch data.
type BatchWriterInterface interface {
	WriteExample(p *Pos, t *Tree)
	Finalize(*Pos, Value) error
	Flush() error
}

func (s *searchState) Reset() error {
	if s.tt == nil {
		s.tt = &TranspositionTable{}
	}
	s.tt.Resize(50)
	s.tree = NewEmptyTree(s.tt)
	if s.model == nil {
		model, err := NewModel()
		if err != nil {
			return err
		}
		s.model = model
	}
	s.wg = sync.WaitGroup{}
	s.stopping = 0
	s.running = 0
	return nil
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

	e.tree.Reset()
	e.tree.UpdateRoot(p, e.model)
	e.tree.SetSample(e.UseSampledMove)

	for i := 0; i < 1600; i++ {
		n, p := e.tree.Select(p)
		n.Expand(p, e.model)
	}

	m, value, _, ok := e.tree.BestMove(r)

	if e.UseDatasetWriter {
		if p.MoveNum() == 1 && p.Side() == Gold {
			e.batchWriter.WriteExample(p, e.tree)
		}
		for _, s := range m {
			p.Step(s)
			e.batchWriter.WriteExample(p, e.tree)
		}
		for range m {
			p.Unstep()
		}
	}

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
