package glmath

import (
	"paleotronic.com/fmt"
	"math"
)

const M_PI = 3.141592654
const EPSILON = 0.00001
const RAD2DEG = 180 / 3.141592654
const DEG2RAD = M_PI / 180

type Matrix2 [4]float64
type Matrix3 [9]float64
type Matrix4 [16]float64

func aEquals(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, av := range a {
		if math.Abs(av-b[i]) >= EPSILON {
			return false
		}
	}
	return true
}

// Matrix2
// ---------------------------------------------------------------------------

func NewMatrix2(m0, m1, m2, m3 float64) *Matrix2 {
	return &Matrix2{
		m0, m1, m2, m3,
	}
}

func (m *Matrix2) swap(a, b int) {
	t := m[a]
	m[a] = m[b]
	m[b] = t
}

func (m *Matrix2) Transpose() *Matrix2 {
	m.swap(1, 2)
	return m
}

func (m Matrix2) getDeterminant() float64 {
	return m[0]*m[3] - m[1]*m[2]
}

func (m *Matrix2) Invert() *Matrix2 {
	determinant := m.getDeterminant()
	if math.Abs(determinant) <= EPSILON {
		return m.Identity()
	}

	tmp := m[0]
	invDeterminant := 1 / determinant
	m[0] = invDeterminant * m[3]
	m[1] = -invDeterminant * m[1]
	m[2] = -invDeterminant * m[2]
	m[3] = invDeterminant * tmp

	return m
}

func (m *Matrix2) Identity() *Matrix2 {
	m[0] = 1
	m[3] = 1
	m[1] = 0
	m[2] = 0
	return m
}

func (m Matrix2) Add(rhs *Matrix2) *Matrix2 {
	return &Matrix2{
		m[0] + rhs[0], m[1] + rhs[1], m[2] + rhs[2], m[3] + rhs[3],
	}
}

func (m Matrix2) Sub(rhs *Matrix2) *Matrix2 {
	return &Matrix2{
		m[0] - rhs[0], m[1] - rhs[1], m[2] - rhs[2], m[3] - rhs[3],
	}
}

func (m Matrix2) Mul(rhs *Matrix2) *Matrix2 {
	return &Matrix2{
		m[0]*rhs[0] + m[2]*rhs[1], m[1]*rhs[0] + m[3]*rhs[1],
		m[0]*rhs[2] + m[2]*rhs[3], m[1]*rhs[2] + m[3]*rhs[3],
	}
}

func (m Matrix2) Equals(o Matrix2) bool {
	return aEquals(m[:], o[:])
}

func (m Matrix2) GetAngle() float64 {
	return RAD2DEG * math.Atan2(m[1], m[0])
}

// Matrix3
// ---------------------------------------------------------------------------

func NewMatrix3(m0, m1, m2, m3, m4, m5, m6, m7, m8 float64) *Matrix3 {

	return &Matrix3{
		m0, m1, m2,
		m3, m4, m5,
		m6, m7, m8,
	}

}

func (m Matrix3) Equals(o Matrix3) bool {
	return aEquals(m[:], o[:])
}

func (m *Matrix3) swap(a, b int) {
	t := m[a]
	m[a] = m[b]
	m[b] = t
}

func (m *Matrix3) Transpose() *Matrix3 {
	m.swap(1, 3)
	m.swap(2, 6)
	m.swap(5, 7)
	return m
}

func (m Matrix3) getDeterminant() float64 {
	return m[0]*(m[4]*m[8]-m[5]*m[7]) -
		m[1]*(m[3]*m[8]-m[5]*m[6]) +
		m[2]*(m[3]*m[7]-m[4]*m[6])
}

func (m *Matrix3) Identity() *Matrix3 {
	m[0] = 1
	m[4] = 1
	m[8] = 1
	m[1] = 0
	m[2] = 0
	m[3] = 0
	m[5] = 0
	m[6] = 0
	m[7] = 0
	return m
}

func (m *Matrix3) Invert() *Matrix3 {
	var determinant, invDeterminant float64
	var tmp Matrix3

	tmp[0] = m[4]*m[8] - m[5]*m[7]
	tmp[1] = m[2]*m[7] - m[1]*m[8]
	tmp[2] = m[1]*m[5] - m[2]*m[4]
	tmp[3] = m[5]*m[6] - m[3]*m[8]
	tmp[4] = m[0]*m[8] - m[2]*m[6]
	tmp[5] = m[2]*m[3] - m[0]*m[5]
	tmp[6] = m[3]*m[7] - m[4]*m[6]
	tmp[7] = m[1]*m[6] - m[0]*m[7]
	tmp[8] = m[0]*m[4] - m[1]*m[3]

	// check determinant if it is 0
	determinant = m[0]*tmp[0] + m[1]*tmp[3] + m[2]*tmp[6]
	if math.Abs(determinant) <= EPSILON {
		return m.Identity() // cannot inverse, make it idenety matrix
	}

	// divide by the determinant
	invDeterminant = 1.0 / determinant
	m[0] = invDeterminant * tmp[0]
	m[1] = invDeterminant * tmp[1]
	m[2] = invDeterminant * tmp[2]
	m[3] = invDeterminant * tmp[3]
	m[4] = invDeterminant * tmp[4]
	m[5] = invDeterminant * tmp[5]
	m[6] = invDeterminant * tmp[6]
	m[7] = invDeterminant * tmp[7]
	m[8] = invDeterminant * tmp[8]

	return m
}

func (m Matrix3) Add(rhs Matrix3) *Matrix3 {
	return &Matrix3{
		m[0] + rhs[0], m[1] + rhs[1], m[2] + rhs[2],
		m[3] + rhs[3], m[4] + rhs[4], m[5] + rhs[5],
		m[6] + rhs[6], m[7] + rhs[7], m[8] + rhs[8],
	}
}

func (m Matrix3) Sub(rhs Matrix3) *Matrix3 {
	return &Matrix3{
		m[0] - rhs[0], m[1] - rhs[1], m[2] - rhs[2],
		m[3] - rhs[3], m[4] - rhs[4], m[5] - rhs[5],
		m[6] - rhs[6], m[7] - rhs[7], m[8] - rhs[8],
	}
}

func (m Matrix3) Mul(rhs Matrix3) *Matrix3 {
	return &Matrix3{
		m[0]*rhs[0] + m[3]*rhs[1] + m[6]*rhs[2], m[1]*rhs[0] + m[4]*rhs[1] + m[7]*rhs[2], m[2]*rhs[0] + m[5]*rhs[1] + m[8]*rhs[2],
		m[0]*rhs[3] + m[3]*rhs[4] + m[6]*rhs[5], m[1]*rhs[3] + m[4]*rhs[4] + m[7]*rhs[5], m[2]*rhs[3] + m[5]*rhs[4] + m[8]*rhs[5],
		m[0]*rhs[6] + m[3]*rhs[7] + m[6]*rhs[8], m[1]*rhs[6] + m[4]*rhs[7] + m[7]*rhs[8], m[2]*rhs[6] + m[5]*rhs[7] + m[8]*rhs[8],
	}
}

func (m Matrix3) MulVector3(rhs Vector3) *Vector3 {
	return NewVector3(
		m[0]*rhs.X()+m[3]*rhs.Y()+m[6]*rhs.Z(),
		m[1]*rhs.X()+m[4]*rhs.Y()+m[7]*rhs.Z(),
		m[2]*rhs.X()+m[5]*rhs.Y()+m[8]*rhs.Z(),
	)
}

func (m Matrix3) GetAngle() *Vector3 {
	var pitch, yaw, roll float64 // 3 angles

	// find yaw (around y-axis) first
	// NOTE: asin() returns -90~+90, so correct the angle range -180~+180
	// using z value of forward vector
	yaw = RAD2DEG * math.Asin(m[6])
	if m[8] < 0 {
		if yaw >= 0 {
			yaw = 180 - yaw
		} else {
			yaw = -180 - yaw
		}
	}

	// find roll (around z-axis) and pitch (around x-axis)
	// if forward vector is (1,0,0) or (-1,0,0), then m[0]=m[4]=m[9]=m[10]=0
	if m[0] > -EPSILON && m[0] < EPSILON {
		roll = 0 //@@ assume roll=0
		pitch = RAD2DEG * math.Atan2(m[1], m[4])
	} else {
		roll = RAD2DEG * math.Atan2(-m[3], m[0])
		pitch = RAD2DEG * math.Atan2(-m[7], m[8])
	}

	return NewVector3(pitch, yaw, roll)
}

// Matrix4
// ---------------------------------------------------------------------------

func NewMatrix4(m0, m1, m2, m3, m4, m5, m6, m7, m8, m9, m10, m11, m12, m13, m14, m15 float64) *Matrix4 {

	return &Matrix4{
		m0, m1, m2, m3,
		m4, m5, m6, m7,
		m8, m9, m10, m11,
		m12, m13, m14, m15,
	}

}

func (m Matrix4) Equals(o Matrix4) bool {
	return aEquals(m[:], o[:])
}

func (m *Matrix4) swap(a, b int) {
	t := m[a]
	m[a] = m[b]
	m[b] = t
}

func (m *Matrix4) Transpose() *Matrix4 {

	m.swap(1, 4)
	m.swap(2, 8)
	m.swap(3, 12)
	m.swap(6, 9)
	m.swap(7, 13)
	m.swap(11, 14)

	return m

}

func (m *Matrix4) Identity() *Matrix4 {
	m[0] = 1
	m[5] = 1
	m[10] = 1
	m[15] = 1
	m[1] = 0
	m[2] = 0
	m[3] = 0
	m[4] = 0
	m[6] = 0
	m[7] = 0
	m[8] = 0
	m[9] = 0
	m[11] = 0
	m[12] = 0
	m[13] = 0
	m[14] = 0

	return m
}

// Use InvertAffine() if the matrix has scale and shear transformation.
//
// M = [ R | T ]
//     [ --+-- ]    (R denotes 3x3 rotation/reflection matrix)
//     [ 0 | 1 ]    (T denotes 1x3 translation matrix)
//
// y = M*x  ->  y = R*x + T  ->  x = R^-1*(y - T)  ->  x = R^T*y - R^T*T
// (R is orthogonal,  R^-1 = R^T)
//
//  [ R | T ]-1    [ R^T | -R^T * T ]    (R denotes 3x3 rotation matrix)
//  [ --+-- ]   =  [ ----+--------- ]    (T denotes 1x3 translation)
//  [ 0 | 1 ]      [  0  |     1    ]    (R^T denotes R-transpose)
func (m *Matrix4) InvertEuclidean() *Matrix4 {
	var tmp float64
	tmp = m[1]
	m[1] = m[4]
	m[4] = tmp
	tmp = m[2]
	m[2] = m[8]
	m[8] = tmp
	tmp = m[6]
	m[6] = m[9]
	m[9] = tmp

	// compute translation part -R^T * T
	// | 0 | -R^T x |
	// | --+------- |
	// | 0 |   0    |
	x := m[12]
	y := m[13]
	z := m[14]
	m[12] = -(m[0]*x + m[4]*y + m[8]*z)
	m[13] = -(m[1]*x + m[5]*y + m[9]*z)
	m[14] = -(m[2]*x + m[6]*y + m[10]*z)

	// last row should be unchanged (0,0,0,1)
	return m
}

// compute the inverse of a 4x4 affine transformation matrix
//
// Affine transformations are generalizations of Euclidean transformations.
// Affine transformation includes translation, rotation, reflection, scaling,
// and shearing. Length and angle are NOT preserved.
// M = [ R | T ]
//     [ --+-- ]    (R denotes 3x3 rotation/scale/shear matrix)
//     [ 0 | 1 ]    (T denotes 1x3 translation matrix)
//
// y = M*x  ->  y = R*x + T  ->  x = R^-1*(y - T)  ->  x = R^-1*y - R^-1*T
//
//  [ R | T ]-1   [ R^-1 | -R^-1 * T ]
//  [ --+-- ]   = [ -----+---------- ]
//  [ 0 | 1 ]     [  0   +     1     ]
func (m *Matrix4) InvertAffine() *Matrix4 {
	// R^-1
	var r = &Matrix3{m[0], m[1], m[2], m[4], m[5], m[6], m[8], m[9], m[10]}
	r.Invert()

	m[0] = r[0]
	m[1] = r[1]
	m[2] = r[2]
	m[4] = r[3]
	m[5] = r[4]
	m[6] = r[5]
	m[8] = r[6]
	m[9] = r[7]
	m[10] = r[8]

	// -R^-1 * T
	x := m[12]
	y := m[13]
	z := m[14]
	m[12] = -(r[0]*x + r[3]*y + r[6]*z)
	m[13] = -(r[1]*x + r[4]*y + r[7]*z)
	m[14] = -(r[2]*x + r[5]*y + r[8]*z)

	// last row should be unchanged (0,0,0,1)
	//m[3] = m[7] = m[11] = 0.0f;
	//m[15] = 1.0f;

	return m
}

func (m *Matrix4) InvertProjective() *Matrix4 {
	// partition
	a := &Matrix2{m[0], m[1], m[4], m[5]}
	b := &Matrix2{m[8], m[9], m[12], m[13]}
	c := &Matrix2{m[2], m[3], m[6], m[7]}
	d := &Matrix2{m[10], m[11], m[14], m[15]}

	// pre-compute repeated parts
	a.Invert()         // A^-1
	ab := a.Mul(b)     // A^-1 * B
	ca := c.Mul(a)     // C * A^-1
	cab := ca.Mul(b)   // C * A^-1 * B
	dcab := d.Sub(cab) // D - C * A^-1 * B

	// check determinant if |D - C * A^-1 * B| = 0
	//NOTE: this function assumes det(A) is already checked. if |A|=0 then,
	//      cannot use this function.
	determinant := dcab[0]*dcab[3] - dcab[1]*dcab[2]
	if math.Abs(determinant) <= EPSILON {
		return m.Identity()
	}

	// compute D' and -D'
	d1 := dcab              //  (D - C * A^-1 * B)
	d1.Invert()             //  (D - C * A^-1 * B)^-1
	d2 := Matrix2{}.Sub(d1) // -(D - C * A^-1 * B)^-1

	// compute C'
	c1 := d2.Mul(ca) // -D' * (C * A^-1)

	// compute B'
	b1 := ab.Mul(d2) // (A^-1 * B) * -D'

	// compute A'
	a1 := a.Sub(ab.Mul(c1)) // A^-1 - (A^-1 * B) * C'

	// assemble inverse matrix
	m[0] = a1[0]
	m[4] = a1[2] /*|*/
	m[8] = b1[0]
	m[12] = b1[2]
	m[1] = a1[1]
	m[5] = a1[3] /*|*/
	m[9] = b1[1]
	m[13] = b1[3]
	/*-----------------------------+-----------------------------*/
	m[2] = c1[0]
	m[6] = c1[2] /*|*/
	m[10] = d1[0]
	m[14] = d1[2]
	m[3] = c1[1]
	m[7] = c1[3] /*|*/
	m[11] = d1[1]
	m[15] = d1[3]

	return m
}

func (m *Matrix4) getCofactor(m0, m1, m2, m3, m4, m5, m6, m7, m8 float64) float64 {
	return m0*(m4*m8-m5*m7) -
		m1*(m3*m8-m5*m6) +
		m2*(m3*m7-m4*m6)
}

func (m *Matrix4) getDeterminant() float64 {
	return m[0]*m.getCofactor(m[5], m[6], m[7], m[9], m[10], m[11], m[13], m[14], m[15]) -
		m[1]*m.getCofactor(m[4], m[6], m[7], m[8], m[10], m[11], m[12], m[14], m[15]) +
		m[2]*m.getCofactor(m[4], m[5], m[7], m[8], m[9], m[11], m[12], m[13], m[15]) -
		m[3]*m.getCofactor(m[4], m[5], m[6], m[8], m[9], m[10], m[12], m[13], m[14])
}

func (m *Matrix4) InvertGeneral() *Matrix4 {
	// get cofactors of minor matrices
	cofactor0 := m.getCofactor(m[5], m[6], m[7], m[9], m[10], m[11], m[13], m[14], m[15])
	cofactor1 := m.getCofactor(m[4], m[6], m[7], m[8], m[10], m[11], m[12], m[14], m[15])
	cofactor2 := m.getCofactor(m[4], m[5], m[7], m[8], m[9], m[11], m[12], m[13], m[15])
	cofactor3 := m.getCofactor(m[4], m[5], m[6], m[8], m[9], m[10], m[12], m[13], m[14])

	// get determinant
	determinant := m[0]*cofactor0 - m[1]*cofactor1 + m[2]*cofactor2 - m[3]*cofactor3
	if math.Abs(determinant) <= EPSILON {
		return m.Identity()
	}

	// get rest of cofactors for adj(M)
	cofactor4 := m.getCofactor(m[1], m[2], m[3], m[9], m[10], m[11], m[13], m[14], m[15])
	cofactor5 := m.getCofactor(m[0], m[2], m[3], m[8], m[10], m[11], m[12], m[14], m[15])
	cofactor6 := m.getCofactor(m[0], m[1], m[3], m[8], m[9], m[11], m[12], m[13], m[15])
	cofactor7 := m.getCofactor(m[0], m[1], m[2], m[8], m[9], m[10], m[12], m[13], m[14])

	cofactor8 := m.getCofactor(m[1], m[2], m[3], m[5], m[6], m[7], m[13], m[14], m[15])
	cofactor9 := m.getCofactor(m[0], m[2], m[3], m[4], m[6], m[7], m[12], m[14], m[15])
	cofactor10 := m.getCofactor(m[0], m[1], m[3], m[4], m[5], m[7], m[12], m[13], m[15])
	cofactor11 := m.getCofactor(m[0], m[1], m[2], m[4], m[5], m[6], m[12], m[13], m[14])

	cofactor12 := m.getCofactor(m[1], m[2], m[3], m[5], m[6], m[7], m[9], m[10], m[11])
	cofactor13 := m.getCofactor(m[0], m[2], m[3], m[4], m[6], m[7], m[8], m[10], m[11])
	cofactor14 := m.getCofactor(m[0], m[1], m[3], m[4], m[5], m[7], m[8], m[9], m[11])
	cofactor15 := m.getCofactor(m[0], m[1], m[2], m[4], m[5], m[6], m[8], m[9], m[10])

	// build inverse matrix = adj(M) / det(M)
	// adjugate of M is the transpose of the cofactor matrix of M
	invDeterminant := 1 / determinant
	m[0] = invDeterminant * cofactor0
	m[1] = -invDeterminant * cofactor4
	m[2] = invDeterminant * cofactor8
	m[3] = -invDeterminant * cofactor12

	m[4] = -invDeterminant * cofactor1
	m[5] = invDeterminant * cofactor5
	m[6] = -invDeterminant * cofactor9
	m[7] = invDeterminant * cofactor13

	m[8] = invDeterminant * cofactor2
	m[9] = -invDeterminant * cofactor6
	m[10] = invDeterminant * cofactor10
	m[11] = -invDeterminant * cofactor14

	m[12] = -invDeterminant * cofactor3
	m[13] = invDeterminant * cofactor7
	m[14] = -invDeterminant * cofactor11
	m[15] = invDeterminant * cofactor15

	return m

}

func (m *Matrix4) Translate(x, y, z float64) *Matrix4 {
	m[0] += m[3] * x
	m[4] += m[7] * x
	m[8] += m[11] * x
	m[12] += m[15] * x
	m[1] += m[3] * y
	m[5] += m[7] * y
	m[9] += m[11] * y
	m[13] += m[15] * y
	m[2] += m[3] * z
	m[6] += m[7] * z
	m[10] += m[11] * z
	m[14] += m[15] * z

	return m
}

func (m *Matrix4) TranslateV(v *Vector3) *Matrix4 {
	return m.Translate(v[0], v[1], v[2])
}

func (m *Matrix4) Scale(x, y, z float64) *Matrix4 {
	m[0] *= x
	m[4] *= x
	m[8] *= x
	m[12] *= x
	m[1] *= y
	m[5] *= y
	m[9] *= y
	m[13] *= y
	m[2] *= z
	m[6] *= z
	m[10] *= z
	m[14] *= z
	return m
}

func (m *Matrix4) ScaleV(v *Vector3) *Matrix4 {
	return m.Scale(v[0], v[1], v[2])
}

func (m *Matrix4) ScaleF(f float64) *Matrix4 {
	return m.ScaleV(&Vector3{f, f, f})
}

func (m *Matrix4) RotateV(angle float64, axis *Vector3) *Matrix4 {
	return m.Rotate(angle, axis[0], axis[1], axis[2])
}

func (m *Matrix4) Rotate(angle float64, x, y, z float64) *Matrix4 {
	c := math.Cos(angle * DEG2RAD) // cosine
	s := math.Sin(angle * DEG2RAD) // sine
	c1 := 1.0 - c                  // 1 - c

	m0 := m[0]
	m4 := m[4]
	m8 := m[8]
	m12 := m[12]
	m1 := m[1]
	m5 := m[5]
	m9 := m[9]
	m13 := m[13]
	m2 := m[2]
	m6 := m[6]
	m10 := m[10]
	m14 := m[14]

	// build rotation matrix
	r0 := x*x*c1 + c
	r1 := x*y*c1 + z*s
	r2 := x*z*c1 - y*s
	r4 := x*y*c1 - z*s
	r5 := y*y*c1 + c
	r6 := y*z*c1 + x*s
	r8 := x*z*c1 + y*s
	r9 := y*z*c1 - x*s
	r10 := z*z*c1 + c

	// multiply rotation matrix
	m[0] = r0*m0 + r4*m1 + r8*m2
	m[1] = r1*m0 + r5*m1 + r9*m2
	m[2] = r2*m0 + r6*m1 + r10*m2
	m[4] = r0*m4 + r4*m5 + r8*m6
	m[5] = r1*m4 + r5*m5 + r9*m6
	m[6] = r2*m4 + r6*m5 + r10*m6
	m[8] = r0*m8 + r4*m9 + r8*m10
	m[9] = r1*m8 + r5*m9 + r9*m10
	m[10] = r2*m8 + r6*m9 + r10*m10
	m[12] = r0*m12 + r4*m13 + r8*m14
	m[13] = r1*m12 + r5*m13 + r9*m14
	m[14] = r2*m12 + r6*m13 + r10*m14

	return m
}

func (m *Matrix4) RotateX(angle float64) *Matrix4 {
	c := math.Cos(angle * DEG2RAD) // cosine
	s := math.Sin(angle * DEG2RAD) // sine

	m1 := m[1]
	m2 := m[2]
	m5 := m[5]
	m6 := m[6]
	m9 := m[9]
	m10 := m[10]
	m13 := m[13]
	m14 := m[14]

	m[1] = m1*c + m2*-s
	m[2] = m1*s + m2*c
	m[5] = m5*c + m6*-s
	m[6] = m5*s + m6*c
	m[9] = m9*c + m10*-s
	m[10] = m9*s + m10*c
	m[13] = m13*c + m14*-s
	m[14] = m13*s + m14*c

	return m
}

func (m *Matrix4) RotateZ(angle float64) *Matrix4 {
	c := math.Cos(angle * DEG2RAD) // cosine
	s := math.Sin(angle * DEG2RAD) // sine

	m0 := m[0]
	m1 := m[1]
	m4 := m[4]
	m5 := m[5]
	m8 := m[8]
	m9 := m[9]
	m12 := m[12]
	m13 := m[13]

	m[0] = m0*c + m1*-s
	m[1] = m0*s + m1*c
	m[4] = m4*c + m5*-s
	m[5] = m4*s + m5*c
	m[8] = m8*c + m9*-s
	m[9] = m8*s + m9*c
	m[12] = m12*c + m13*-s
	m[13] = m12*s + m13*c

	return m
}

func (m *Matrix4) RotateY(angle float64) *Matrix4 {
	c := math.Cos(angle * DEG2RAD) // cosine
	s := math.Sin(angle * DEG2RAD) // sine

	m0 := m[0]
	m2 := m[2]
	m4 := m[4]
	m6 := m[6]
	m8 := m[8]
	m10 := m[10]
	m12 := m[12]
	m14 := m[14]

	m[0] = m0*c + m2*s
	m[2] = m0*-s + m2*c
	m[4] = m4*c + m6*s
	m[6] = m4*-s + m6*c
	m[8] = m8*c + m10*s
	m[10] = m8*-s + m10*c
	m[12] = m12*c + m14*s
	m[14] = m12*-s + m14*c

	return m
}

func (m *Matrix4) SetColumnV3(index int, v *Vector3) *Matrix4 {
	m[index*4] = v.X()
	m[index*4+1] = v.Y()
	m[index*4+2] = v.Z()
	return m
}

func (m *Matrix4) SetRowV3(index int, v *Vector3) *Matrix4 {
	m[index] = v.X()
	m[index+4] = v.Y()
	m[index+8] = v.Z()
	return m
}

func (m *Matrix4) SetColumnV4(index int, v *Vector4) *Matrix4 {
	m[index*4] = v.X()
	m[index*4+1] = v.Y()
	m[index*4+2] = v.Z()
	m[index*4+3] = v.W()
	return m
}

func (m *Matrix4) SetRowV4(index int, v *Vector4) *Matrix4 {
	m[index] = v.X()
	m[index+4] = v.Y()
	m[index+8] = v.Z()
	m[index+12] = v.W()
	return m
}

func (m *Matrix4) SetColumn4(index int, col [4]float64) *Matrix4 {
	m[index*4] = col[0]
	m[index*4+1] = col[1]
	m[index*4+2] = col[2]
	m[index*4+3] = col[3]
	return m
}

func (m *Matrix4) SetRow4(index int, row [4]float64) *Matrix4 {
	m[index] = row[0]
	m[index+4] = row[1]
	m[index+8] = row[2]
	m[index+12] = row[3]
	return m
}

func (m *Matrix4) LookAt(target *Vector3) *Matrix4 {

	// compute forward vector and normalize
	position := NewVector3(m[12], m[13], m[14])
	forward := target.Sub(position)
	forward.Normalize()
	up := &Vector3{}   // up vector of object
	left := &Vector3{} // left vector of object

	// compute temporal up vector
	// if forward vector is near Y-axis, use up vector (0,0,-1) or (0,0,1)
	if math.Abs(forward.X()) < EPSILON && math.Abs(forward.Z()) < EPSILON {
		// forward vector is pointing +Y axis
		if forward.Y() > 0 {
			up.Set(0, 0, -1)
			// forward vector is pointing -Y axis
		} else {
			up.Set(0, 0, 1)
		}
	} else {
		// assume up vector is +Y axis
		up.Set(0, 1, 0)
	}

	// compute left vector
	left = up.Cross(forward)
	left.Normalize()

	// re-compute up vector
	up = forward.Cross(left)
	//up.normalize();

	// NOTE: overwrite rotation and scale info of the current matrix
	m.SetColumnV3(0, left)
	m.SetColumnV3(1, up)
	m.SetColumnV3(2, forward)

	return m
}

func (m *Matrix4) LookAtWithUp(target *Vector3, upVec *Vector3) *Matrix4 {

	// compute forward vector and normalize
	position := NewVector3(m[12], m[13], m[14])
	forward := target.Sub(position)
	forward.Normalize()

	// compute left vector
	left := upVec.Cross(forward)
	left.Normalize()

	// compute orthonormal up vector
	up := forward.Cross(left)
	up.Normalize()

	// NOTE: overwrite rotation and scale info of the current matrix
	m.SetColumnV3(0, left)
	m.SetColumnV3(1, up)
	m.SetColumnV3(2, forward)

	return m
}

func (m *Matrix4) LookAtF(tx, ty, tz float64) *Matrix4 {
	return m.LookAt(&Vector3{tx, ty, tz})
}

func (m *Matrix4) LookAtWithUpF(tx, ty, tz float64, ux, uy, uz float64) *Matrix4 {
	return m.LookAtWithUp(&Vector3{tx, ty, tz}, &Vector3{ux, uy, uz})
}

func (m *Matrix4) GetAngle() *Vector3 {
	var pitch, yaw, roll float64 // 3 angles

	// find yaw (around y-axis) first
	// NOTE: asin() returns -90~+90, so correct the angle range -180~+180
	// using z value of forward vector
	yaw = RAD2DEG * math.Asin(m[8])
	if m[10] < 0 {
		if yaw >= 0 {
			yaw = 180.0 - yaw
		} else {
			yaw = -180.0 - yaw
		}
	}

	// find roll (around z-axis) and pitch (around x-axis)
	// if forward vector is (1,0,0) or (-1,0,0), then m[0]=m[4]=m[9]=m[10]=0
	if m[0] > -EPSILON && m[0] < EPSILON {
		roll = 0 //@@ assume roll=0
		pitch = RAD2DEG * math.Atan2(m[1], m[5])
	} else {
		roll = RAD2DEG * math.Atan2(-m[4], m[0])
		pitch = RAD2DEG * math.Atan2(-m[9], m[10])
	}

	return NewVector3(pitch, yaw, roll)
}

func (m *Matrix4) String() string {

	out := ""
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			out += fmt.Sprintf(" %.6f", m[c*4+r])
		}
		out += "\r\n"
	}
	return out

}

func (matrix *Matrix4) MulV4(pvector *Vector4) *Vector4 {
	resultvector := &Vector4{}
	resultvector[0] = matrix[0]*pvector[0] + matrix[4]*pvector[1] + matrix[8]*pvector[2] + matrix[12]*pvector[3]
	resultvector[1] = matrix[1]*pvector[0] + matrix[5]*pvector[1] + matrix[9]*pvector[2] + matrix[13]*pvector[3]
	resultvector[2] = matrix[2]*pvector[0] + matrix[6]*pvector[1] + matrix[10]*pvector[2] + matrix[14]*pvector[3]
	resultvector[3] = matrix[3]*pvector[0] + matrix[7]*pvector[1] + matrix[11]*pvector[2] + matrix[15]*pvector[3]
	return resultvector
}

func (matrix1 *Matrix4) Mul(matrix2 *Matrix4) *Matrix4 {

	result := &Matrix4{}
	result[0] = matrix1[0]*matrix2[0] +
		matrix1[4]*matrix2[1] +
		matrix1[8]*matrix2[2] +
		matrix1[12]*matrix2[3]
	result[4] = matrix1[0]*matrix2[4] +
		matrix1[4]*matrix2[5] +
		matrix1[8]*matrix2[6] +
		matrix1[12]*matrix2[7]
	result[8] = matrix1[0]*matrix2[8] +
		matrix1[4]*matrix2[9] +
		matrix1[8]*matrix2[10] +
		matrix1[12]*matrix2[11]
	result[12] = matrix1[0]*matrix2[12] +
		matrix1[4]*matrix2[13] +
		matrix1[8]*matrix2[14] +
		matrix1[12]*matrix2[15]
	result[1] = matrix1[1]*matrix2[0] +
		matrix1[5]*matrix2[1] +
		matrix1[9]*matrix2[2] +
		matrix1[13]*matrix2[3]
	result[5] = matrix1[1]*matrix2[4] +
		matrix1[5]*matrix2[5] +
		matrix1[9]*matrix2[6] +
		matrix1[13]*matrix2[7]
	result[9] = matrix1[1]*matrix2[8] +
		matrix1[5]*matrix2[9] +
		matrix1[9]*matrix2[10] +
		matrix1[13]*matrix2[11]
	result[13] = matrix1[1]*matrix2[12] +
		matrix1[5]*matrix2[13] +
		matrix1[9]*matrix2[14] +
		matrix1[13]*matrix2[15]
	result[2] = matrix1[2]*matrix2[0] +
		matrix1[6]*matrix2[1] +
		matrix1[10]*matrix2[2] +
		matrix1[14]*matrix2[3]
	result[6] = matrix1[2]*matrix2[4] +
		matrix1[6]*matrix2[5] +
		matrix1[10]*matrix2[6] +
		matrix1[14]*matrix2[7]
	result[10] = matrix1[2]*matrix2[8] +
		matrix1[6]*matrix2[9] +
		matrix1[10]*matrix2[10] +
		matrix1[14]*matrix2[11]
	result[14] = matrix1[2]*matrix2[12] +
		matrix1[6]*matrix2[13] +
		matrix1[10]*matrix2[14] +
		matrix1[14]*matrix2[15]
	result[3] = matrix1[3]*matrix2[0] +
		matrix1[7]*matrix2[1] +
		matrix1[11]*matrix2[2] +
		matrix1[15]*matrix2[3]
	result[7] = matrix1[3]*matrix2[4] +
		matrix1[7]*matrix2[5] +
		matrix1[11]*matrix2[6] +
		matrix1[15]*matrix2[7]
	result[11] = matrix1[3]*matrix2[8] +
		matrix1[7]*matrix2[9] +
		matrix1[11]*matrix2[10] +
		matrix1[15]*matrix2[11]
	result[15] = matrix1[3]*matrix2[12] +
		matrix1[7]*matrix2[13] +
		matrix1[11]*matrix2[14] +
		matrix1[15]*matrix2[15]

	return result
}

func (m *Matrix4) ToGLFloat() [16]float32 {

	var out [16]float32
	for i, v := range m {
		out[i] = float32(v)
	}
	return out

}
