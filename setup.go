package zoo

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

const randomSetupAttempts = 100

var setupValue = [][]Value{{}, { // Rabbit
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, -50, -50, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // Cat
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 5, 5, 0, 0, 0,
}, { // Dog
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	-20, 0, 0, 0, 0, 0, 0, -20,
}, { // Horse
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	-50, 0, 0, 0, 0, 0, 0, -50,
}, { // Camel
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -50, 50, 50, -50, 0, 0,
	-50, -50, -50, -50, -50, -50, -50, -50,
}, { // Elephant
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -50, 50, 50, -50, 0, 0,
	-50, -50, -50, -50, -50, -50, -50, -50,
}}

func (p *Pos) setupValue(side Color) (value Value) {
	c := 7
	m := -1
	if side != Gold {
		c = 0
		m = 1
	}
	for _, t := range []Piece{
		GRabbit,
		GCat,
		GDog,
		GHorse,
		GCamel,
		GElephant,
	} {
		ps := setupValue[t]
		for b := p.bitboards[t]; b > 0; b &= b - 1 {
			at := b.Square()
			value += ps[8*(c+m*int(at)/8)+c+m*(int(at)%8)]
		}
	}
	return value
}

// RandomSetup creates setup moves by trying positions randomly and evaluating results.
// It repeats this a few times and returns the best setup. This keeps setups generally
// kosher while allowing for the rare "fun" setup.
func (e *Engine) RandomSetup() []Step {
	p := e.Pos()
	side := p.Side()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var best []Step
	bestScore := -Inf
	for i := 0; i < randomSetupAttempts; i++ {
		move := e.randomSetup(r)
		if err := p.Move(move); err != nil {
			log.Println(fmt.Errorf("random_setup_move: %v", err))
			ppanic(p, fmt.Errorf("random_setup_move: %v", err))
		}
		if score := p.positionScore(side); score > bestScore {
			best = move
			bestScore = score
		}
		if err := p.Unmove(); err != nil {
			log.Println(fmt.Errorf("random_setup_unmove: %v", err))

			ppanic(p, fmt.Errorf("random_setup_unmove: %v", err))
		}
	}
	return best
}

func (e *Engine) randomSetup(r *rand.Rand) []Step {
	p := e.Pos()
	c := p.side
	rank := 7
	if c == Gold {
		rank = 1
	}
	ps := []Piece{
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GCat.MakeColor(c),
		GCat.MakeColor(c),
		GDog.MakeColor(c),
		GDog.MakeColor(c),
		GHorse.MakeColor(c),
		GHorse.MakeColor(c),
		GCamel.MakeColor(c),
		GElephant.MakeColor(c),
	}
	r.Shuffle(len(ps), func(i, j int) {
		ps[i], ps[j] = ps[j], ps[i]
	})
	var setup []Step
	for i := rank; i >= rank-1; i-- {
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			setup = append(setup, MakeSetup(ps[0], at))
			ps = ps[1:]
		}
	}
	setup = append(setup, Pass)
	return setup
}
