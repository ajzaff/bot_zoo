package zoo

import (
	"math/rand"
	"sync"
	"time"
)

type searchState struct {
	tree *Tree
	tt   *TranspositionTable

	wg         sync.WaitGroup
	resultChan chan Move

	// semi-atomic
	stopping int32
	running  int32
}

func (s *searchState) Reset() {
	if s.tt == nil {
		s.tt = &TranspositionTable{}
	}
	s.tt.Resize(50)
	if s.tree == nil {
		s.tree = NewEmptyTree(s.tt)
	}
	s.wg = sync.WaitGroup{}
	s.resultChan = make(chan Move)
	s.stopping = 0
	s.running = 0
}

func (e *Engine) searchRoot(ponder bool) {
	defer e.Stop()

	if e.UseTranspositionTable {
		e.tt.NewSearch()
	}

	p := e.Pos.Clone()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	e.tree.UpdateRoot(p)

	playouts, _ := e.LookupOption("playouts")
	if playouts == 0 {
		playouts = 1
	}

	for i := 0; e.tree.Len() > 0 && i < 1000; i++ {
		n := e.tree.Select()
		n.Expand()
		v := n.Simulate(r, playouts.(int))
		n.Backprop(v, playouts.(int))
	}

	m, value, ok := e.tree.RetainBestMove(r)
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
}
