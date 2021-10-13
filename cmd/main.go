package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gioporta/mapcutter"
	"github.com/gioporta/mapcutter/handlers"
	"github.com/spf13/cobra"
)

var (
	inputFile string
	outputDir string

	rootCmd = &cobra.Command{
		Use:   "mapcutter",
		Short: "Blazing fast map tile cutter written in Go.",
		Long:  `A map cutter developed by Giovanni.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := runCutter(outputDir, inputFile)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&inputFile, "in", "i", "", "input file (required)")
	rootCmd.Flags().StringVarP(&outputDir, "out", "o", "output", "output directory")
	rootCmd.MarkFlagRequired("in")
}

func runCutter(outDir string, inputFile string) error {
	imageFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("read image file error: %w", err)
	}
	defer imageFile.Close()

	tileMap, err := mapcutter.NewMap(imageFile)
	if err != nil {
		return err
	}

	fileHandler := handlers.NewFileHandler(outDir)
	tileMap.CutAllTiles(fileHandler)

	return nil
}
