package glmath

import (
	"errors"
	"math"
)

func Perspective(fovy, aspect, near, far float64) *Matrix4 {
	// fovy = (fovy * math.Pi) / 180.0 // convert from degrees to radians

	nmf, f := near-far, float64(1./math.Tan(float64(fovy*DEG2RAD)/2.0))

	return &Matrix4{
		float64(f / aspect), 0, 0, 0,
		0, float64(f), 0, 0,
		0, 0, float64((near + far) / nmf), -1,
		0, 0, float64((2. * far * near) / nmf), 0}
}

func Frustrum(left, right, bottom, top, znear, zfar float64) *Matrix4 {

	var temp, temp2, temp3, temp4 float64
	temp = 2.0 * znear
	temp2 = right - left
	temp3 = top - bottom
	temp4 = zfar - znear

	matrix := &Matrix4{}

	matrix[0] = temp / temp2
	matrix[1] = 0.0
	matrix[2] = 0.0
	matrix[3] = 0.0
	matrix[4] = 0.0
	matrix[5] = temp / temp3
	matrix[6] = 0.0
	matrix[7] = 0.0
	matrix[8] = (right + left) / temp2
	matrix[9] = (top + bottom) / temp3
	matrix[10] = (-zfar - znear) / temp4
	matrix[11] = -1.0
	matrix[12] = 0.0
	matrix[13] = 0.0
	matrix[14] = (-temp * zfar) / temp4
	matrix[15] = 0.0

	return matrix
}

func UnProject(win *Vector3, modelview, projection *Matrix4, initialX, initialY, width, height int) (obj *Vector3, err error) {
	inv := projection.Mul(modelview).InvertProjective()
	blank := &Matrix4{}
	if inv == blank {
		return &Vector3{}, errors.New("Could not find matrix inverse (projection times modelview is probably non-singular)")
	}

	obj4 := inv.MulV4(&Vector4{
		(2 * (win[0] - float64(initialX)) / float64(width)) - 1,
		(2 * (win[1] - float64(initialY)) / float64(height)) - 1,
		2*win[2] - 1,
		1.0,
	})
	obj = &Vector3{obj4.X(), obj4.Y(), obj4.Z()}

	//if obj4[3] > MinValue {}
	obj[0] /= obj4[3]
	obj[1] /= obj4[3]
	obj[2] /= obj4[3]

	return obj, nil
}
