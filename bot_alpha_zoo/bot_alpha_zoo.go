package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	zoo "github.com/ajzaff/bot_zoo"
)

var (
	engineSettings = zoo.RegisterEngineFlags(flag.CommandLine)
	aeiSettings    = zoo.RegisterAEIFlags(flag.CommandLine)
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix("")

	flag.Parse()
	engine, err := zoo.NewEngine(engineSettings, aeiSettings)
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	log.Println("bot_alpha_zoo by Alan Zaffetti")
	log.Println("For operation instructions: <https://github.com/ajzaff/bot_zoo>")
	log.Println(`To quit: type "quit"`)

	if engineSettings.MoveList != "" {
		bs, err := ioutil.ReadFile(engineSettings.MoveList)
		if err != nil {
			log.Fatalf("Failed to load movelist file: %v", err)
		}
		l, err := zoo.ParseMoveList(string(bs))
		if err != nil {
			log.Fatalf("Failed to parse movelist: %v", err)
		}
		for _, m := range l {
			engine.Move(m)
		}
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
