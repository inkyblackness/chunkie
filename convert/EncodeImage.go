package convert

import (
	"bytes"
	goimage "image"

	"github.com/inkyblackness/res/image"
)

// EncodeImage takes a paletted image and encodes it as a block.
func EncodeImage(img *goimage.Paletted, withPrivatePalette bool) []byte {
	palette := img.Palette

	if !withPrivatePalette {
		palette = nil
	}
	bmp := image.ToBitmap(img, palette)
	buf := bytes.NewBuffer(nil)
	image.Write(buf, bmp, image.CompressedBitmap, 0)

	return buf.Bytes()
}
