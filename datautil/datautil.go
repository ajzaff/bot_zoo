package datautil

import (
	"bytes"
	"os"
	"path/filepath"

	zoo "github.com/ajzaff/bot_zoo"
)

const bufferSize = 10 * 1024 * 1024

type BatchWriter struct {
	buf bytes.Buffer
	dir string
	f   *os.File
}

func NewBatchWriter() *BatchWriter {
	return &BatchWriter{
		dir: filepath.Join("data", "training"),
	}
}

func (w *BatchWriter) WriteGame(pgn zoo.MoveList, result zoo.Value) {
}

type BatchReader struct {
}
