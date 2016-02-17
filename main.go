package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	goImage "image"
	"image/color"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	docopt "github.com/docopt/docopt-go"

	"github.com/inkyblackness/res"
	"github.com/inkyblackness/res/audio"
	"github.com/inkyblackness/res/chunk"
	"github.com/inkyblackness/res/chunk/dos"
	"github.com/inkyblackness/res/compress/rle"
	"github.com/inkyblackness/res/data"
	"github.com/inkyblackness/res/image"
	"github.com/inkyblackness/res/movi"
	"github.com/inkyblackness/res/serial"

	"github.com/inkyblackness/chunkie/convert"
	"github.com/inkyblackness/chunkie/convert/wav"
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
  chunkie export <resource-file> <chunk-id> [--block=<block-id>] [--raw] [--pal=<palette-file>] [--fps=<framerate>] [<folder>]
  chunkie import <resource-file> <chunk-id> [--block=<block-id>] [--data-type=<id>] <source-file>
  chunkie -h | --help
  chunkie --version

Options:
  <resource-file>       The resource file to work on.
  <chunk-id>            The chunk identifier. Defaults to decimal, use "0x" as prefix for hexadecimal.
  --block=<block-id>    The block identifier. Defaults to decimal, use "0x" as prefix for hexadecimal. [default: 0]
  --raw                 With this flag, the chunk will be exported without conversion to a common file format.
  --pal=<palette-file>  For handling bitmaps & models, use this palette file to write color information
  --fps=<framerate>     The frames per second to emulate when exporting movies. 0 names files after timestamp. [default: 0]
  --data-type=<id>      The type of the chunk to write.
  <folder>              The path of the folder to use. [default: .]
  <source-file>         The source file to import.
  -h --help             Show this screen.
  --version             Show version.
`
}

func main() {
	arguments, _ := docopt.Parse(usage(), nil, true, Title, false)
	fmt.Printf("%v\n", arguments)

	if arguments["export"].(bool) {
		resourceFile := arguments["<resource-file>"].(string)
		inFile, _ := os.Open(resourceFile)
		defer inFile.Close()
		provider, _ := dos.NewChunkProvider(inFile)
		chunkID, _ := strconv.ParseUint(arguments["<chunk-id>"].(string), 0, 16)
		blockID, _ := strconv.ParseUint(arguments["--block"].(string), 0, 16)
		framesPerSecond, _ := strconv.ParseFloat(arguments["--fps"].(string), 32)
		raw := arguments["--raw"].(bool)
		palArgument := arguments["--pal"]
		var paletteFile string
		folderArgument := arguments["<folder>"]
		folder := "."

		if palArgument != nil {
			paletteFile = palArgument.(string)
		}
		if folderArgument != nil {
			folder = folderArgument.(string)
		}
		os.MkdirAll(folder, os.FileMode(0755))

		holder := provider.Provide(res.ResourceID(chunkID))
		outFileName := fmt.Sprintf("%04X_%03d", int(chunkID), blockID)
		exportFile(provider, holder, uint16(blockID), path.Join(folder, outFileName), raw, paletteFile, float32(framesPerSecond))
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

func exportFile(provider chunk.Provider, holder chunk.BlockHolder, blockID uint16,
	outFileName string, raw bool, paletteFile string, framesPerSecond float32) {
	blockData := holder.BlockData(blockID)
	contentType := holder.ContentType()
	exportRaw := raw

	if !exportRaw {
		if contentType == res.Sound {
			soundData, _ := audio.DecodeSoundChunk(blockData)
			wav.ExportToWav(outFileName+".wav", soundData)
		} else if contentType == res.Media {
			exportRaw = exportMedia(blockData, outFileName, framesPerSecond)
		} else if contentType == res.Bitmap {
			palette := loadPalette(paletteFile)
			exportRaw = !convert.ToPng(outFileName+".png", blockData, palette)
		} else if contentType == res.Geometry {
			palette := loadPalette(paletteFile)
			exportRaw = !convert.ToWavefrontObj(outFileName, blockData, palette)
		} else if contentType == res.VideoClip {
			palette := loadPalette(paletteFile)
			exportRaw = exportVideoClip(provider, blockData, outFileName, framesPerSecond, palette)
		} else {
			exportRaw = true
		}
	}
	if exportRaw {
		ioutil.WriteFile(outFileName+".bin", blockData, os.FileMode(0644))
	}
}

func loadPalette(fileName string) (pal color.Palette) {
	if len(fileName) > 0 {
		inFile, _ := os.Open(fileName)
		defer inFile.Close()
		provider, _ := dos.NewChunkProvider(inFile)

		ids := provider.IDs()
		for _, id := range ids {
			blockHolder := provider.Provide(id)

			if blockHolder.ContentType() == res.Palette && pal == nil {
				pal, _ = image.LoadPalette(bytes.NewReader(blockHolder.BlockData(0)))
			}
		}
	}
	return
}

func exportMedia(blockData []byte, fileBaseName string, framesPerSecond float32) (failed bool) {
	container, err := movi.Read(bytes.NewReader(blockData))

	if err == nil {
		handler := newExportingMediaHandler(fileBaseName, container.MediaDuration(), framesPerSecond, float32(container.AudioSampleRate()))
		dispatcher := movi.NewMediaDispatcher(container, handler)
		more := true

		for more && err == nil {
			more, err = dispatcher.DispatchNext()
		}
		if !more {
			handler.finish()
		}
	}

	if err != nil {
		failed = true
	}
	return
}

func exportVideoClip(provider chunk.Provider, blockData []byte, fileBaseName string, framesPerSecond float32, pal color.Palette) (failed bool) {
	reader := bytes.NewReader(blockData)
	sequence := data.DefaultVideoClipSequence((len(blockData) - data.VideoClipSequenceBaseSize) / data.VideoClipSequenceEntrySize)
	var err error

	serial.MapData(sequence, serial.NewDecoder(reader))
	{
		times := make([]float32, 0)
		mediaDuration := float32(0.0)
		for _, entry := range sequence.Entries {
			frameTime := float32(entry.Unknown0003[1]) + float32(entry.Unknown0003[0])/256.0
			for i := 0; i < int(entry.LastFrame-entry.FirstFrame)+1; i++ {
				times = append(times, mediaDuration)
				mediaDuration += frameTime
			}
		}

		framesData := provider.Provide(sequence.FramesID)

		imageRect := goImage.Rect(0, 0, int(sequence.Width), int(sequence.Height))
		img := goImage.NewPaletted(imageRect, pal)
		handler := newExportingMediaHandler(fileBaseName, mediaDuration, framesPerSecond, 0.0)
		for frameId := uint16(0); frameId < framesData.BlockCount() && err == nil; frameId++ {
			frameReader := bytes.NewReader(framesData.BlockData(frameId))
			var header image.BitmapHeader

			binary.Read(frameReader, binary.LittleEndian, &header)
			err = rle.Decompress(frameReader, img.Pix)
			handler.OnVideo(times[int(frameId)], img)
		}
		handler.finish()
	}

	if err != nil {
		fmt.Printf("error exporting video clip: %v\n", err)
		failed = true
	}
	return
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
	case ".png":
		{
			if dataType == res.Bitmap {
				data = convert.FromPng(sourceFile, false)
			}
		}
	default:
		{
			data, _ = ioutil.ReadFile(sourceFile)
		}
	}

	return
}
