package main

import (
	"fmt"
	"log"
	"os"

	zoo "ajz.dev/games/arimaa-zoo"
)

func main() {
	engine := zoo.NewEngine()
	aei := zoo.NewAEI(engine)
	fmt.Println("zoo v0 by Alan Zaffetti")
	fmt.Println("For operation instructions: <http://arimaa.janzert.com/aei/aei-protocol.html>")
	fmt.Println(`To quit: type "quit"`)
	if err := aei.Run(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Goodbye!")
}
