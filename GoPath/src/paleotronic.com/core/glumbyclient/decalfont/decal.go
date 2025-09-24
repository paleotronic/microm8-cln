package decalfont

import (
	"github.com/go-gl/gl/v2.1/gl"
	//	"paleotronic.com/log"
	"paleotronic.com/glumby"
)

type Decal struct {
	Texture              *glumby.Texture // Pointer to a unique decal texture
	Mesh                 *glumby.Mesh
	Position             glumby.Vector3
	Name                 string
	Width, Height, Depth float32
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
func (db *DecalBatch) Render() {

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

	for tex, list := range collected {
		// bind the texture
		tex.Bind()
		texcount++

		// iterate over the meshes pushing them out to the buffer
		//gl.PushMatrix()
		//gl.Scalef(12.8, 16, 1)
		glumby.MeshBuffer_Begin(gl.QUADS)
		for _, decal := range list {
			decal.Mesh.DrawWithMeshBuffer(decal.Position.X, decal.Position.Y, decal.Position.Z)
		}
		glumby.MeshBuffer_End()
		//gl.PopMatrix()
		// unbind the texture
		//tex.Unbind()
	}

	//log.Printf("Pushed %d unique textures\n", texcount)
}
