package mapcutter

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"sync"

	"github.com/oliamb/cutter"
	"golang.org/x/image/draw"
)

type tileMap struct {
	images    []image.Image
	SourceRes int
	MaxZoom   int
}

type TileCutter interface {
	CutTile(w io.Writer, zoomLevel int, posX int, posY int) error
	CutAllTiles(TileHandler) error
}

type TileHandler interface {
	HandleTile(r io.Reader, zoom int, x int, y int)
}

type HandleTile func(r io.Reader, zoom int, x int, y int)

func (f HandleTile) HandleTile(r io.Reader, zoom int, x int, y int) {
	f(r, zoom, x, y)
}

func NewMap(sourceFile io.Reader) (TileCutter, error) {
	tileMap := new(tileMap)

	img, imgRes, err := readImage(sourceFile)
	if err != nil {
		return nil, err
	}

	tileMap.SourceRes = imgRes

	tileMap.MaxZoom = calcZoomLevels(float64(tileMap.SourceRes), 256)

	tileMap.images = make([]image.Image, tileMap.MaxZoom+1)

	for zoomLevel := tileMap.MaxZoom; zoomLevel >= 0; zoomLevel-- {
		resizeRes := 256 * int(math.Pow(2, float64(zoomLevel)))

		if resizeRes != tileMap.SourceRes {
			img = resizeImage(img, resizeRes)
			fmt.Println("Image resized to", resizeRes)
		}

		tileMap.images[zoomLevel] = img
	}

	return tileMap, nil
}

func (tileMap *tileMap) CutAllTiles(h TileHandler) error {
	var wg sync.WaitGroup

	for zoomLevel := tileMap.MaxZoom; zoomLevel >= 0; zoomLevel-- {
		wg.Add(1)
		zoomLevel := zoomLevel
		imageToCut := tileMap.images[zoomLevel]

		go func() {
			defer wg.Done()
			err := cutTilesByZoomLevel(imageToCut, zoomLevel, h)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}

	wg.Wait()
	return nil
}

func (tileMap *tileMap) CutTile(w io.Writer, zoomLevel int, posX int, posY int) error {
	img, err := cutTile(tileMap.images[zoomLevel], zoomLevel, posX, posY)
	if err != nil {
		return err
	}

	err = png.Encode(w, img)
	if err != nil {
		return err
	}

	return nil
}

func readImage(r io.Reader) (image.Image, int, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, 0, fmt.Errorf("image decode error: %w", err)
	}

	imgRes := img.Bounds().Max.X

	return img, imgRes, err
}

func cutTilesByZoomLevel(img image.Image, zoomLevel int, h TileHandler) error {
	numTiles := 0
	maxPos := int(math.Pow(2, float64(zoomLevel)) - 1)

	fmt.Println("Beginning zoom", zoomLevel)

	for x := 0; x <= maxPos; x++ {
		for y := 0; y <= maxPos; y++ {
			cutImg, err := cutTile(img, zoomLevel, x, y)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			err = png.Encode(&buf, cutImg)
			if err != nil {
				return err
			}

			h.HandleTile(&buf, zoomLevel, x, y)
			numTiles++
		}
	}
	fmt.Println(numTiles, "tiles cut for zoom", zoomLevel)
	return nil
}

func cutTile(img image.Image, zoomLevel int, x int, y int) (image.Image, error) {
	cutRes := 256
	img, err := cutter.Crop(img, cutter.Config{
		Width:  cutRes,
		Height: cutRes,
		Anchor: image.Point{cutRes * x, cutRes * y},
	})
	if err != nil {
		return nil, err
	}

	return img, nil
}

func resizeImage(img image.Image, width int) image.Image {
	rect := image.Rect(0, 0, width, width)
	scale := draw.BiLinear
	dst := image.NewRGBA(rect)
	scale.Scale(dst, rect, img, img.Bounds(), draw.Over, nil)
	return dst
}

func calcZoomLevels(srcRes float64, dstRes float64) int {
	return int(math.Log(srcRes/dstRes) / math.Log(2))
}
