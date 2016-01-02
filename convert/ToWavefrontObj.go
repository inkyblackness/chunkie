package convert

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"os"
	"path"

	"github.com/inkyblackness/res/geometry"
	"github.com/inkyblackness/res/geometry/command"
)

type wavefrontWriter struct {
	objFile io.Writer
	mtlFile io.Writer

	palette color.Palette

	vtCounter int

	usedMaterials map[string]bool
	lastMaterial  string
}

func (writer *wavefrontWriter) Nodes(anchor geometry.NodeAnchor) {
	anchor.Left().WalkAnchors(writer)
	anchor.Right().WalkAnchors(writer)
}

func (writer *wavefrontWriter) Faces(anchor geometry.FaceAnchor) {
	// fmt.Printf("Face normal: %f %f %f\n", anchor.Normal().X(), anchor.Normal().Y(), anchor.Normal().Z())
	// fmt.Printf("Face refere: %f %f %f\n", anchor.Reference().X(), anchor.Reference().Y(), anchor.Reference().Z())
	anchor.WalkFaces(writer)
}

func (writer *wavefrontWriter) defineMaterial(name string) {
	writer.usedMaterials[name] = true
	fmt.Fprintf(writer.mtlFile, "newmtl %s\n", name)
}

func (writer *wavefrontWriter) defineMaterialColor(color geometry.ColorIndex) {
	r, g, b, _ := writer.palette[int(color)].RGBA()
	fmt.Fprintf(writer.mtlFile, "Ka %f %f %f\n", float32(r)/256.0, float32(g)/256.0, float32(b)/256.0)
	fmt.Fprintf(writer.mtlFile, "Kd %f %f %f\n", float32(r)/256.0, float32(g)/256.0, float32(b)/256.0)
}

func (writer *wavefrontWriter) useMaterial(name string) {
	if writer.lastMaterial != name {
		writer.lastMaterial = name
		fmt.Fprintf(writer.objFile, "usemtl %s\n", name)
	}
}

func (writer *wavefrontWriter) useFlatColorMaterial(color geometry.ColorIndex) {
	name := fmt.Sprintf("mat_col_%02X", int(color))

	if !writer.usedMaterials[name] {
		writer.defineMaterial(name)
		writer.defineMaterialColor(color)
	}
	writer.useMaterial(name)
}

func (writer *wavefrontWriter) useShadeColorMaterial(color geometry.ColorIndex, shade uint16) {
	name := fmt.Sprintf("mat_col_%02X_shade%d", int(color), shade)

	if !writer.usedMaterials[name] {
		writer.defineMaterial(name)
		writer.defineMaterialColor(color)
		fmt.Fprintf(writer.mtlFile, "d %f\n", float32(shade)/3.0)
	}
	writer.useMaterial(name)
}

func (writer *wavefrontWriter) useTextureMaterial(textureId uint16) {
	name := fmt.Sprintf("mat_tex_%04X", textureId)

	if !writer.usedMaterials[name] {
		writer.defineMaterial(name)
		fmt.Fprintf(writer.mtlFile, "map_Kd %04X_000.png\n", 0x01DB+textureId)
	}
	writer.useMaterial(name)
}

func (writer *wavefrontWriter) writeSimpleFaces(vertices []int) {
	fmt.Fprintf(writer.objFile, "f")
	for _, vertexIndex := range vertices {
		fmt.Fprintf(writer.objFile, " %d", vertexIndex+1)
	}
	fmt.Fprintf(writer.objFile, "\n")
}

func (writer *wavefrontWriter) FlatColored(face geometry.FlatColoredFace) {
	writer.useFlatColorMaterial(face.Color())
	writer.writeSimpleFaces(face.Vertices())
}

func (writer *wavefrontWriter) ShadeColored(face geometry.ShadeColoredFace) {
	writer.useShadeColorMaterial(face.Color(), face.Shade())
	writer.writeSimpleFaces(face.Vertices())
}

func (writer *wavefrontWriter) TextureMapped(face geometry.TextureMappedFace) {
	writer.useTextureMaterial(face.TextureID())

	for _, coord := range face.TextureCoordinates() {
		fmt.Fprintf(writer.objFile, "vt %f %f\n", coord.U(), coord.V())
	}

	fmt.Fprintf(writer.objFile, "f")
	for index, vertexIndex := range face.Vertices() {
		fmt.Fprintf(writer.objFile, " %d/%d", vertexIndex+1, writer.vtCounter+index+1)
	}
	fmt.Fprintf(writer.objFile, "\n")
	writer.vtCounter += len(face.TextureCoordinates())
}

// ToWavefrontObj extracts a geometry model from given block data and saves
// the 3D model as a Wavefront OBJ file
func ToWavefrontObj(fileName string, blockData []byte, palette color.Palette) (result bool) {
	model, err := command.LoadModel(bytes.NewReader(blockData))

	if err == nil {
		objFile, _ := os.Create(fileName + ".obj")
		mtlFile, _ := os.Create(fileName + ".mtl")

		if objFile != nil {
			defer objFile.Close()
		}
		if mtlFile != nil {
			defer mtlFile.Close()
		}
		if objFile != nil && mtlFile != nil {
			fmt.Fprintf(objFile, "mtllib %s\n", path.Base(fileName+".mtl"))
			vertexCount := model.VertexCount()
			for i := 0; i < vertexCount; i++ {
				position := model.Vertex(i).Position()
				fmt.Fprintf(objFile, "v %f %f %f\n", position.X(), -position.Y(), position.Z())
			}

			writer := &wavefrontWriter{
				objFile:       objFile,
				mtlFile:       mtlFile,
				palette:       palette,
				usedMaterials: make(map[string]bool)}
			model.WalkAnchors(writer)

			result = true
		}
	}

	return
}
