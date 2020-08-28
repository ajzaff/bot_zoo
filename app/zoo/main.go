package main

import (
	"log"
	"os"

	zoo "ajz.dev/games/arimaa-zoo"
)

func main() {
	engine := zoo.NewEngine()
	aei := zoo.NewAEI(engine)
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix("")
	log.Println("zoo v0 by Alan Zaffetti")
	log.Println("For operation instructions: <http://arimaa.janzert.com/aei/aei-protocol.html>")
	log.Println(`To quit: type "quit"`)
	if err := aei.Run(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
	log.Println("Goodbye!")
}
