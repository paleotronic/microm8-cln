package glumby

import (
	"paleotronic.com/gl"
	"math"
)

type Camera struct {
	Position                     Vector3
	ViewDir                      Vector3
	UP                           Vector3
	ViewDirChanged               bool
	RotatedX, RotatedY, RotatedZ float32
	Near, Far                    float64
	Frustrum                     Rect
	Perspective                  bool
}

func AddVector(u *Vector3, v *Vector3) Vector3 {
	var result Vector3
	result.X = u.X + v.X
	result.Y = u.Y + v.Y
	result.Z = u.Z + v.Z
	return result
}

func AddVectorToVector(dst *Vector3, v2 *Vector3) {
	dst.X += v2.X
	dst.Y += v2.Y
	dst.Z += v2.Z
}

// NewCamera creates a new Camera object
func NewCamera(bounds Rect, near, far float64, perspective bool) *Camera {
	this := &Camera{}

	this.Position = Vector3{0, 0, 0}
	this.ViewDir = Vector3{0, 0, -1} // View dir should always be a unit vector
	this.UP = Vector3{0, 1, 0}
	this.ViewDirChanged = false
	this.RotatedX = 0
	this.RotatedY = 0
	this.RotatedZ = 0
	this.Frustrum = bounds
	this.Near = near
	this.Far = far
	this.Perspective = perspective

	this.Apply()

	return this
}

func (c *Camera) SetPos(x, y, z float32) {
	c.Position.X = x
	c.Position.Y = y
	c.Position.Z = z
}

func (c *Camera) Apply() {

	if c.ViewDirChanged {
		c.GetViewDir()
	}

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	if c.Perspective {
		gl.Frustum(c.Frustrum.Left, c.Frustrum.Right, c.Frustrum.Bottom, c.Frustrum.Top, c.Near, c.Far)
	} else {
		gl.Ortho(c.Frustrum.Left, c.Frustrum.Right, c.Frustrum.Bottom, c.Frustrum.Top, c.Near, c.Far)
	}
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Rotatef(-c.RotatedX, 1.0, 0.0, 0.0)
	gl.Rotatef(-c.RotatedY, 0.0, 1.0, 0.0)
	gl.Rotatef(-c.RotatedZ, 0.0, 0.0, 1.0)

	gl.Translatef(-c.Position.X, -c.Position.Y, -c.Position.Z)

}

func (c *Camera) RotateX(angle float32) {
	c.RotatedX += angle
	c.ViewDirChanged = true
}

func (c *Camera) RotateY(angle float32) {
	c.RotatedY += angle
	c.ViewDirChanged = true
}

func (c *Camera) RotateZ(angle float32) {
	c.RotatedZ += angle
	c.ViewDirChanged = true
}

func (c *Camera) GetViewDir() {
	var s1, s2 Vector3
	// Y rotation
	s1.X = float32(math.Cos(float64((c.RotatedY + 90) * (math.Pi / 180))))
	s1.Z = float32(-math.Sin(float64((c.RotatedY + 90) * (math.Pi / 180))))
	// X rotation
	cosX := math.Cos(float64(c.RotatedX * (math.Pi / 180)))
	s2.X = s1.X * float32(cosX)
	s2.Z = s1.Z * float32(cosX)
	s2.Y = float32(math.Sin(float64(c.RotatedX * (math.Pi / 180))))
	// Result
	c.ViewDir = s2
}

func (c *Camera) Forward(distance float32) {
	if c.ViewDirChanged {
		c.GetViewDir()
	}
	var MoveVector Vector3
	MoveVector.X = c.ViewDir.X * -distance
	MoveVector.Y = c.ViewDir.Y * -distance
	MoveVector.Z = c.ViewDir.Z * -distance
	AddVectorToVector(&c.Position, &MoveVector)
}

func (c *Camera) Backward(distance float32) {
	if c.ViewDirChanged {
		c.GetViewDir()
	}
	var MoveVector Vector3
	MoveVector.X = c.ViewDir.X * distance
	MoveVector.Y = c.ViewDir.Y * distance
	MoveVector.Z = c.ViewDir.Z * distance
	AddVectorToVector(&c.Position, &MoveVector)
}
