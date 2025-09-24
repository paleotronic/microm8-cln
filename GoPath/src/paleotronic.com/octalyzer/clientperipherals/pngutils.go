package clientperipherals

import (
	"bytes"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
	"paleotronic.com/octalyzer/video"
)

func GetDecalMesh(w, h, d float32) *glumby.Mesh {

	m := glumby.NewMesh(gl.QUADS)

	m.Normal3f(0, 0, 1)
	m.TexCoord2f(0, 1)
	m.Vertex3f(-1*w, -1*h, 0)

	m.Normal3f(0, 0, 1)
	m.TexCoord2f(1, 1)
	m.Vertex3f(1*w, -1*h, 0)

	m.Normal3f(0, 0, 1)
	m.TexCoord2f(1, 0)
	m.Vertex3f(1*w, 1*h, 0)

	m.Normal3f(0, 0, 1)
	m.TexCoord2f(0, 0)
	m.Vertex3f(-1*w, 1*h, 0)

	return m
}

// CreateFullscreenSplash() creates a Decal dimensioned to fill the screen
func CreateDecal(pngdata []byte, w, h float32) *video.Decal {
	var b bytes.Buffer

	_, _ = b.Write(pngdata)

	tx, _ := glumby.NewTextureFromBytes(&b)

	decal := &video.Decal{}
	decal.Name = "splash"
	decal.Position = glumby.Vector3{X: w / 2, Y: h / 2, Z: 0.999}
	decal.Texture = tx
	decal.Mesh = GetDecalMesh(w/2, h/2, 0)

	return decal
}
