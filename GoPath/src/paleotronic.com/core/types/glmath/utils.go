package glmath

func Scale(scaleX, scaleY, scaleZ float64) *Matrix4 {
	return &Matrix4{scaleX, 0, 0, 0, 0, scaleY, 0, 0, 0, 0, scaleZ, 0, 0, 0, 0, 1}
}

func Translate(x, y, z float64) *Matrix4 {
	return &Matrix4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, x, y, z, 1}
}

func TransformVector3(v *Vector3, m *Matrix4) *Vector3 {

	t := &Vector4{v.X(), v.Y(), v.Z(), 1}
	t = m.MulV4(t)
	t = t.MulF(1 / t[3])

	return &Vector3{t.X(), t.Y(), t.Z()}
}
