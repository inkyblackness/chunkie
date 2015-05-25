package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	docopt "github.com/docopt/docopt-go"

	"github.com/inkyblackness/res"
	"github.com/inkyblackness/res/chunk"
	"github.com/inkyblackness/res/chunk/dos"
)

const (
	// Version contains the current version number
	Version = "0.1.0"
	// Name is the name of the application
	Name = "InkyBlackness Chunkie"
	// Title contains a combined string of name and version
	Title = Name + " v." + Version
)

func main() {
	arguments, _ := docopt.Parse(usage(), nil, true, Title, false)
	//fmt.Printf("%v\n", arguments)

	if arguments["export"].(bool) {
		resourceFile := arguments["<resource-file>"].(string)
		inFile, _ := os.Open(resourceFile)
		defer inFile.Close()
		provider, _ := dos.NewChunkProvider(inFile)
		chunkID, _ := strconv.ParseUint(arguments["<chunk-id>"].(string), 0, 16)
		blockID, _ := strconv.ParseUint(arguments["--block"].(string), 0, 16)
		folder := arguments["<folder>"].(string)

		os.MkdirAll(folder, os.FileMode(0755))
		export(provider, folder, res.ResourceID(chunkID), uint16(blockID))
	}
}

func export(provider chunk.Provider, folder string, chunkID res.ResourceID, blockID uint16) {
	holder := provider.Provide(chunkID)
	blockData := holder.BlockData(blockID)

	outFileName := fmt.Sprintf("%04X_%03d.bin", int(chunkID), blockID)
	ioutil.WriteFile(path.Join(folder, outFileName), blockData, os.FileMode(0644))
}

func usage() string {
	return Title + `

Usage:
  chunkie export <resource-file> <chunk-id> [--block=<block-id>] [<folder>]
  chunkie -h | --help
  chunkie --version

Options:
  <resource-file>     The resource file to work on.
  <chunk-id>          The chunk identifier. Defaults to decimal, use "0x" as prefix for hexadecimal.
  --block=<block-id>  The block identifier. Defaults to decimal, use "0x" as prefix for hexadecimal. [default: 0]
  <folder>            The path of the folder to use. [default: "."]
  -h --help           Show this screen.
  --version           Show version.
`
}
