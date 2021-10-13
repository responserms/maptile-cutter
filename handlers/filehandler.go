package handlers

import (
	"fmt"
	"io"
	"os"

	"github.com/gioporta/mapcutter"
)

type FileHandler struct {
	OutputDir string
}

func NewFileHandler(outDir string) mapcutter.TileHandler {
	h := new(FileHandler)
	h.OutputDir = outDir
	return h
}

func (f FileHandler) HandleTile(r io.Reader, zoom int, x int, y int) {
	filename := fmt.Sprintf("%s/%d_%d_%d.png", f.OutputDir, zoom, x, y)
	imageFile, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer imageFile.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		fmt.Println(err)
	}

	imageFile.Write(b)
}
