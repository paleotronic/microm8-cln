package glumby

/*
This is a mesh object.
*/

import (
	"unsafe"

	"paleotronic.com/gl"
	//"paleotronic.com/fmt"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	MAX_MESH_VERTICES = 131072
)

var (
	iv_arr                  [MAX_MESH_VERTICES * 3]int32
	v_arr                   [MAX_MESH_VERTICES * 3]float32
	c_arr                   [MAX_MESH_VERTICES * 4]float32
	n_arr                   [MAX_MESH_VERTICES * 3]float32
	t_arr                   [MAX_MESH_VERTICES * 2]float32
	lastMeshCopy            *Mesh
	MeshBufferVertexCount   int32 // always between 0 and MAX_MESH_VERTICES - 1
	MeshBufferNormalCount   uint32
	MeshBufferColorCount    uint32
	MeshBufferTexCoordCount uint32
	MeshBufferObjectType    uint32
	IntegerVertices         bool
)

type Mesh struct {
	thing          uint32
	vertexCount    int32
	vboIdVertices  uint32
	vboIdTexcoords uint32
	vboIdNormals   uint32
	ivertices      []int32
	vertices       []float32
	verticesT      []float32
	iverticesT     []int32
	texcoords      []float32
	normals        []float32
	colors         []float32
	Texture        *Texture
	hasTexCoords   bool
	hasNormals     bool
	hasColors      bool
	hasIntVerts    bool
}

// Init the mesh buffer
func MeshBuffer_Begin(thing uint32) {
	MeshBufferObjectType = thing
	MeshBufferVertexCount = 0
	MeshBufferColorCount = 0
	MeshBufferNormalCount = 0
	MeshBufferTexCoordCount = 0
}

func MeshBuffer_Flush() {
	if MeshBufferVertexCount > 0 {
		//log.Printf("Flushing %d vertices from buffer\n", MeshBufferVertexCount)
		// enable arrays
		gl.EnableClientState(gl.VERTEX_ARRAY)

		if MeshBufferNormalCount > 0 {
			gl.EnableClientState(gl.NORMAL_ARRAY)
		}

		if MeshBufferTexCoordCount > 0 {
			gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		}

		if MeshBufferColorCount > 0 {
			gl.EnableClientState(gl.COLOR_ARRAY)
		}

		// do the send here
		if !IntegerVertices {
			gl.VertexPointer(3, gl.FLOAT, 0, unsafe.Pointer(&v_arr[0]))
		} else {
			gl.VertexPointer(3, gl.INT, 0, unsafe.Pointer(&iv_arr[0]))
		}
		if MeshBufferTexCoordCount > 0 {
			gl.TexCoordPointer(2, gl.FLOAT, 0, unsafe.Pointer(&t_arr[0]))
		}
		if MeshBufferNormalCount > 0 {
			gl.NormalPointer(gl.FLOAT, 0, unsafe.Pointer(&n_arr[0]))
		}
		if MeshBufferColorCount > 0 {
			gl.ColorPointer(4, gl.FLOAT, 0, unsafe.Pointer(&c_arr[0]))
		}

		gl.Enable(gl.COLOR_MATERIAL)
		gl.DrawArrays(MeshBufferObjectType, 0, MeshBufferVertexCount)

		// disable arrays
		gl.DisableClientState(gl.VERTEX_ARRAY)

		if MeshBufferNormalCount > 0 {
			gl.DisableClientState(gl.NORMAL_ARRAY)
		}

		if MeshBufferTexCoordCount > 0 {
			gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
		}

		if MeshBufferColorCount > 0 {
			gl.DisableClientState(gl.COLOR_ARRAY)
		}
	}
	MeshBufferVertexCount = 0
	MeshBufferColorCount = 0
	MeshBufferNormalCount = 0
	MeshBufferTexCoordCount = 0
}

func MeshBuffer_End() {
	MeshBuffer_Flush()
}

func NewMesh(thing uint32) *Mesh {
	this := &Mesh{}

	this.thing = thing
	this.normals = make([]float32, 0)
	this.vertices = make([]float32, 0)
	this.texcoords = make([]float32, 0)
	this.verticesT = make([]float32, 0)
	this.ivertices = make([]int32, 0)
	this.iverticesT = make([]int32, 0)

	return this
}

// Add a vertex
func (m *Mesh) Vertex3i(x, y, z int32) {
	m.ivertices = append(m.ivertices, x, y, z)
	m.iverticesT = append(m.iverticesT, x, y, z)
	m.vertexCount += 1
	m.hasIntVerts = true
}

func (m *Mesh) LinePair3v(a, b, c mgl32.Vec3) {
	m.Vertex3f(a[0], a[1], a[2])
	m.Vertex3f(b[0], b[1], b[2])
	m.Vertex3f(b[0], b[1], b[2])
	m.Vertex3f(c[0], c[1], c[2])
	m.Color4f(1, 1, 1, 1)
	m.Color4f(1, 1, 1, 1)
	m.Color4f(1, 1, 1, 1)
	m.Color4f(1, 1, 1, 1)
}

func (m *Mesh) Triangle3v(a, b, c mgl32.Vec3, n mgl32.Vec3) {
	m.Vertex3f(a[0], a[1], a[2])
	m.Vertex3f(b[0], b[1], b[2])
	m.Vertex3f(c[0], c[1], c[2])
	m.Normal3f(n[0], n[1], n[2])
	m.Normal3f(n[0], n[1], n[2])
	m.Normal3f(n[0], n[1], n[2])
	m.Color4f(1, 1, 1, 1)
	m.Color4f(1, 1, 1, 1)
	m.Color4f(1, 1, 1, 1)
}

// Add a vertex
func (m *Mesh) Vertex3f(x, y, z float32) {
	m.vertices = append(m.vertices, x, y, z)
	m.verticesT = append(m.verticesT, x, y, z)
	m.vertexCount += 1
}

// Add a normal
func (m *Mesh) Normal3f(x, y, z float32) {
	m.hasNormals = true
	m.normals = append(m.normals, x, y, z)
}

// Add a texcoord
func (m *Mesh) TexCoord2f(u, v float32) {
	m.hasTexCoords = true
	m.texcoords = append(m.texcoords, u, v)
}

// Add a color
func (m *Mesh) Color4f(r, g, b, a float32) {
	m.hasColors = true
	m.colors = append(m.colors, r, g, b, a)
}

func (m *Mesh) Size() int {
	return int(m.vertexCount)
}

func (m *Mesh) Draw(x, y, z float32) {

	if m.Texture != nil {
		m.Texture.Bind()
	}

	// enable arrays
	gl.EnableClientState(gl.VERTEX_ARRAY)

	if m.hasNormals {
		gl.EnableClientState(gl.NORMAL_ARRAY)
	}

	if m.hasTexCoords {
		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	}

	if m.hasColors {
		gl.EnableClientState(gl.COLOR_ARRAY)
	}

	// Copy to static buffers
	if lastMeshCopy != m {
		for i, v := range m.vertices {
			v_arr[i] = v
		}
		if m.hasNormals {
			for i, v := range m.normals {
				n_arr[i] = v
			}
		}
		if m.hasTexCoords {
			for i, v := range m.texcoords {
				t_arr[i] = v
			}
		}
		if m.hasColors {
			for i, v := range m.colors {
				c_arr[i] = v
			}
		}
		lastMeshCopy = m
	}

	// setup pointers
	if !m.hasIntVerts {
		gl.VertexPointer(3, gl.FLOAT, 0, unsafe.Pointer(&v_arr[0]))
	} else {
		gl.VertexPointer(3, gl.INT, 0, unsafe.Pointer(&iv_arr[0]))
	}
	if m.hasTexCoords {
		gl.TexCoordPointer(2, gl.FLOAT, 0, unsafe.Pointer(&t_arr[0]))
	}
	if m.hasNormals {
		gl.NormalPointer(gl.FLOAT, 0, unsafe.Pointer(&n_arr[0]))
	}
	if m.hasColors {
		gl.NormalPointer(gl.FLOAT, 0, unsafe.Pointer(&c_arr[0]))
	}

	// draw it
	//log.Printf("gl.DrawArrays(%d, %d, %d)", m.thing, 0, m.vertexCount)
	gl.PushMatrix()
	gl.Translatef(x, y, z)
	gl.DrawArrays(m.thing, 0, m.vertexCount)
	gl.PopMatrix()

	// disable arrays

	gl.DisableClientState(gl.VERTEX_ARRAY)

	if m.hasNormals {
		gl.DisableClientState(gl.NORMAL_ARRAY)
	}

	if m.hasTexCoords {
		gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	}

	if m.hasColors {
		gl.DisableClientState(gl.COLOR_ARRAY)
	}

	if m.Texture != nil {
		m.Texture.Unbind()
	}
}

func (m *Mesh) DrawWithMeshBuffer(x, y, z float32) {

	//log.Println("Vertex count is", m.vertexCount)

	//m.SetTranslation(x, y, z)
	if !IntegerVertices && MeshBufferVertexCount > 0 && m.hasIntVerts {
		panic("Meshbuffer cannot mix INT and FLOAT types for vertices")
	}
	if IntegerVertices && MeshBufferVertexCount > 0 && !m.hasIntVerts {
		panic("Meshbuffer cannot mix INT and FLOAT types for vertices")
	}
	IntegerVertices = m.hasIntVerts

	if MAX_MESH_VERTICES-MeshBufferVertexCount <= m.vertexCount {
		// flush
		MeshBuffer_Flush()
		MeshBuffer_Begin(m.thing)
	} else if MeshBufferObjectType != m.thing {
		MeshBuffer_Flush()
		MeshBuffer_Begin(m.thing)
	}

	// Copy to static buffers - do translation here
	if !m.hasIntVerts {
		vals := []float32{x, y, z}

		for i, v := range m.vertices {
			v_arr[int(MeshBufferVertexCount)*3+i] = v + vals[i%3]
		}
		MeshBufferVertexCount += int32(len(m.vertices) / 3)

	} else {
		vals := []int32{int32(x), int32(y), int32(z)}

		for i, v := range m.ivertices {
			iv_arr[int(MeshBufferVertexCount)*3+i] = v + vals[i%3]
		}
		MeshBufferVertexCount += int32(len(m.ivertices) / 3)

	}
	//copy(v_arr[MeshBufferVertexCount:], m.verticesT)

	if m.hasNormals {
		for i, v := range m.normals {
			n_arr[int(MeshBufferNormalCount)*3+i] = v
		}

		//copy(n_arr[MeshBufferNormalCount:], m.normals)

		MeshBufferNormalCount += uint32(len(m.normals) / 3)
	}
	if m.hasTexCoords {
		for i, v := range m.texcoords {
			t_arr[int(MeshBufferTexCoordCount)*2+i] = v
		}
		//copy(n_arr[MeshBufferTexCoordCount:], m.texcoords)
		MeshBufferTexCoordCount += uint32(len(m.texcoords) / 2)
	}
	if m.hasColors {
		for i, v := range m.colors {
			c_arr[int(MeshBufferColorCount)*4+i] = v
		}
		//copy(n_arr[MeshBufferColorCount:], m.colors)
		MeshBufferColorCount += uint32(len(m.colors) / 4)
	}

}

func (m *Mesh) SetColor(r, g, b, a float32) {

	for i, _ := range m.colors {
		switch i % 4 {
		case 0:
			m.colors[i] = r
		case 1:
			m.colors[i] = g
		case 2:
			m.colors[i] = b
		case 3:
			m.colors[i] = a
		}
	}

}

func (m *Mesh) SetVertex3f(i int, x, y, z float32) {
	if m.vertexCount <= int32(i) {
		return
	}
	offs := 3 * i
	m.vertices[offs+0] = x
	m.vertices[offs+1] = y
	m.vertices[offs+2] = z
}

func (m *Mesh) SetNormal3f(i int, x, y, z float32) {
	if m.vertexCount <= int32(i) {
		return
	}
	offs := 3 * i
	m.normals[offs+0] = x
	m.normals[offs+1] = y
	m.normals[offs+2] = z
}

func (m *Mesh) SetColor4f(i int, r, g, b, a float32) {
	if m.vertexCount <= int32(i) {
		return
	}
	offs := 4 * i
	m.colors[offs+0] = r
	m.colors[offs+1] = g
	m.colors[offs+2] = b
	m.colors[offs+3] = a
}

func (m *Mesh) Translate3f(x, y, z float32) {

	copy(m.verticesT, m.vertices)
	var off int
	for i := 0; i < int(m.vertexCount); i++ {
		off = i * 3
		m.verticesT[off] = m.verticesT[off] + x
		m.verticesT[off+1] = m.verticesT[off+1] + y
		m.verticesT[off+2] = m.verticesT[off+2] + z
	}
}

func (m *Mesh) Translate3i(x, y, z int32) {

	copy(m.iverticesT, m.ivertices)
	var off int
	for i := 0; i < int(m.vertexCount); i++ {
		off = i * 3
		m.iverticesT[off] = m.iverticesT[off] + x
		m.iverticesT[off+1] = m.iverticesT[off+1] + y
		m.iverticesT[off+2] = m.iverticesT[off+2] + z
	}
}

// DrawSubMesh() draws a Mesh into another Mesh...
func (m *Mesh) DrawSubMesh(sm *Mesh, x, y, z float32) {

	//log.Println("Vertex count is", m.vertexCount)

	//m.SetTranslation(x, y, z)
	if !m.hasIntVerts && m.vertexCount > 0 && sm.hasIntVerts {
		panic("Meshbuffer cannot mix INT and FLOAT types for vertices")
	}
	if m.hasIntVerts && m.vertexCount > 0 && sm.hasIntVerts {
		panic("Meshbuffer cannot mix INT and FLOAT types for vertices")
	}
	m.hasIntVerts = sm.hasIntVerts

	// Copy to static buffers - do translation here
	if !sm.hasIntVerts {
		vals := []float32{x, y, z}

		for i, v := range sm.vertices {
			m.vertices = append(m.vertices, v+vals[i%3])
		}
		m.vertexCount += int32(len(sm.vertices) / 3)

	} else {
		vals := []int32{int32(x), int32(y), int32(z)}

		for i, v := range sm.ivertices {
			m.ivertices = append(m.ivertices, v+vals[i%3])
		}
		m.vertexCount += int32(len(sm.ivertices) / 3)

	}
	//copy(v_arr[MeshBufferVertexCount:], m.verticesT)

	if sm.hasNormals {
		m.normals = append(m.normals, sm.normals...)
	}
	if sm.hasTexCoords {
		m.texcoords = append(m.texcoords, sm.texcoords...)
	}
	if sm.hasColors {
		m.colors = append(m.colors, sm.colors...)
	}

}

// MeshPlacer holds a Mesh reference and a position
type MeshReference struct {
	TextureID uint32
	Primitive uint32
	Index     int
}

type MeshPlacer struct {
	X, Y, Z   float32
	M         *Mesh
	ID        string
	TextureID uint32
}

// MeshFlinger buckets meshes efficiently so it can make optimal use
// of the MeshBuffer
type MeshFlinger struct {
	// Map: m->{textureid} -> {primitive} -> []*Mesh
	Items       map[uint32]map[uint32][]*MeshPlacer
	YellowPages map[string]MeshReference
}

// Clear() removes all Mesh objects from the MeshFlinger
func (mf *MeshFlinger) Clear() {
	mf.Items = make(map[uint32]map[uint32][]*MeshPlacer)
	mf.YellowPages = make(map[string]MeshReference)
}

// Add() adds a new Mesh to the MeshFlinger
func (mf *MeshFlinger) Add3f(id string, m *Mesh, textureid uint32, x, y, z float32) {

	var primitive uint32
	if m.Texture != nil {
		textureid = m.Texture.handle
	}

	// now add
	tm, ex := mf.Items[textureid]
	if !ex {
		tm = make(map[uint32][]*MeshPlacer)
		mf.Items[textureid] = tm
	}

	pm, ex := tm[primitive]
	if !ex {
		pm = make([]*MeshPlacer, 0)
		tm[primitive] = pm
	}

	mp := &MeshPlacer{M: m, X: x, Y: y, Z: z, ID: id, TextureID: textureid}

	pm = append(pm, mp)

	// Set up a reference to the Mesh so we can find it quickly later
	index := len(pm) - 1
	tm[primitive] = pm
	mf.YellowPages[id] = MeshReference{Primitive: primitive, TextureID: textureid, Index: index}
}

// Fling() pushes the meshes efficiently to OpenGL
// Since this interacts with the GL context it should only ever be
// called from the same thread.
// Its also assumed that you have pre-applied your translation matrices
// (camera) etc at the time it is called.
// It may eat your hamster, but let's hope not.
func (mf *MeshFlinger) Fling() {

	var textureid uint32
	var primitive uint32
	var list []*MeshPlacer
	var tm map[uint32][]*MeshPlacer
	var mp *MeshPlacer
	var count int

	for textureid, tm = range mf.Items {

		for primitive, list = range tm {

			gl.BindTexture(gl.TEXTURE_2D, textureid)
			MeshBuffer_Begin(primitive)

			for _, mp = range list {
				mp.M.DrawWithMeshBuffer(mp.X, mp.Y, mp.Z)
				////fmt.Printf("----> Tex: %d at %d, %d, %d\n", textureid, mp.X, mp.Y, mp.Z )
			}

			MeshBuffer_End()

			if textureid != 0 {
				count++
			}
		}

	}

	////fmt.Printf("Flung %d meshbuffers\n", count)

}

// Exists returns the MeshReference and true/false if it is valid
func (mf *MeshFlinger) Exists(id string) (MeshReference, bool) {
	var ref MeshReference
	var ex bool

	ref, ex = mf.YellowPages[id]

	return ref, ex
}

// UpdateMesh updates the mesh used for the MeshPlacer, moving it in
// the tree if necessary
func (mf *MeshFlinger) UpdateMesh(id string, m *Mesh, textureid uint32) bool {

	ref, ex := mf.Exists(id)
	if !ex {
		return false // can't update non-existent mesh
	}

	//	//fmt.Printf("%d, %d, %d - %d\n", ref.TextureID, ref.Primitive, ref.Index, len(mf.Items[ref.TextureID][ref.Primitive]))

	mp := mf.Items[ref.TextureID][ref.Primitive][ref.Index]

	if mp.M == m && mp.TextureID == textureid {
		return true // no need to update as Mesh is *exactly* the same reference
	}

	// If we are here, then we need to move this MeshPlacer
	// delete it
	_ = mf.RemoveMesh(id)

	// Now we can cheat... use the Add3f() method to readd it with the same id, but new Mesh
	mf.Add3f(mp.ID, m, textureid, mp.X, mp.Y, mp.Z)

	return true
}

// RemoveMesh() removes a mesh using its ID
func (mf *MeshFlinger) RemoveMesh(id string) bool {
	ref, ex := mf.Exists(id)
	if !ex {
		return false // can't update non-existent mesh
	}

	//	//fmt.Printf("%d, %d, %d - %d (remove)\n", ref.TextureID, ref.Primitive, ref.Index, len(mf.Items[ref.TextureID][ref.Primitive]))

	l := mf.Items[ref.TextureID][ref.Primitive]

	lnew := make([]*MeshPlacer, len(l)-1)
	var found bool
	var sidx int
	for _, v := range l {
		if found {
			mr, ok := mf.YellowPages[v.ID]
			if ok {
				mr.Index = mr.Index - 1
				mf.YellowPages[v.ID] = mr
			}
		}

		if v.ID == id {
			found = true
		} else {
			lnew[sidx] = v
			sidx++
		}
	}
	mf.Items[ref.TextureID][ref.Primitive] = lnew

	// remove from map
	delete(mf.YellowPages, id)

	return true
}

// NewMeshFlinger() returns a new MeshFlinger instance
func NewMeshFlinger() *MeshFlinger {
	this := &MeshFlinger{}
	this.Clear()
	return this
}
