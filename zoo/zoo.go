package main

import (
	"flag"
	"log"
	"os"
	"time"

	zoo "ajz.dev/games/arimaa-zoo"
)

var (
	seed    = flag.Int64("seed", 0, "random seed passed to the engine")
	verbose = flag.Bool("verbose", false, "log all protocol messages sent and received")
)

func main() {
	flag.Parse()
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	engine := zoo.NewEngine(*seed)
	engine.SetVerbose(*verbose)
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix("")
	log.Println("bot_zoo v0 by Alan Zaffetti")
	log.Println("For operation instructions: <http://arimaa.janzert.com/aei/aei-protocol.html>")
	log.Println(`To quit: type "quit"`)
	if err := engine.RunAEI(os.Stdin); err != nil {
		log.Fatal(err)
	}
	log.Println("Goodbye!")
}
