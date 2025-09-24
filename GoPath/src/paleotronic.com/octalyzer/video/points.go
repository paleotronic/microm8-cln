package video

import (
	//"paleotronic.com/core/settings"

	"math"

	"github.com/go-gl/mathgl/mgl32"
	"paleotronic.com/gl"
	"paleotronic.com/glumby"
	//"os"
)

func GetWorldToLocalMatrix(V, U, R mgl32.Vec3, v, u, r mgl32.Vec3, sc mgl32.Vec3) mgl32.Mat4 {

	// var s = mgl32.Mat4{
	// 	sc[0],  // col 0
	// 	0,
	// 	0,
	// 	0,

	// 	0,		// col 1
	// 	sc[1],
	// 	0,
	// 	0,

	// 	0, 		// col 2
	// 	0,
	// 	sc[2],
	// 	0,

	// 	0,		// col 3
	// 	0,
	// 	0,
	// 	1,
	// }

	// log2.Printf("%v", mgl32.TransformCoordinate(mgl32.Vec3{1,1,1}, s))
	// os.Exit(1)

	var m = mgl32.Mat4{
		r.Dot(R), // col 0
		r.Dot(V),
		r.Dot(U),
		0,

		v.Dot(R), // col 1
		v.Dot(V),
		v.Dot(U),
		0,

		u.Dot(R), // col 2
		u.Dot(V),
		u.Dot(U),
		0,

		0, // col 3
		0,
		0,
		1,
	}

	// M11 = (float)Vector3d.DotProduct(X1Prime, X1),
	// M12 = (float)Vector3d.DotProduct(X1Prime, X2),
	// M13 = (float)Vector3d.DotProduct(X1Prime, X3),
	// M21 = (float)Vector3d.DotProduct(X2Prime, X1),
	// M22 = (float)Vector3d.DotProduct(X2Prime, X2),
	// M23 = (float)Vector3d.DotProduct(X2Prime, X3),
	// M31 = (float)Vector3d.DotProduct(X3Prime, X1),
	// M32 = (float)Vector3d.DotProduct(X3Prime, X2),
	// M33 = (float)Vector3d.DotProduct(X3Prime, X3),

	//return s.Mul4(m)
	return m
}

func GetRelativePoint(
	origin mgl32.Vec3,
	relative mgl32.Vec3,
	viewdir, updir, rightdir mgl32.Vec3,
) mgl32.Vec3 {

	//result := origin

	// X
	//result[0] += relative[0]*rightdir[0] + relative[1]*viewdir[0] + relative[2]*updir[0]
	//result[1] += relative[0]*rightdir[1] + relative[1]*viewdir[1] + relative[2]*updir[1]
	//result[2] += relative[0]*rightdir[2] + relative[1]*viewdir[2] + relative[2]*updir[2]

	result := origin

	// X
	result[0] = result[0] + relative[0]*rightdir[0]
	result[1] = result[1] + relative[0]*rightdir[1]
	result[2] = result[2] + relative[0]*rightdir[2]

	// Y
	result[0] = result[0] + relative[1]*viewdir[0]
	result[1] = result[1] + relative[1]*viewdir[1]
	result[2] = result[2] + relative[1]*viewdir[2]

	// Z
	result[0] = result[0] + relative[2]*updir[0]
	result[1] = result[1] + relative[2]*updir[1]
	result[2] = result[2] + relative[2]*updir[2]

	//log2.Printf("Translate point (%f,%f,%f) to (%f, %f, %f)", relative[0], relative[1], relative[2], result[0], result[1], result[2])

	return result

}

var cv []float32 = []float32{

	1, 1, 1, 1, 1, 0, 0, 1, 0, // v0-v5-v6 (top)
	0, 1, 0, 0, 1, 1, 1, 1, 1, // v6-v1-v0

	0, 0, 0, 1, 0, 0, 1, 0, 1, // v7-v4-v3 (bottom)
	1, 0, 1, 0, 0, 1, 0, 0, 0, // v3-v2-v7

	1, 1, 1, 0, 1, 1, 0, 0, 1, // v0-v1-v2 (front)
	0, 0, 1, 1, 0, 1, 1, 1, 1, // v2-v3-v0

	1, 0, 0, 0, 0, 0, 0, 1, 0, // v4-v7-v6 (back)
	0, 1, 0, 1, 1, 0, 1, 0, 0, // v6-v5-v4

	0, 1, 1, 0, 1, 0, 0, 0, 0, // v1-v6-v7 (left)
	0, 0, 0, 0, 0, 1, 0, 1, 1, // v7-v2-v1

	1, 1, 1, 1, 0, 1, 1, 0, 0, // v0-v3-v4 (right)
	1, 0, 0, 1, 1, 0, 1, 1, 1, // v4-v5-v0

}

func DegToRad(angle float32) float64 {
	return float64(angle) * math.Pi / 180
}

func RotatePointYPRQ(v0 mgl32.Vec3, yaw, pitch, roll float32) mgl32.Vec3 {
	u := mgl32.DegToRad(roll)
	v := mgl32.DegToRad(pitch)
	w := mgl32.DegToRad(yaw)
	q := mgl32.AnglesToQuat(w, v, u, mgl32.ZYX)
	return q.Rotate(v0)
}

func RotatePointYPR(v0 mgl32.Vec3, yaw, pitch, roll float32) mgl32.Vec3 {
	u := DegToRad(roll)
	v := DegToRad(pitch)
	w := DegToRad(yaw)

	x0, y0, z0 := float64(v0[0]), float64(v0[1]), float64(v0[2])

	x1 := x0
	y1 := y0*math.Cos(u) - z0*math.Sin(u)
	z1 := y0*math.Sin(u) + z0*math.Cos(u)

	x2 := x1*math.Cos(v) + z1*math.Sin(v)
	y2 := y1
	z2 := -x1*math.Sin(v) + z1*math.Cos(v)

	x3 := x2*math.Cos(w) - y2*math.Sin(w)
	y3 := x2*math.Sin(w) + y2*math.Cos(w)
	z3 := z2

	return mgl32.Vec3{
		float32(x3),
		float32(y3),
		float32(z3),
	}
}

func GetCubeAsTrianglesRel(
	w, h, d float32,
	o, v, u, r mgl32.Vec3,
) *glumby.Mesh {

	m := glumby.NewMesh(gl.TRIANGLES)

	var rv, p mgl32.Vec3

	mat := GetWorldToLocalMatrix(
		mgl32.Vec3{0, 1, 0},
		mgl32.Vec3{0, 0, 1},
		mgl32.Vec3{1, 0, 0},
		v,
		u,
		r,
		mgl32.Vec3{w, h, d},
	)

	for vv := 0; vv < 36; vv++ {
		p = mgl32.Vec3{cv[vv*3+0], cv[vv*3+1], cv[vv*3+2]}
		//rv = GetRelativePoint(o, p, v, u, r)

		rv = mgl32.TransformCoordinate(p, mat)
		//r = q.Rotate(p)
		m.Vertex3f(rv[0], rv[1], rv[2])
		m.Normal3f(cubeNormalsAlt[vv*3+0], cubeNormalsAlt[vv*3+1], cubeNormalsAlt[vv*3+2])
		m.TexCoord2f(cubeTexcoords[vv*2+0], cubeTexcoords[vv*2+1])
		m.Color4f(1, 1, 1, 1)
	}

	return m
}

var lineCubeVertices = []float32{
	1.0, 1.0, 1.0, // Vertex 0 (X, Y, Z)
	0.0, 1.0, 1.0, // Vertex 1 (X, Y, Z)
	0.0, 1.0, 1.0, // Vertex 1 (X, Y, Z)
	0.0, 0.0, 1.0, // Vertex 2 (X, Y, Z)
	0.0, 0.0, 1.0, // Vertex 2 (X, Y, Z)
	1.0, 0.0, 1.0, // Vertex 3 (X, Y, Z)
	1.0, 0.0, 1.0, // Vertex 3 (X, Y, Z)
	1.0, 1.0, 1.0, // Vertex 0 (X, Y, Z)
	1.0, 1.0, 0.0, // Vertex 4 (X, Y, Z)
	0.0, 1.0, 0.0, // Vertex 5 (X, Y, Z)
	0.0, 1.0, 0.0, // Vertex 5 (X, Y, Z)
	0.0, 0.0, 0.0, // Vertex 6 (X, Y, Z)
	0.0, 0.0, 0.0, // Vertex 6 (X, Y, Z)
	1.0, 0.0, 0.0, // Vertex 7 (X, Y, Z)
	1.0, 0.0, 0.0, // Vertex 7 (X, Y, Z)
	1.0, 1.0, 0.0, // Vertex 4 (X, Y, Z)
	1.0, 1.0, 1.0, // Vertex 0 (X, Y, Z)
	1.0, 1.0, 0.0, // Vertex 4 (X, Y, Z)
	0.0, 1.0, 1.0, // Vertex 1 (X, Y, Z)
	0.0, 1.0, 0.0, // Vertex 5 (X, Y, Z)
	0.0, 0.0, 1.0, // Vertex 2 (X, Y, Z)
	0.0, 0.0, 0.0, // Vertex 6 (X, Y, Z)
	1.0, 0.0, 1.0, // Vertex 3 (X, Y, Z)
	1.0, 0.0, 0.0, // Vertex 7 (X, Y, Z)
}

func GetCubeAsLinesRel(
	w, h, d float32,
	v, u, r mgl32.Vec3,
) *glumby.Mesh {

	m := glumby.NewMesh(gl.LINES)

	var ra, p mgl32.Vec3

	mat := GetWorldToLocalMatrix(
		mgl32.Vec3{0, 1, 0},
		mgl32.Vec3{0, 0, 1},
		mgl32.Vec3{1, 0, 0},
		v,
		u,
		r,
		mgl32.Vec3{w, h, d},
	)

	for v := 0; v < 24; v++ {
		p = mgl32.Vec3{lineCubeVertices[v*3+0], lineCubeVertices[v*3+1], -lineCubeVertices[v*3+2]}
		ra = mgl32.TransformCoordinate(p, mat)
		m.Vertex3f(ra[0], ra[1], ra[2])
		m.Color4f(1, 1, 1, 1)
	}

	return m
}

var smallUnitCircle = GetCircleMesh(10)
var largeUnitCircle = GetCircleMesh(60)

func GetCircleMesh(segments int) []mgl32.Vec3 {
	var out = make([]mgl32.Vec3, segments)
	var theta, x, y float64
	for ii := 0; ii < segments; ii++ {
		theta = 2.0 * 3.1415926 * float64(ii) / float64(segments) //get the current angle
		x = math.Cos(theta)                                       //calculate the x component
		y = math.Sin(theta)                                       //calculate the y component
		out[ii] = mgl32.Vec3{float32(x), float32(y), 0}
	}
	return out
}

func GetArcMesh(segments int, angle float64, sangle float64, addCenter bool) []mgl32.Vec3 {
	var out = make([]mgl32.Vec3, 0, segments)
	var theta, x, y float64
	var ca = angle / float64(segments)
	for ii := 0; ii <= segments; ii++ {
		theta = ((ca * float64(ii) / 360) + (sangle / 360)) * 2.0 * 3.1415926 //get the current angle
		x = math.Cos(theta)                                                   //calculate the x component
		y = math.Sin(theta)                                                   //calculate the y component
		out = append(out, mgl32.Vec3{float32(x), float32(y), 0})
	}
	if addCenter {
		out = append(out, mgl32.Vec3{0, 0, 0})
		//log2.Printf("Adding extra zero point for pie... mmmm pie")
	}
	return out
}

func GetArcAsLinesRel(
	t uint32,
	rad float32,
	ang float32,
	sang float32,
	v, u, r mgl32.Vec3,
) *glumby.Mesh {

	m := glumby.NewMesh(t)

	var ra mgl32.Vec3

	var vv []mgl32.Vec3

	if rad < 30 {
		vv = GetArcMesh(10, float64(ang), float64(sang), t == gl.TRIANGLE_FAN)
	} else {
		vv = GetArcMesh(30, float64(ang), float64(sang), t == gl.TRIANGLE_FAN)
	}

	mat := GetWorldToLocalMatrix(
		mgl32.Vec3{0, 1, 0},
		mgl32.Vec3{0, 0, 1},
		mgl32.Vec3{1, 0, 0},
		v,
		u,
		r,
		mgl32.Vec3{1, 1, 1},
	)

	for _, p := range vv {
		//ra = GetRelativePoint(mgl32.Vec3{}, p.Mul(rad), v, u, r )
		ra = mgl32.TransformCoordinate(p.Mul(rad), mat)
		m.Vertex3f(ra[0], ra[1], ra[2])
		m.Color4f(1, 1, 1, 1)
	}

	return m
}

func GetCircleAsLinesRel(
	t uint32,
	rad float32,
	v, u, r mgl32.Vec3,
) *glumby.Mesh {

	m := glumby.NewMesh(t)

	var ra mgl32.Vec3

	var vv []mgl32.Vec3

	if rad > 30 {
		vv = largeUnitCircle
	} else {
		vv = smallUnitCircle
	}

	mat := GetWorldToLocalMatrix(
		mgl32.Vec3{0, 1, 0},
		mgl32.Vec3{0, 0, 1},
		mgl32.Vec3{1, 0, 0},
		v,
		u,
		r,
		mgl32.Vec3{1, 1, 1},
	)

	for _, p := range vv {
		//ra = GetRelativePoint(mgl32.Vec3{}, p.Mul(rad), v, u, r )
		ra = mgl32.TransformCoordinate(p.Mul(rad), mat)
		m.Vertex3f(ra[0], ra[1], ra[2])
		m.Color4f(1, 1, 1, 1)
	}

	return m
}

var polyCache = map[int][]mgl32.Vec3{}

func GetPolyAsLinesRel(
	t uint32,
	rad float32,
	sides float32,
	v, u, r mgl32.Vec3,
) *glumby.Mesh {

	m := glumby.NewMesh(t)

	var ra mgl32.Vec3

	var vv []mgl32.Vec3
	var ok bool

	if vv, ok = polyCache[int(sides)]; !ok {
		vv = GetCircleMesh(int(sides))
		polyCache[int(sides)] = vv
	}

	mat := GetWorldToLocalMatrix(
		mgl32.Vec3{0, 1, 0},
		mgl32.Vec3{0, 0, 1},
		mgl32.Vec3{1, 0, 0},
		v,
		u,
		r,
		mgl32.Vec3{1, 1, 1},
	)

	for _, p := range vv {
		//ra = GetRelativePoint(mgl32.Vec3{}, p.Mul(rad), v, u, r )
		ra = mgl32.TransformCoordinate(p.Mul(rad), mat)
		m.Vertex3f(ra[0], ra[1], ra[2])
		m.Color4f(1, 1, 1, 1)
	}

	return m
}

func EulerToQuat(ya, pi, ro float32) mgl32.Quat {

	heading := DegToRad(ya)
	attitude := DegToRad(pi)
	bank := DegToRad(ro)

	C1 := math.Cos(heading)
	C2 := math.Cos(attitude)
	C3 := math.Cos(bank)
	S1 := math.Sin(heading)
	S2 := math.Sin(attitude)
	S3 := math.Sin(bank)

	w := math.Sqrt(1.0+C1*C2+C1*C3-S1*S2*S3+C2*C3) / 2
	x := (C2*S3 + C1*S3 + S1*S2*C3) / (4.0 * w)
	y := (S1*C2 + S1*C3 + C1*S2*S3) / (4.0 * w)
	z := (-S1*S3 + C1*S2*C3 + S2) / (4.0 * w)

	return mgl32.Quat{W: float32(w), V: mgl32.Vec3{float32(x), float32(y), float32(z)}}

}

/*
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
*/

func GetPyramidAsTrianglesV(a, b, c, d, e mgl32.Vec3) *glumby.Mesh {
	m := glumby.NewMesh(gl.TRIANGLES)

	//log2.Printf("pyramid called with %v, %v, %v, %v, %v", a, b, c, d, e)

	// front
	m.Triangle3v(d, a, e, mgl32.Vec3{0, 0, 1})
	// back
	m.Triangle3v(b, c, e, mgl32.Vec3{0, 0, -1})
	// left
	m.Triangle3v(a, b, e, mgl32.Vec3{-1, 0, 0})
	// right
	m.Triangle3v(c, d, e, mgl32.Vec3{1, 0, 0})
	// bot
	m.Triangle3v(d, a, b, mgl32.Vec3{0, -1, 0})
	m.Triangle3v(d, b, c, mgl32.Vec3{0, -1, 0})
	// triangle mesh with normal
	return m
}

func GetCubeAsTrianglesV(v0, v1, v2, v3, v4, v5, v6, v7 mgl32.Vec3) *glumby.Mesh {
	m := glumby.NewMesh(gl.TRIANGLES)
	// front
	m.Triangle3v(v0, v4, v7, mgl32.Vec3{0, 0, 1})
	m.Triangle3v(v0, v3, v7, mgl32.Vec3{0, 0, 1})
	// back
	m.Triangle3v(v1, v5, v6, mgl32.Vec3{0, 0, -1})
	m.Triangle3v(v1, v2, v6, mgl32.Vec3{0, 0, -1})
	// left
	m.Triangle3v(v1, v5, v4, mgl32.Vec3{-1, 0, 0})
	m.Triangle3v(v1, v0, v4, mgl32.Vec3{-1, 0, 0})
	// right
	m.Triangle3v(v2, v6, v7, mgl32.Vec3{1, 0, 0})
	m.Triangle3v(v2, v3, v7, mgl32.Vec3{1, 0, 0})
	// bot
	m.Triangle3v(v0, v1, v2, mgl32.Vec3{0, -1, 0})
	m.Triangle3v(v0, v3, v2, mgl32.Vec3{0, -1, 0})
	// top
	m.Triangle3v(v4, v5, v6, mgl32.Vec3{0, 1, 0})
	m.Triangle3v(v4, v7, v6, mgl32.Vec3{0, 1, 0})
	// triangle mesh with normal
	return m
}

func GetCubeAsLinesV(v0, v1, v2, v3, v4, v5, v6, v7 mgl32.Vec3) *glumby.Mesh {
	m := glumby.NewMesh(gl.LINES)
	// front
	m.LinePair3v(v0, v4, v7)
	m.LinePair3v(v0, v3, v7)
	// back
	m.LinePair3v(v1, v5, v6)
	m.LinePair3v(v1, v2, v6)
	// left
	m.LinePair3v(v1, v5, v4)
	m.LinePair3v(v1, v0, v4)
	// right
	m.LinePair3v(v2, v6, v7)
	m.LinePair3v(v2, v3, v7)
	// bot
	m.LinePair3v(v0, v1, v2)
	m.LinePair3v(v0, v3, v2)
	// top
	m.LinePair3v(v4, v5, v6)
	m.LinePair3v(v4, v7, v6)
	// triangle mesh with normal
	return m
}
