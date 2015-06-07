package gif

import (
	"bytes"
	"encoding/binary"
	goimage "image"
	"image/color"
	"image/gif"
	"io"
	"os"

	"github.com/inkyblackness/res/compress/rle"
	"github.com/inkyblackness/res/image"
)

// ExportToGif reads a bitmap from given block data and saves it to given fileName. If the image has no
// private palette, the given will be used. If no palette can be resolved, the return value is false.
func ExportToGif(fileName string, blockData []byte, palette color.Palette) (result bool) {
	reader := bytes.NewReader(blockData)
	var header image.BitmapHeader
	var bitmap []byte
	usedPalette := palette

	binary.Read(reader, binary.LittleEndian, &header)

	if header.Type == 0x04 {
		bitmap = decompress(reader, &header)
	} else {
		bitmap = make([]byte, int(header.Height)*int(header.Width))
		reader.Read(bitmap)
	}

	{
		curPos, _ := reader.Seek(0, 1)
		remain := len(blockData) - int(curPos)

		if remain > 0 {
			paletteFlag := uint32(0)

			binary.Read(reader, binary.LittleEndian, &paletteFlag)
			usedPalette, _ = image.LoadPalette(reader)
		}
	}

	if usedPalette != nil {
		img := goimage.NewPaletted(goimage.Rect(0, 0, int(header.Width), int(header.Height)), usedPalette)
		for row := 0; row < int(header.Height); row++ {
			rowStart := row * int(header.Stride)
			copy(img.Pix[row*img.Stride:], bitmap[rowStart:rowStart+int(header.Width)])
		}

		container := gif.GIF{
			Image:     []*goimage.Paletted{img},
			Delay:     []int{0},
			LoopCount: 0}
		gifFile, _ := os.Create(fileName)
		gif.EncodeAll(gifFile, &container)
		gifFile.Close()
		result = true
	}

	return
}

func decompress(reader io.ReadSeeker, header *image.BitmapHeader) (result []byte) {
	expectedSize := int(header.Height) * int(header.Stride)

	result, _ = rle.Decompress(reader, expectedSize)

	return
}
