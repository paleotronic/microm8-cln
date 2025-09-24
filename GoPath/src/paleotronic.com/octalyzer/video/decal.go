package video

import (
	//"paleotronic.com/fmt"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
)

type DecalPos struct {
	TX0, TY0, TX1, TY1 float32
}

type Decal struct {
	Texture              *glumby.Texture // Pointer to a unique decal texture
	Mesh                 *glumby.Mesh
	Position             glumby.Vector3
	Name                 string
	Width, Height, Depth float32
	BlinkInterval        int32 // ms
	CursorHere           bool
	HSkip, VSkip         int
	SrcPos               DecalPos // texture source co-ords
}

type DecalBatch struct {
	Items []*Decal
}

func NewDecal(w, h float32) *Decal {
	this := &Decal{Width: w, Height: h, Depth: 1}
	return this
}

func NewDecalBatch() *DecalBatch {
	this := &DecalBatch{Items: make([]*Decal, 0)}
	return this
}

// Render the decals, grouped by texture id
func (db *DecalBatch) Render(mbo *glumby.MeshBufferObject) {

	// Sort the decals based on texture for performance :)
	collected := make(map[*glumby.Texture][]*Decal)

	var texcount int = 0

	for _, decal := range db.Items {
		// sort the decals
		tex := decal.Texture
		if tex == nil {
			continue
		}
		list, ok := collected[tex]
		if !ok {
			list = make([]*Decal, 0)
		}

		list = append(list, decal)
		collected[tex] = list
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	for tex, list := range collected {
		tex.Bind()
		texcount++

		mbo.Begin(gl.TRIANGLES)

		for _, decal := range list {
			mbo.Draw(decal.Position.X, decal.Position.Y, decal.Position.Z, decal.Mesh)
		}

		mbo.Send(true)
	}

	//log.Printf("Pushed %d unique textures\n", texcount)
}

func (db *DecalBatch) RenderFromAtlas(mbo *glumby.MeshBufferObject, tex *glumby.Texture) {

	var planeTexcoords []float32 = []float32{
		1, 1,
		0, 1,
		0, 0,
		0, 0,
		1, 0,
		1, 1,
	}

	tex.Bind()

	// Sort the decals based on texture for performance :)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	mbo.Begin(gl.TRIANGLES)
	for _, decal := range db.Items {
		// sort the decals
		if decal.Texture == nil {
			continue
		}

		//fmt.Println("TEXCOORDS =", planeTexcoords, decal.SrcPos)

		planeTexcoords = []float32{
			decal.SrcPos.TX1, decal.SrcPos.TY1,
			decal.SrcPos.TX0, decal.SrcPos.TY1,
			decal.SrcPos.TX0, decal.SrcPos.TY0,
			decal.SrcPos.TX0, decal.SrcPos.TY0,
			decal.SrcPos.TX1, decal.SrcPos.TY0,
			decal.SrcPos.TX1, decal.SrcPos.TY1,
		}

		mbo.DrawWithTexCoords(decal.Position.X, decal.Position.Y, decal.Position.Z, decal.Mesh, planeTexcoords)
	}
	mbo.Send(true)

	//log.Printf("Pushed %d unique textures\n", texcount)
}
