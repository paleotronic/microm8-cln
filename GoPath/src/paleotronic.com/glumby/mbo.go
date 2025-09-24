package glumby

import (
	"paleotronic.com/fmt"
	"unsafe"

	"paleotronic.com/gl"
	//"paleotronic.com/fmt"
)

type MeshBufferObject struct {
	iv_arr          [MAX_MESH_VERTICES * 3]int32
	v_arr           [MAX_MESH_VERTICES * 3]float32
	c_arr           [MAX_MESH_VERTICES * 4]float32
	n_arr           [MAX_MESH_VERTICES * 3]float32
	t_arr           [MAX_MESH_VERTICES * 2]float32
	VertexCount     int32 // always between 0 and MAX_MESH_VERTICES - 1
	NormalCount     uint32
	ColorCount      uint32
	TexCoordCount   uint32
	ObjectType      uint32
	IntegerVertices bool
	FlushCounter    int
}

func (mbo *MeshBufferObject) Begin(thing uint32) {
	mbo.ObjectType = thing
	mbo.VertexCount = 0
	mbo.ColorCount = 0
	mbo.NormalCount = 0
	mbo.TexCoordCount = 0
}

func (mbo *MeshBufferObject) ResetCount() {
	mbo.FlushCounter = 0
}

func (mbo *MeshBufferObject) GetFlushCount() int {
	return mbo.FlushCounter
}

func (mbo *MeshBufferObject) Send(final bool) {
	if mbo.VertexCount > 0 {

		final = true

		//log.Printf("Flushing %d vertices from buffer\n", mbo.VertexCount)
		// enable arrays
		gl.EnableClientState(gl.VERTEX_ARRAY)

		if mbo.NormalCount > 0 {
			gl.EnableClientState(gl.NORMAL_ARRAY)
		}

		if mbo.TexCoordCount > 0 {
			gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		}

		if mbo.ColorCount > 0 {
			gl.EnableClientState(gl.COLOR_ARRAY)
		}

		// do the send here
		if !IntegerVertices {
			gl.VertexPointer(3, gl.FLOAT, 0, unsafe.Pointer(&mbo.v_arr[0]))
		} else {
			gl.VertexPointer(3, gl.INT, 0, unsafe.Pointer(&mbo.iv_arr[0]))
		}
		if mbo.TexCoordCount > 0 {
			gl.TexCoordPointer(2, gl.FLOAT, 0, unsafe.Pointer(&mbo.t_arr[0]))
		}
		if mbo.NormalCount > 0 {
			gl.NormalPointer(gl.FLOAT, 0, unsafe.Pointer(&mbo.n_arr[0]))
		}
		if mbo.ColorCount > 0 {
			gl.ColorPointer(4, gl.FLOAT, 0, unsafe.Pointer(&mbo.c_arr[0]))
		}

		gl.Enable(gl.COLOR_MATERIAL)
		gl.DrawArrays(mbo.ObjectType, 0, mbo.VertexCount)

		// disable arrays
		gl.DisableClientState(gl.VERTEX_ARRAY)

		if mbo.NormalCount > 0 {
			gl.DisableClientState(gl.NORMAL_ARRAY)
		}

		if mbo.TexCoordCount > 0 {
			gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
		}

		if mbo.ColorCount > 0 {
			gl.DisableClientState(gl.COLOR_ARRAY)
		}

		if final {
			gl.Flush()
		}
		mbo.FlushCounter++
	}
	//~ mbo.VertexCount = 0
	//~ mbo.ColorCount = 0
	//~ mbo.NormalCount = 0
	//~ mbo.TexCoordCount = 0
}

func (mbo *MeshBufferObject) Draw(x, y, z float32, m *Mesh) {
	//log.Println("Vertex count is", m.vertexCount)

	//m.SetTranslation(x, y, z)
	if !mbo.IntegerVertices && mbo.VertexCount > 0 && m.hasIntVerts {
		panic("mbo. cannot mix INT and FLOAT types for vertices")
	}
	if mbo.IntegerVertices && mbo.VertexCount > 0 && !m.hasIntVerts {
		panic("mbo. cannot mix INT and FLOAT types for vertices")
	}
	mbo.IntegerVertices = m.hasIntVerts

	if MAX_MESH_VERTICES-mbo.VertexCount <= m.vertexCount {
		// flush
		fmt.Println("Flush")
		mbo.Send(false)
		mbo.Begin(m.thing)
	} else if mbo.ObjectType != m.thing {
		mbo.Send(false)
		mbo.Begin(m.thing)
	}

	// Copy to static buffers - do translation here
	if !m.hasIntVerts {
		vals := []float32{x, y, z}

		for i, v := range m.vertices {
			mbo.v_arr[int(mbo.VertexCount)*3+i] = v + vals[i%3]
		}
		mbo.VertexCount += int32(len(m.vertices) / 3)

	} else {
		vals := []int32{int32(x), int32(y), int32(z)}

		for i, v := range m.ivertices {
			mbo.iv_arr[int(mbo.VertexCount)*3+i] = v + vals[i%3]
		}
		mbo.VertexCount += int32(len(m.ivertices) / 3)

	}
	//copy(v_arr[mbo.VertexCount:], m.verticesT)

	if m.hasNormals {
		for i, v := range m.normals {
			mbo.n_arr[int(mbo.NormalCount)*3+i] = v
		}

		//copy(n_arr[mbo.NormalCount:], m.normals)

		mbo.NormalCount += uint32(len(m.normals) / 3)
	}
	if m.hasTexCoords {
		for i, v := range m.texcoords {
			mbo.t_arr[int(mbo.TexCoordCount)*2+i] = v
		}
		//copy(n_arr[mbo.TexCoordCount:], m.texcoords)
		mbo.TexCoordCount += uint32(len(m.texcoords) / 2)
	}
	if m.hasColors {
		for i, v := range m.colors {
			mbo.c_arr[int(mbo.ColorCount)*4+i] = v
		}
		//copy(n_arr[mbo.ColorCount:], m.colors)
		mbo.ColorCount += uint32(len(m.colors) / 4)
	}
}

func (mbo *MeshBufferObject) DrawWithTexCoords(x, y, z float32, m *Mesh, tc []float32) {
	//log.Println("Vertex count is", m.vertexCount)

	//m.SetTranslation(x, y, z)
	if !mbo.IntegerVertices && mbo.VertexCount > 0 && m.hasIntVerts {
		panic("mbo. cannot mix INT and FLOAT types for vertices")
	}
	if mbo.IntegerVertices && mbo.VertexCount > 0 && !m.hasIntVerts {
		panic("mbo. cannot mix INT and FLOAT types for vertices")
	}
	mbo.IntegerVertices = m.hasIntVerts

	if MAX_MESH_VERTICES-mbo.VertexCount <= m.vertexCount {
		// flush
		mbo.Send(false)
		mbo.Begin(m.thing)
	} else if mbo.ObjectType != m.thing {
		mbo.Send(false)
		mbo.Begin(m.thing)
	}

	// Copy to static buffers - do translation here
	if !m.hasIntVerts {
		vals := []float32{x, y, z}

		for i, v := range m.vertices {
			mbo.v_arr[int(mbo.VertexCount)*3+i] = v + vals[i%3]
		}
		mbo.VertexCount += int32(len(m.vertices) / 3)

	} else {
		vals := []int32{int32(x), int32(y), int32(z)}

		for i, v := range m.ivertices {
			mbo.iv_arr[int(mbo.VertexCount)*3+i] = v + vals[i%3]
		}
		mbo.VertexCount += int32(len(m.ivertices) / 3)

	}
	//copy(v_arr[mbo.VertexCount:], m.verticesT)

	if m.hasNormals {
		for i, v := range m.normals {
			mbo.n_arr[int(mbo.NormalCount)*3+i] = v
		}

		//copy(n_arr[mbo.NormalCount:], m.normals)

		mbo.NormalCount += uint32(len(m.normals) / 3)
	}
	if m.hasTexCoords {
		for i, v := range tc {
			mbo.t_arr[int(mbo.TexCoordCount)*2+i] = v
		}
		//copy(n_arr[mbo.TexCoordCount:], m.texcoords)
		mbo.TexCoordCount += uint32(len(m.texcoords) / 2)
	}
	if m.hasColors {
		for i, v := range m.colors {
			mbo.c_arr[int(mbo.ColorCount)*4+i] = v
		}
		//copy(n_arr[mbo.ColorCount:], m.colors)
		mbo.ColorCount += uint32(len(m.colors) / 4)
	}
}
