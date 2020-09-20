package zoo

import "log"

func assert(message string, cond bool) {
	if !cond {
		panic(message)
	}
}

func passert(p *Pos, message string, cond bool) {
	if !cond {
		ppanic(p, message)
	}
}

func ppanic(p *Pos, v interface{}) {
	log.Println(p.String())
	log.Println(p.ShortString())
	log.Println(p.moves.String())
	panic(v)
}
