package main

import (
	"flag"
	"log"
	"os"
	"time"

	zoo "github.com/ajzaff/bot_zoo"
)

var (
	seed     = flag.Int64("seed", 0, "Random seed passed to the engine.")
	verbose  = flag.Bool("verbose", false, "Log all protocol messages sent and received")
	movelist = flag.String("movelist", "",
		`Execute "newgame" followed by the given movelist.
This should be the path to a move list file.  Each line should be prefixed by move number and color (e.g. 5s Rd4e).
Setup moves must be included. The last line may include a move number and color to indice the side to move.
The file may be newline terminated.`)
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
	if *movelist != "" {
		f, err := os.Open(*movelist)
		if err != nil {
			log.Printf("movelist: %v", err)
			return
		}
		if err := engine.NewGameFromMoveList(f); err != nil {
			log.Fatalf("movelist: %v", err)
		}
		f.Close()
	}
	if err := engine.RunAEI(os.Stdin); err != nil {
		log.Fatal(err)
	}
	log.Println("Goodbye!")
}
