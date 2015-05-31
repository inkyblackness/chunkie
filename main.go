package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	docopt "github.com/docopt/docopt-go"

	"github.com/inkyblackness/res"
	"github.com/inkyblackness/res/audio"
	"github.com/inkyblackness/res/chunk"
	"github.com/inkyblackness/res/chunk/dos"
	"github.com/inkyblackness/res/movi"
	"github.com/inkyblackness/res/serial"

	"github.com/inkyblackness/chunkie/conv/wav"
)

const (
	// Version contains the current version number
	Version = "0.1.0"
	// Name is the name of the application
	Name = "InkyBlackness Chunkie"
	// Title contains a combined string of name and version
	Title = Name + " v." + Version
)

func usage() string {
	return Title + `

Usage:
  chunkie export <resource-file> <chunk-id> [--block=<block-id>] [--raw] [<folder>]
  chunkie import <resource-file> <chunk-id> [--block=<block-id>] [--data-type=<id>] <source-file>
  chunkie -h | --help
  chunkie --version

Options:
  <resource-file>     The resource file to work on.
  <chunk-id>          The chunk identifier. Defaults to decimal, use "0x" as prefix for hexadecimal.
  --block=<block-id>  The block identifier. Defaults to decimal, use "0x" as prefix for hexadecimal. [default: 0]
	--raw               With this flag, the chunk will be exported without conversion to a common file format.
	--data-type=<id>    The type of the chunk to write.
  <folder>            The path of the folder to use. [default: .]
	<source-file>       The source file to import.
  -h --help           Show this screen.
  --version           Show version.
`
}

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
		raw := arguments["--raw"].(bool)
		folderArgument := arguments["<folder>"]
		folder := "."

		if folderArgument != nil {
			folder = folderArgument.(string)
		}
		os.MkdirAll(folder, os.FileMode(0755))

		holder := provider.Provide(res.ResourceID(chunkID))
		outFileName := fmt.Sprintf("%04X_%03d", int(chunkID), blockID)
		export(holder, uint16(blockID), path.Join(folder, outFileName), raw)
	} else if arguments["import"].(bool) {
		resourceFile := arguments["<resource-file>"].(string)
		chunkID, _ := strconv.ParseUint(arguments["<chunk-id>"].(string), 0, 16)
		blockID, _ := strconv.ParseUint(arguments["--block"].(string), 0, 16)
		sourceFile := arguments["<source-file>"].(string)
		dataType := -1
		dataTypeArgument := arguments["--data-type"]
		if dataTypeArgument != nil {
			result, _ := strconv.ParseUint(dataTypeArgument.(string), 0, 8)
			dataType = int(result)
		}

		importData(resourceFile, res.ResourceID(chunkID), uint16(blockID), dataType, sourceFile)
	}
}

func export(holder chunk.BlockHolder, blockID uint16, outFileName string, raw bool) {
	blockData := holder.BlockData(blockID)
	contentType := holder.ContentType()

	if !raw && contentType == res.Sound {
		soundData, _ := audio.DecodeSoundChunk(blockData)
		wav.ExportToWav(outFileName+".wav", soundData)
	} else if contentType == res.Media {
		soundData, _ := movi.ExtractAudio(blockData)
		wav.ExportToWav(outFileName+".wav", soundData)
	} else {
		ioutil.WriteFile(outFileName+".bin", blockData, os.FileMode(0644))
	}
}

func importData(resourceFile string, chunkID res.ResourceID, blockID uint16, dataType int, sourceFile string) {
	buffer := serial.NewByteStore()
	writer := dos.NewChunkConsumer(buffer)

	{
		inFile, _ := os.Open(resourceFile)
		defer inFile.Close()
		provider, _ := dos.NewChunkProvider(inFile)

		ids := provider.IDs()
		for _, id := range ids {
			sourceChunk := provider.Provide(id)
			blockCount := sourceChunk.BlockCount()
			blocks := make([][]byte, blockCount)
			for block := uint16(0); block < blockCount; block++ {
				if id == chunkID && block == blockID {
					blocks[block] = importFile(sourceFile, sourceChunk.ContentType())
				} else {
					blocks[block] = sourceChunk.BlockData(block)
				}
			}

			destChunk := chunk.NewBlockHolder(sourceChunk.ChunkType(), sourceChunk.ContentType(), blocks)
			writer.Consume(id, destChunk)
		}
	}
	writer.Finish()

	err := ioutil.WriteFile(resourceFile, buffer.Data(), os.FileMode(0644))
	if err != nil {
		panic(err)
	}
}

func importFile(sourceFile string, dataType res.DataTypeID) (data []byte) {
	extension := path.Ext(sourceFile)
	switch extension {
	case ".wav":
		{
			soundData := wav.ImportFromWav(sourceFile)
			if dataType == res.Sound {
				data = audio.EncodeSoundChunk(soundData)
			} else if dataType == res.Media {
				data = movi.ContainSoundData(soundData)
			}
		}
	default:
		{
			data, _ = ioutil.ReadFile(sourceFile)
		}
	}

	return
}
