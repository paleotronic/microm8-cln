package video

import (
	//"os"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
)

// plane Primitive Definitions and methods
var planeTexcoords []float32 = []float32{
	1, 1,
	0, 1,
	0, 0,
	0, 0,
	1, 0,
	1, 1,
}

var planeTexcoordsInv []float32 = []float32{
	1, 0,
	0, 0,
	0, 1,
	0, 1,
	1, 1,
	1, 0,
}

var planeVertices []float32 = []float32{
	1, 1, 1, -1, 1, 1, -1, -1, 1, // v0-v1-v2 (front)
	-1, -1, 1, 1, -1, 1, 1, 1, 1, // v2-v3-v0
}

var planeNormals []float32 = []float32{
	0, 0, 1, 0, 0, 1, 0, 0, 1, // v0-v1-v2 (front)
	0, 0, 1, 0, 0, 1, 0, 0, 1, // v2-v3-v0
}

func GetTriangle(w, h float32) *glumby.Mesh {

	aw := w / 2
	ah := h / 2
	ad := float32(0)

	m := glumby.NewMesh(gl.TRIANGLES)

	for v := 3; v < 6; v++ {
		m.Vertex3f(planeVertices[v*3+0]*aw, planeVertices[v*3+1]*ah, planeVertices[v*3+2]*ad)
		m.Normal3f(planeNormals[v*3+0], planeNormals[v*3+1], planeNormals[v*3+2])
		m.TexCoord2f(planeTexcoords[v*2+0], planeTexcoords[v*2+1])
		m.Color4f(1, 1, 1, 1)
	}

	//log.Printf("Build plane with dimensions %f x %f\n", w, h)
	//os.Exit(1)

	return m
}

func GetPlaneAsTriangles(w, h float32) *glumby.Mesh {

	aw := w / 2
	ah := h / 2
	ad := float32(0)

	m := glumby.NewMesh(gl.TRIANGLES)

	for v := 0; v < 6; v++ {
		m.Vertex3f(planeVertices[v*3+0]*aw, planeVertices[v*3+1]*ah, planeVertices[v*3+2]*ad)
		m.Normal3f(planeNormals[v*3+0], planeNormals[v*3+1], planeNormals[v*3+2])
		m.TexCoord2f(planeTexcoords[v*2+0], planeTexcoords[v*2+1])
		m.Color4f(1, 1, 1, 1)
	}

	//log.Printf("Build plane with dimensions %f x %f\n", w, h)
	//os.Exit(1)

	return m
}

func GetPlaneAsTrianglesInv(w, h float32) *glumby.Mesh {

	aw := w / 2
	ah := h / 2
	ad := float32(0)

	m := glumby.NewMesh(gl.TRIANGLES)

	for v := 0; v < 6; v++ {
		m.Vertex3f(planeVertices[v*3+0]*aw, planeVertices[v*3+1]*ah, planeVertices[v*3+2]*ad)
		m.Normal3f(planeNormals[v*3+0], planeNormals[v*3+1], planeNormals[v*3+2])
		m.TexCoord2f(planeTexcoordsInv[v*2+0], planeTexcoordsInv[v*2+1])
		m.Color4f(1, 1, 1, 1)
	}

	//log.Printf("Build plane with dimensions %f x %f\n", w, h)
	//os.Exit(1)

	return m
}
