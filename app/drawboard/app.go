package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten"
)

func main() {
	game := newGame()

	if err := ebiten.RunGame(game); err != nil {
		if err != errExit {
			log.Fatal(err)
		}
	}
}

var errExit = errors.New("exit")

type game struct{}

func newGame() *game {
	return &game{}
}

func (g *game) Update(screen *ebiten.Image) error {
	return nil
}

const (
	gold   = 0
	silver = 1
)

const (
	rabbit   = 0
	cat      = 1
	dog      = 2
	horse    = 3
	camel    = 4
	elephant = 5
)

var (
	boardImage  image.Image
	pieceImages [6][2]image.Image
)

func init() {
	{
		f, err := os.Open(filepath.Join("images", "board.png"))
		if err != nil {
			panic(err)
		}
		im, err := png.Decode(f)
		if err != nil {
			panic(err)
		}
		boardImage = im
		f.Close()
	}

	for i, name := range []string{
		"rabbit",
		"cat",
		"dog",
		"horse",
		"camel",
		"elephant",
	} {
		for j, color := range []string{
			"gold",
			"silver",
		} {
			f, err := os.Open(filepath.Join("images", fmt.Sprintf("%s_%s.png", name, color)))
			if err != nil {
				panic(err)
			}
			im, err := png.Decode(f)
			if err != nil {
				panic(err)
			}
			pieceImages[i][j] = im
			f.Close()
		}
	}
}

func (g *game) Draw(screen *ebiten.Image) {
	drawBoard(screen)

	drawScaledImage(screen, rabbit, gold, 1, 1)
	drawScaledImage(screen, rabbit, gold, 1, 2)
	drawScaledImage(screen, rabbit, gold, 1, 3)
	drawScaledImage(screen, dog, gold, 1, 4)
	drawScaledImage(screen, dog, gold, 1, 5)
	drawScaledImage(screen, rabbit, gold, 1, 6)
	drawScaledImage(screen, rabbit, gold, 1, 7)
	drawScaledImage(screen, rabbit, gold, 1, 8)

	drawScaledImage(screen, cat, gold, 2, 1)
	drawScaledImage(screen, rabbit, gold, 2, 2)
	drawScaledImage(screen, horse, gold, 2, 3)
	drawScaledImage(screen, rabbit, gold, 2, 4)
	drawScaledImage(screen, elephant, gold, 2, 5)
	drawScaledImage(screen, camel, gold, 2, 6)
	drawScaledImage(screen, horse, gold, 2, 7)
	drawScaledImage(screen, cat, gold, 2, 8)

	drawScaledImage(screen, rabbit, silver, 8, 1)
	drawScaledImage(screen, rabbit, silver, 8, 2)
	drawScaledImage(screen, dog, silver, 8, 3)
	drawScaledImage(screen, cat, silver, 8, 4)
	drawScaledImage(screen, cat, silver, 8, 5)
	drawScaledImage(screen, dog, silver, 8, 6)
	drawScaledImage(screen, rabbit, silver, 8, 7)
	drawScaledImage(screen, rabbit, silver, 8, 8)

	drawScaledImage(screen, horse, silver, 7, 1)
	drawScaledImage(screen, rabbit, silver, 7, 2)
	drawScaledImage(screen, rabbit, silver, 7, 3)
	drawScaledImage(screen, camel, silver, 7, 4)
	drawScaledImage(screen, elephant, silver, 7, 5)
	drawScaledImage(screen, rabbit, silver, 7, 6)
	drawScaledImage(screen, rabbit, silver, 7, 7)
	drawScaledImage(screen, horse, silver, 7, 8)
}

func drawBoard(screen *ebiten.Image) {
	im, _ := ebiten.NewImageFromImage(boardImage, ebiten.FilterDefault)
	screen.DrawImage(im, &ebiten.DrawImageOptions{})
}

func drawScaledImage(screen *ebiten.Image, piece, color, rank, file int) {
	im, _ := ebiten.NewImageFromImage(pieceImages[piece][color], ebiten.FilterDefault)
	var geoM ebiten.GeoM
	geoM.Scale(.5, .5)
	geoM.Translate(60*float64(file-1), 60*float64(8-rank))
	screen.DrawImage(im, &ebiten.DrawImageOptions{
		GeoM: geoM,
	})
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 60 * 8, 60 * 8
}
