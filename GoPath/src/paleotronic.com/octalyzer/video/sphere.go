package video

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
)

type TriangleIndices struct {
	v1, v2, v3 int
}

type SphereBuilder struct {
	vertices           []mgl32.Vec3
	index              int
	faces              []TriangleIndices
	midpointIndexCache map[uint64]int
	recursionLevel     int
}

func NewSphereBuilder(recursions int) *SphereBuilder {
	sb := &SphereBuilder{
		vertices:           make([]mgl32.Vec3, 0, 12),
		index:              0,
		faces:              make([]TriangleIndices, 0, 20),
		midpointIndexCache: map[uint64]int{},
		recursionLevel:     recursions,
	}
	return sb.Build()
}

func (sb *SphereBuilder) addVertex(p mgl32.Vec3) int {
	length := p.Len()
	sb.vertices = append(sb.vertices, mgl32.Vec3{p[0] / length, p[1] / length, p[2] / length})
	return len(sb.vertices) - 1
}

func (sb *SphereBuilder) getMiddlePoint(p1, p2 int) int {
	var firstIsSmaller = p1 < p2
	var smallerIndex, greaterIndex int
	if firstIsSmaller {
		smallerIndex = p1
		greaterIndex = p2
	} else {
		smallerIndex = p2
		greaterIndex = p1
	}
	var key = (uint64(smallerIndex) << 32) | uint64(greaterIndex)
	var ret int
	var ok bool
	if ret, ok = sb.midpointIndexCache[key]; ok {
		return ret
	}
	var point1 = sb.vertices[p1]
	var point2 = sb.vertices[p2]
	var middle = mgl32.Vec3{
		(point1[0] + point2[0]) / 2,
		(point1[1] + point2[1]) / 2,
		(point1[2] + point2[2]) / 2,
	}

	var i = sb.addVertex(middle)
	sb.midpointIndexCache[key] = i
	return i
}

func (sb *SphereBuilder) Build() *SphereBuilder {
	var t = float32((1.0 + math.Sqrt(5.0)) / 2.0)

	sb.addVertex(mgl32.Vec3{-1, t, 0})
	sb.addVertex(mgl32.Vec3{1, t, 0})
	sb.addVertex(mgl32.Vec3{-1, -t, 0})
	sb.addVertex(mgl32.Vec3{1, -t, 0})

	sb.addVertex(mgl32.Vec3{0, -1, t})
	sb.addVertex(mgl32.Vec3{0, 1, t})
	sb.addVertex(mgl32.Vec3{0, -1, -t})
	sb.addVertex(mgl32.Vec3{0, 1, -t})

	sb.addVertex(mgl32.Vec3{t, 0, -1})
	sb.addVertex(mgl32.Vec3{t, 0, 1})
	sb.addVertex(mgl32.Vec3{-t, 0, -1})
	sb.addVertex(mgl32.Vec3{-t, 0, 1})

	sb.faces = append(sb.faces, TriangleIndices{0, 11, 5})
	sb.faces = append(sb.faces, TriangleIndices{0, 5, 1})
	sb.faces = append(sb.faces, TriangleIndices{0, 1, 7})
	sb.faces = append(sb.faces, TriangleIndices{0, 7, 10})
	sb.faces = append(sb.faces, TriangleIndices{0, 10, 11})

	// 5 adjacent faces
	sb.faces = append(sb.faces, TriangleIndices{1, 5, 9})
	sb.faces = append(sb.faces, TriangleIndices{5, 11, 4})
	sb.faces = append(sb.faces, TriangleIndices{11, 10, 2})
	sb.faces = append(sb.faces, TriangleIndices{10, 7, 6})
	sb.faces = append(sb.faces, TriangleIndices{7, 1, 8})

	// 5 faces around point 3
	sb.faces = append(sb.faces, TriangleIndices{3, 9, 4})
	sb.faces = append(sb.faces, TriangleIndices{3, 4, 2})
	sb.faces = append(sb.faces, TriangleIndices{3, 2, 6})
	sb.faces = append(sb.faces, TriangleIndices{3, 6, 8})
	sb.faces = append(sb.faces, TriangleIndices{3, 8, 9})

	// 5 adjacent faces
	sb.faces = append(sb.faces, TriangleIndices{4, 9, 5})
	sb.faces = append(sb.faces, TriangleIndices{2, 4, 11})
	sb.faces = append(sb.faces, TriangleIndices{6, 2, 10})
	sb.faces = append(sb.faces, TriangleIndices{8, 6, 7})
	sb.faces = append(sb.faces, TriangleIndices{9, 8, 1})

	// refine triangles
	for i := 0; i < sb.recursionLevel; i++ {
		var faces2 = make([]TriangleIndices, 0, len(sb.faces))
		for _, tri := range sb.faces {
			// replace triangle by 4 triangles
			var a = sb.getMiddlePoint(tri.v1, tri.v2)
			var b = sb.getMiddlePoint(tri.v2, tri.v3)
			var c = sb.getMiddlePoint(tri.v3, tri.v1)

			faces2 = append(faces2, TriangleIndices{tri.v1, a, c})
			faces2 = append(faces2, TriangleIndices{tri.v2, b, a})
			faces2 = append(faces2, TriangleIndices{tri.v3, c, b})
			faces2 = append(faces2, TriangleIndices{a, b, c})
		}
		sb.faces = faces2
	}

	return sb
}

func (sb *SphereBuilder) MeshTriangles(radius float32) *glumby.Mesh {
	m := glumby.NewMesh(gl.TRIANGLES)
	var v1, v2, v3 mgl32.Vec3
	for _, face := range sb.faces {
		v1, v2, v3 = sb.vertices[face.v1].Mul(radius), sb.vertices[face.v2].Mul(radius), sb.vertices[face.v3].Mul(radius)
		m.Triangle3v(v1, v2, v3, calcNormalForTriangleInv(v1, v2, v3))
	}
	return m
}

var sphereBuilder = NewSphereBuilder(2)
