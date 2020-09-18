package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	zoo "github.com/ajzaff/bot_zoo"
)

var (
	seed     = flag.Int64("seed", 0, "Random seed passed to the engine.")
	movelist = flag.String("movelist", "",
		`Execute "newgame" followed by the given movelist.
This should be the path to a move list file.  Each line should be prefixed by move number and color (e.g. 5s Rd4e).
Setup moves must be included. The last line may include a move number and color to indice the side to move.
The file may be newline terminated.`)

	aeiSettings = zoo.RegisterAEIFlags(flag.CommandLine)
)

func main() {
	flag.Parse()
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	engine := zoo.NewEngine(*seed)
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix("")
	log.Println("bot_alpha_zoo by Alan Zaffetti")
	log.Println("For operation instructions: <http://arimaa.janzert.com/aei/aei-protocol.html>")
	log.Println(`To quit: type "quit"`)
	if *movelist != "" {
		bs, err := ioutil.ReadFile(*movelist)
		if err != nil {
			log.Printf("movelist: %v", err)
			return
		}
		_, err = zoo.ParseMoveList(string(bs))
		if err != nil {
			log.Fatalf("movelist: %v", err)
		}
		// TODO(ajzaff): Initialize game state from movelist.
	}
	// Execute AEI loop:
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		if err := engine.ExecuteCommand(text); err != nil {
			if zoo.IsQuit(err) {
				break
			}
			log.Println(err)
		}
	}
	log.Println("Goodbye!")
}
