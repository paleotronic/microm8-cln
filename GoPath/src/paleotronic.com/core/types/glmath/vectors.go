package glmath

import "math"

// Vector3 is a 3D point
type Vector3 [3]float64

// NewVector3 creates a new Vector3
func NewVector3(x, y, z float64) *Vector3 {
	return &Vector3{x, y, z}
}

// Vec3 creates a new Vector3
func Vec3(x, y, z float64) *Vector3 {
	return NewVector3(x, y, z)
}

// X returns the x component of v
func (v *Vector3) X() float64 {
	return v[0]
}

// Y returns the y component of v
func (v *Vector3) Y() float64 {
	return v[1]
}

// Z returns the z component of v
func (v *Vector3) Z() float64 {
	return v[2]
}

// SetX sets the x component
func (v *Vector3) SetX(val float64) *Vector3 {
	v[0] = val
	return v
}

// SetY sets the y component
func (v *Vector3) SetY(val float64) *Vector3 {
	v[1] = val
	return v
}

// SetZ sets the z component
func (v *Vector3) SetZ(val float64) *Vector3 {
	v[2] = val
	return v
}

func (v *Vector3) Set(x, y, z float64) *Vector3 {
	v[0], v[1], v[2] = x, y, z
	return v
}

// Len returns the length
func (v *Vector3) Len() float64 {

	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])

}

// Normal returns a normal (unit vector)
func (v *Vector3) Normalize() *Vector3 {

	l := 1 / v.Len()

	v[0] *= l
	v[1] *= l
	v[2] *= l

	return v

}

func (v *Vector3) Dot(rhs *Vector3) float64 {
	return v[0]*rhs[0] + v[1]*rhs[1] + v[2]*rhs[2]
}

func (v *Vector3) Cross(rhs *Vector3) *Vector3 {

	return &Vector3{
		v[1]*rhs[2] - v[2]*rhs[1],
		v[2]*rhs[0] - v[0]*rhs[2],
		v[0]*rhs[1] - v[1]*rhs[0],
	}

}

// MulF multiples the vector components by an amount
func (v *Vector3) MulF(l float64) *Vector3 {
	return NewVector3(
		v[0]*l,
		v[1]*l,
		v[2]*l,
	)
}

func (v *Vector3) Mul(rhs *Vector3) *Vector3 {
	return &Vector3{
		v[0] * rhs[0],
		v[1] * rhs[1],
		v[2] * rhs[2],
	}
}

func (v *Vector3) Add(rhs *Vector3) *Vector3 {
	return &Vector3{
		v[0] + rhs[0],
		v[1] + rhs[1],
		v[2] + rhs[2],
	}
}

func (v *Vector3) Sub(rhs *Vector3) *Vector3 {
	return &Vector3{
		v[0] - rhs[0],
		v[1] - rhs[1],
		v[2] - rhs[2],
	}
}

func (m Vector3) Equals(o Vector3) bool {
	return aEquals(m[:], o[:])
}

// Vector3 is a 3D point
type Vector4 [4]float64

// NewVector3 creates a new Vector3
func NewVector4(x, y, z, w float64) *Vector4 {
	return &Vector4{x, y, z, w}
}

// X returns the x component of v
func (v *Vector4) X() float64 {
	return v[0]
}

// Y returns the y component of v
func (v *Vector4) Y() float64 {
	return v[1]
}

// Z returns the z component of v
func (v *Vector4) Z() float64 {
	return v[2]
}

// Z returns the w component of v
func (v *Vector4) W() float64 {
	return v[3]
}

// SetX sets the x component
func (v *Vector4) SetX(val float64) *Vector4 {
	v[0] = val
	return v
}

// SetY sets the y component
func (v *Vector4) SetY(val float64) *Vector4 {
	v[1] = val
	return v
}

// SetZ sets the z component
func (v *Vector4) SetZ(val float64) *Vector4 {
	v[2] = val
	return v
}

// SetW sets the w component
func (v *Vector4) SetW(val float64) *Vector4 {
	v[3] = val
	return v
}

// Len returns the length
func (v *Vector4) Len() float64 {

	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])

}

// Normal returns a normal (unit vector)
func (v *Vector4) Normal() *Vector4 {

	l := v.Len()
	return NewVector4(
		v[0]/l,
		v[1]/l,
		v[2]/l,
		v[3]/l,
	)

}

// MulF multiples the vector components by an amount
func (v *Vector4) MulF(l float64) *Vector4 {
	return NewVector4(
		v[0]*l,
		v[1]*l,
		v[2]*l,
		v[3]*l,
	)
}
