package glmath

import "math"

type Quaternion [4]float64

const (
	qS = 0
	qX = 1
	qY = 2
	qZ = 3
)

func (q *Quaternion) Set(s, x, y, z float64) {
	q[qS] = s
	q[qX] = x
	q[qY] = y
	q[qZ] = z
}

func (q *Quaternion) S() float64 {
	return q[qS]
}

func (q *Quaternion) X() float64 {
	return q[qX]
}

func (q *Quaternion) Y() float64 {
	return q[qY]
}

func (q *Quaternion) Z() float64 {
	return q[qZ]
}

func (q *Quaternion) SetS(v float64) {
	q[qS] = v
}

func (q *Quaternion) SetX(v float64) {
	q[qX] = v
}

func (q *Quaternion) SetY(v float64) {
	q[qY] = v
}

func (q *Quaternion) SetZ(v float64) {
	q[qZ] = v
}

func (q *Quaternion) Length() float64 {
	return math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])
}

func (q *Quaternion) Normalize() *Quaternion {
	d := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]
	if d < EPSILON {
		return q
	}
	invLength := 1 / math.Sqrt(d)
	q[qS] *= invLength
	q[qX] *= invLength
	q[qY] *= invLength
	q[qZ] *= invLength
	return q
}

func (q *Quaternion) Conjugate() *Quaternion {
	q[qX] = -q[qX]
	q[qY] = -q[qY]
	q[qZ] = -q[qZ]
	return q
}

func (q *Quaternion) MulF(a float64) *Quaternion {
	return &Quaternion{
		q[0] * a,
		q[1] * a,
		q[2] * a,
		q[3] * a,
	}
}

func (q *Quaternion) Invert() *Quaternion {
	d := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]
	if d < EPSILON {
		return q
	}
	q = q.Conjugate().MulF(1 / d)
	return q
}

func (q *Quaternion) GetMatrix() *Matrix4 {

	x2 := q[qX] + q[qX]
	y2 := q[qY] + q[qY]
	z2 := q[qZ] + q[qZ]
	xx2 := q[qX] * x2
	xy2 := q[qX] * y2
	xz2 := q[qX] * z2
	yy2 := q[qY] * y2
	yz2 := q[qY] * z2
	zz2 := q[qZ] * z2
	sx2 := q[qS] * x2
	sy2 := q[qS] * y2
	sz2 := q[qS] * z2

	// build 4x4 matrix (column-major) and return
	return &Matrix4{1 - (yy2 + zz2), xy2 + sz2, xz2 - sy2, 0,
		xy2 - sz2, 1 - (xx2 + zz2), yz2 + sx2, 0,
		xz2 + sy2, yz2 - sx2, 1 - (xx2 + yy2), 0,
		0, 0, 0, 1}

}

func GetQuaternion(angles *Vector3) *Quaternion {
	qx := NewQuaternionAA(&Vector3{1, 0, 0}, angles.X()) // rotate along X
	qy := NewQuaternionAA(&Vector3{0, 1, 0}, angles.Y()) // rotate along Y
	qz := NewQuaternionAA(&Vector3{0, 0, 1}, angles.Z()) // rotate along Z
	return qx.Mul(qy).Mul(qz)                            // order: z->y->x
}

func (q *Quaternion) Mul(rhs *Quaternion) *Quaternion {
	v1 := NewVector3(q.X(), q.Y(), q.Z())
	v2 := NewVector3(rhs.X(), rhs.Y(), rhs.Z())

	cross := v1.Cross(v2)                                 // v x v'
	dot := v1.Dot(v2)                                     // v . v'
	v3 := cross.Add(v2.MulF(q.S())).Add(v1.MulF(rhs.S())) // v x v' + sv' + s'v

	return &Quaternion{q.S()*rhs.S() - dot, v3.X(), v3.Y(), v3.Z()}
}

func NewQuaternionAA(v *Vector3, angle float64) *Quaternion {

	return &Quaternion{
		angle,
		v.X(),
		v.Y(),
		v.Z(),
	}

}
