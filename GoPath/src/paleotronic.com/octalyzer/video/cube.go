package video

import (
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
)

// Cube Primitive Definitions and methods
var cubeTexcoords []float32 = []float32{
	1, 1,
	0, 1,
	0, 0,
	0, 0,
	1, 0,
	1, 1,
	1, 1,
	0, 1,
	0, 0,
	0, 0,
	1, 0,
	1, 1,
	1, 1,
	0, 1,
	0, 0,
	0, 0,
	1, 0,
	1, 1,
	1, 1,
	0, 1,
	0, 0,
	0, 0,
	1, 0,
	1, 1,
	1, 1,
	0, 1,
	0, 0,
	0, 0,
	1, 0,
	1, 1,
	1, 1,
	0, 1,
	0, 0,
	0, 0,
	1, 0,
	1, 1,
}

// top, bot, front, back, left, right

var cubeVertices []float32 = []float32{

	1, 1, 1, 1, 1, -1, -1, 1, -1, // v0-v5-v6 (top)
	-1, 1, -1, -1, 1, 1, 1, 1, 1, // v6-v1-v0

	-1, -1, -1, 1, -1, -1, 1, -1, 1, // v7-v4-v3 (bottom)
	1, -1, 1, -1, -1, 1, -1, -1, -1, // v3-v2-v7

	1, 1, 1, -1, 1, 1, -1, -1, 1, // v0-v1-v2 (front)
	-1, -1, 1, 1, -1, 1, 1, 1, 1, // v2-v3-v0

	1, -1, -1, -1, -1, -1, -1, 1, -1, // v4-v7-v6 (back)
	-1, 1, -1, 1, 1, -1, 1, -1, -1, // v6-v5-v4

	-1, 1, 1, -1, 1, -1, -1, -1, -1, // v1-v6-v7 (left)
	-1, -1, -1, -1, -1, 1, -1, 1, 1, // v7-v2-v1

	1, 1, 1, 1, -1, 1, 1, -1, -1, // v0-v3-v4 (right)
	1, -1, -1, 1, 1, -1, 1, 1, 1, // v4-v5-v0

}

var cubeNormals []float32 = []float32{

	0, 1, 0, 0, 1, 0, 0, 1, 0, // v0-v5-v6 (top)
	0, 1, 0, 0, 1, 0, 0, 1, 0, // v6-v1-v0

	0, -1, 0, 0, -1, 0, 0, -1, 0, // v7-v4-v3 (bottom)
	0, -1, 0, 0, -1, 0, 0, -1, 0, // v3-v2-v7

	0, 0, 1, 0, 0, 1, 0, 0, 1, // v0-v1-v2 (front)
	0, 0, 1, 0, 0, 1, 0, 0, 1, // v2-v3-v0

	0, 0, -1, 0, 0, -1, 0, 0, -1, // v4-v7-v6 (back)
	0, 0, -1, 0, 0, -1, 0, 0, -1, // v6-v5-v4

	-1, 0, 0, -1, 0, 0, -1, 0, 0, // v1-v6-v7 (left)
	-1, 0, 0, -1, 0, 0, -1, 0, 0, // v7-v2-v1

	1, 0, 0, 1, 0, 0, 1, 0, 0, // v0-v3-v4 (right)
	1, 0, 0, 1, 0, 0, 1, 0, 0, // v4-v5-v0

}

var cubeNormalsAlt []float32 = []float32{

	0, 1, 0, 0, 1, 0, 0, 1, 0, // v0-v5-v6 (top)
	0, 1, 0, 0, 1, 0, 0, 1, 0, // v6-v1-v0

	0, -1, 0, 0, -1, 0, 0, -1, 0, // v7-v4-v3 (bottom)
	0, -1, 0, 0, -1, 0, 0, -1, 0, // v3-v2-v7

	0, 0, -1, 0, 0, -1, 0, 0, -1, // v0-v1-v2 (front)
	0, 0, -1, 0, 0, -1, 0, 0, -1, // v2-v3-v0

	0, 0, 1, 0, 0, 1, 0, 0, 1, // v4-v7-v6 (back)
	0, 0, 1, 0, 0, 1, 0, 0, 1, // v6-v5-v4

	-1, 0, 0, -1, 0, 0, -1, 0, 0, // v1-v6-v7 (left)
	-1, 0, 0, -1, 0, 0, -1, 0, 0, // v7-v2-v1

	1, 0, 0, 1, 0, 0, 1, 0, 0, // v0-v3-v4 (right)
	1, 0, 0, 1, 0, 0, 1, 0, 0, // v4-v5-v0

}

func GetCubeAsTriangles(w, h, d float32) *glumby.Mesh {

	aw := w / 2
	ah := h / 2
	ad := d / 2

	m := glumby.NewMesh(gl.TRIANGLES)

	for v := 0; v < 36; v++ {
		m.Vertex3f(cubeVertices[v*3+0]*aw, cubeVertices[v*3+1]*ah, cubeVertices[v*3+2]*ad)
		m.Normal3f(cubeNormals[v*3+0], cubeNormals[v*3+1], cubeNormals[v*3+2])
		m.TexCoord2f(cubeTexcoords[v*2+0], cubeTexcoords[v*2+1])
		m.Color4f(1, 1, 1, 1)
	}

	return m
}
