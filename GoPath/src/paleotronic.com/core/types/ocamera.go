package types

import (
	"math"

	"paleotronic.com/fmt"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types/glmath"
)

const (
	OCD_POSITION = 0
	OCD_TARGET   = 3
	OCD_ANGLE    = 6

	OCD_DISTANCE = 9

	OCD_FOV    = 10
	OCD_ASPECT = 11
	OCD_NEAR   = 12
	OCD_FAR    = 13

	OCD_LOCKVIEW = 14
	OCD_DIRTY    = 15

	OCD_DEFAULT_POS      = 16
	OCD_DEFAULT_TARGET   = 19
	OCD_DEFAULT_ANGLE    = 22
	OCD_DEFAULT_DISTANCE = 25

	OCD_DPOS  = 26
	OCD_DVIEW = 27

	OCD_ZOOM = 28

	OCD_QUAT = 29 // 4

	OCD_DOLLYRATE   = 33
	OCD_SHAKEFRAMES = 34
	OCD_SHAKEMAX    = 35

	OCD_MATRIX = 36

	OCD_MATRIX_ROT = 52

	OCD_PAN_X = 68
	OCD_PAN_Y = 69
	OCD_SIZE  = 70
)

func memUnpackVector3(mm *memory.MemoryControlBlock, addr int) *glmath.Vector3 {

	return &glmath.Vector3{
		Uint2Float64(mm.Read(addr + 0)),
		Uint2Float64(mm.Read(addr + 1)),
		Uint2Float64(mm.Read(addr + 2)),
	}

}

func memPackVector3(mm *memory.MemoryControlBlock, addr int, v *glmath.Vector3) {
	mm.Write(addr+0, Float642uint(v[0]))
	mm.Write(addr+1, Float642uint(v[1]))
	mm.Write(addr+2, Float642uint(v[2]))
}

func memUnpackQuat(mm *memory.MemoryControlBlock, addr int) *glmath.Quaternion {

	return &glmath.Quaternion{
		Uint2Float64(mm.Read(addr + 0)),
		Uint2Float64(mm.Read(addr + 1)),
		Uint2Float64(mm.Read(addr + 2)),
		Uint2Float64(mm.Read(addr + 3)),
	}

}

func memPackQuat(mm *memory.MemoryControlBlock, addr int, v *glmath.Quaternion) {
	mm.Write(addr+0, Float642uint(v[0]))
	mm.Write(addr+1, Float642uint(v[1]))
	mm.Write(addr+2, Float642uint(v[2]))
	mm.Write(addr+3, Float642uint(v[3]))
}

func memUnpackMatrix4(mm *memory.MemoryControlBlock, addr int) *glmath.Matrix4 {

	m := &glmath.Matrix4{}

	for i := 0; i < len(m); i++ {
		m[i] = Uint2Float64(mm.Read(addr + i))
	}

	return m

}

func memPackMatrix4(mm *memory.MemoryControlBlock, addr int, v *glmath.Matrix4) {

	for i, vv := range v {
		mm.Write(addr+i, Float642uint(vv))
	}

}

type OrbitCameraData struct {
	MCB      *memory.MemoryControlBlock
	baseaddr int
}

func NewOrbitCamera(RAM *memory.MemoryMap, slotid int, baseaddr int, fov, aspect, near, far float64, eye *glmath.Vector3, target *glmath.Vector3, angle *glmath.Vector3) *OrbitCameraData {

	//log2.Printf("Orbit camera init: near = %f, far = %f", near, far)

	mcb := memory.NewMemoryControlBlock(RAM, slotid, false)

	baseglobal := RAM.MEMBASE(slotid) + baseaddr
	baseglobalend := baseglobal + memory.OCTALYZER_MAPPED_CAM_SIZE
	mcb.Add(RAM.Data[slotid][baseglobal:baseglobalend], baseglobal)

	this := &OrbitCameraData{MCB: mcb, baseaddr: baseaddr}

	for i := 0; i < mcb.Size; i++ {
		mcb.Write(i, 0)
	}

	this.setPosition(eye)
	// this.SetViewDir(mgl64.Vec3{0, 0, 1})
	// this.SetUpDir(mgl64.Vec3{0, 1, 0})
	// this.SetRightDir(mgl64.Vec3{1, 0, 0})
	this.setTarget(target)
	d := target.Sub(eye)
	this.setDistance(d.Len())
	this.SetAngle(angle)

	this.SetFOV(fov)
	this.SetAspect(aspect)
	this.SetNear(near)
	this.SetFar(far)
	// this.SetZoom(1)
	this.SetLockView(false)
	this.SetPanX(0)
	this.SetPanY(0)

	this.SetQuaternion(&glmath.Quaternion{1, 0, 0, 0})
	this.LookAtWithPosition(eye, target)

	return this

}

func NewOrbitControllerOSD(RAM *memory.MemoryMap, slotid int) *OrbitCameraData {

	baseaddr := memory.OCTALYZER_OSD_CAM

	baseglobal := RAM.MEMBASE(slotid) + baseaddr
	baseglobalend := baseglobal + memory.OCTALYZER_MAPPED_CAM_SIZE
	mcb := memory.NewMemoryControlBlock(RAM, slotid, false)
	mcb.Add(RAM.Data[slotid][baseglobal:baseglobalend], baseglobal)

	this := &OrbitCameraData{MCB: mcb, baseaddr: baseaddr}

	return this

}

func NewOrbitController(RAM *memory.MemoryMap, slotid int, camid int) *OrbitCameraData {

	baseaddr := memory.OCTALYZER_MAPPED_CAM_BASE + (camid+1)*memory.OCTALYZER_MAPPED_CAM_SIZE

	baseglobal := RAM.MEMBASE(slotid) + baseaddr
	baseglobalend := baseglobal + memory.OCTALYZER_MAPPED_CAM_SIZE
	mcb := memory.NewMemoryControlBlock(RAM, slotid, false)
	mcb.Add(RAM.Data[slotid][baseglobal:baseglobalend], baseglobal)

	this := &OrbitCameraData{MCB: mcb, baseaddr: baseaddr}

	return this

}

func (pcd *OrbitCameraData) GetMatrix() *glmath.Matrix4 {
	return memUnpackMatrix4(pcd.MCB, OCD_MATRIX)
}

func (pcd *OrbitCameraData) SetMatrix(v *glmath.Matrix4) {
	memPackMatrix4(pcd.MCB, OCD_MATRIX, v)
}

func (pcd *OrbitCameraData) GetMatrixRotation() *glmath.Matrix4 {
	return memUnpackMatrix4(pcd.MCB, OCD_MATRIX_ROT)
}

func (pcd *OrbitCameraData) SetMatrixRotation(v *glmath.Matrix4) {
	memPackMatrix4(pcd.MCB, OCD_MATRIX_ROT, v)
}

func (pcd *OrbitCameraData) GetQuaternion() *glmath.Quaternion {
	return memUnpackQuat(pcd.MCB, OCD_QUAT)
}

func (pcd *OrbitCameraData) SetQuaternion(v *glmath.Quaternion) {
	memPackQuat(pcd.MCB, OCD_QUAT, v)
}

func (pcd *OrbitCameraData) LookAtWithPosition(position, target *glmath.Vector3) {

	pcd.setPosition(position)
	pcd.setTarget(target)

	matrix := pcd.GetMatrix()
	matrixRotation := pcd.GetMatrixRotation()
	angle := pcd.GetAngle()
	quaternion := pcd.GetQuaternion()

	// make sure we update changes
	defer func() {
		pcd.SetMatrix(matrix)
		pcd.SetMatrixRotation(matrixRotation)
		pcd.SetAngle(angle)
		pcd.SetQuaternion(quaternion)
	}()

	if position.Equals(*target) {
		matrix.Identity()
		matrix.SetColumnV3(3, glmath.NewVector3(0, 0, 0).Sub(position))
		matrixRotation.Identity()
		angle.Set(0, 0, 0)
		quaternion.Set(1, 0, 0, 0)
		return
	}

	var left, up, forward *glmath.Vector3
	forward = position.Sub(target)
	forward.Normalize()

	if math.Abs(forward.X()) < glmath.EPSILON && math.Abs(forward.Z()) < glmath.EPSILON {
		// forward vector is pointing +Y axis
		if forward.Y() > 0 {
			up = glmath.NewVector3(0, 0, -1)
		} else {
			up = glmath.NewVector3(0, 0, 1)
		}
	} else {
		up = glmath.NewVector3(0, 1, 0)
	}

	left = up.Cross(forward)
	left.Normalize()

	up = forward.Cross(left)

	matrixRotation.Identity()
	matrixRotation.SetRowV3(0, left)
	matrixRotation.SetRowV3(1, up)
	matrixRotation.SetRowV3(2, forward)

	matrix.Identity()
	matrix.SetRowV3(0, left)
	matrix.SetRowV3(1, up)
	matrix.SetRowV3(2, forward)

	trans := &glmath.Vector3{
		matrix[0]*-position[0] + matrix[4]*-position[1] + matrix[8]*-position[2],
		matrix[1]*-position[0] + matrix[5]*-position[1] + matrix[9]*-position[2],
		matrix[2]*-position[0] + matrix[6]*-position[1] + matrix[10]*-position[2],
	}
	matrix.SetColumnV3(3, trans)

	angle = matrixToAngle(matrixRotation)
	reversedAngle := &glmath.Vector3{angle[0], -angle[1], angle[2]}
	quaternion = glmath.GetQuaternion(reversedAngle.MulF(glmath.DEG2RAD * 0.5))

}

func (pcd *OrbitCameraData) LookAtWithPositionUpDir(position, target, upDir *glmath.Vector3) {

	pcd.setPosition(position)
	pcd.setTarget(target)

	matrix := pcd.GetMatrix()
	matrixRotation := pcd.GetMatrixRotation()
	angle := pcd.GetAngle()
	quaternion := pcd.GetQuaternion()
	distance := pcd.GetDistance()

	// make sure we update changes
	defer func() {
		pcd.SetMatrix(matrix)
		pcd.SetMatrixRotation(matrixRotation)
		pcd.SetAngle(angle)
		pcd.SetQuaternion(quaternion)
		pcd.setDistance(distance)
	}()

	if position.Equals(*target) {
		matrix.Identity()
		matrix.Translate(-position.X(), -position.Y(), -position.Z())
		matrixRotation.Identity()
		angle.Set(0, 0, 0)
		quaternion.Set(1, 0, 0, 0)
		return
	}

	var left, up, forward *glmath.Vector3

	forward = position.Sub(target)
	distance = forward.Len()
	forward.Normalize()

	left = upDir.Cross(forward)
	left.Normalize()

	up = forward.Cross(left)

	matrixRotation.Identity()
	matrixRotation.SetRowV3(0, left)
	matrixRotation.SetRowV3(1, up)
	matrixRotation.SetRowV3(2, forward)

	matrix.Identity()
	matrix.SetRowV3(0, left)
	matrix.SetRowV3(1, up)
	matrix.SetRowV3(2, forward)

	trans := &glmath.Vector3{
		matrix[0]*-position[0] + matrix[4]*-position[1] + matrix[8]*-position[2],
		matrix[1]*-position[0] + matrix[5]*-position[1] + matrix[9]*-position[2],
		matrix[2]*-position[0] + matrix[6]*-position[1] + matrix[10]*-position[2],
	}
	matrix.SetColumnV3(3, trans)

	angle = matrixToAngle(matrixRotation)
	reversedAngle := &glmath.Vector3{angle[0], -angle[1], angle[2]}
	quaternion = glmath.GetQuaternion(reversedAngle.MulF(glmath.DEG2RAD * 0.5))

}

func (pcd *OrbitCameraData) LookAtWithPositionF(px, py, pz, tx, ty, tz float64) {
	pcd.LookAtWithPosition(
		glmath.NewVector3(px, py, pz),
		glmath.NewVector3(tx, ty, tz),
	)
}

func (pcd *OrbitCameraData) LookAtWithPositionUpDirF(px, py, pz, tx, ty, tz, ux, uy, uz float64) {
	pcd.LookAtWithPositionUpDir(
		glmath.NewVector3(px, py, pz),
		glmath.NewVector3(tx, ty, tz),
		glmath.NewVector3(ux, uy, uz),
	)
}

func matrixToAngle(matrix *glmath.Matrix4) *glmath.Vector3 {
	angle := matrix.GetAngle()
	angle.SetY(-angle.Y())
	return angle
}

func (pcd *OrbitCameraData) GetPosition() *glmath.Vector3 {

	return memUnpackVector3(pcd.MCB, OCD_POSITION)

}

func (pcd *OrbitCameraData) setPosition(v *glmath.Vector3) {

	memPackVector3(pcd.MCB, OCD_POSITION, v)

}

func (pcd *OrbitCameraData) computeMatrix() {

	matrixRotation := pcd.GetMatrixRotation()
	matrix := pcd.GetMatrix()

	left := glmath.NewVector3(matrixRotation[0], matrixRotation[1], matrixRotation[2])
	up := glmath.NewVector3(matrixRotation[4], matrixRotation[5], matrixRotation[6])
	forward := glmath.NewVector3(matrixRotation[8], matrixRotation[9], matrixRotation[10])

	target := pcd.GetTarget()
	distance := pcd.GetDistance()

	// compute translation vector
	trans := &glmath.Vector3{
		left.X()*-target.X() + up.X()*-target.Y() + forward.X()*-target.Z(),
		left.Y()*-target.X() + up.Y()*-target.Y() + forward.Y()*-target.Z(),
		left.Z()*-target.X() + up.Z()*-target.Y() + forward.Z()*-target.Z() - distance,
	}

	// construct matrix
	matrix.Identity()
	matrix.SetColumnV3(0, left)
	matrix.SetColumnV3(1, up)
	matrix.SetColumnV3(2, forward)
	matrix.SetColumnV3(3, trans)

	// re-compute camera position
	forward.Set(-matrix[2], -matrix[6], -matrix[10])
	pcd.setPosition(target.Sub(forward.MulF(distance)))
	pcd.SetMatrix(matrix)
}

func (pcd *OrbitCameraData) SetTarget(v *glmath.Vector3) {
	pcd.setTarget(v)
	matrix := pcd.GetMatrix()
	forward := &glmath.Vector3{-matrix[2], -matrix[6], -matrix[10]}
	position := v.Sub(forward.MulF(pcd.GetDistance()))
	pcd.setPosition(position)
	pcd.computeMatrix()
}

func (pcd *OrbitCameraData) SetPosition(v *glmath.Vector3) {

	pcd.LookAtWithPosition(v, pcd.GetTarget())

}

func (pcd *OrbitCameraData) GetAngle() *glmath.Vector3 {

	return memUnpackVector3(pcd.MCB, OCD_ANGLE)

}

func (pcd *OrbitCameraData) SetAngle(v *glmath.Vector3) {

	memPackVector3(pcd.MCB, OCD_ANGLE, v)

}

func (pcd *OrbitCameraData) GetTarget() *glmath.Vector3 {

	return memUnpackVector3(pcd.MCB, OCD_TARGET)

}

func (pcd *OrbitCameraData) setTarget(v *glmath.Vector3) {

	memPackVector3(pcd.MCB, OCD_TARGET, v)

}

func (pcd *OrbitCameraData) SetDefaultPos(v *glmath.Vector3) {

	memPackVector3(pcd.MCB, OCD_DEFAULT_POS, v)

}

func (pcd *OrbitCameraData) GetDefaultPos() *glmath.Vector3 {

	return memUnpackVector3(pcd.MCB, OCD_DEFAULT_POS)

}

func (pcd *OrbitCameraData) SetDefaultTarget(v *glmath.Vector3) {

	memPackVector3(pcd.MCB, OCD_DEFAULT_TARGET, v)

}

func (pcd *OrbitCameraData) GetDefaultTarget() *glmath.Vector3 {

	return memUnpackVector3(pcd.MCB, OCD_DEFAULT_TARGET)

}

func (pcd *OrbitCameraData) SetDefaultAngle(v *glmath.Vector3) {

	memPackVector3(pcd.MCB, OCD_DEFAULT_TARGET, v)

}

func (pcd *OrbitCameraData) GetDefaultAngle() *glmath.Vector3 {

	return memUnpackVector3(pcd.MCB, OCD_DEFAULT_TARGET)

}

func (pcd *OrbitCameraData) SetFOV(v float64) {

	memPackFloat32(pcd.MCB, OCD_FOV, v)

}

func (pcd *OrbitCameraData) GetFOV() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_FOV)

}

func (pcd *OrbitCameraData) SetPanX(v float64) {

	memPackFloat32(pcd.MCB, OCD_PAN_X, v)

}

func (pcd *OrbitCameraData) GetPanX() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_PAN_X)

}

func (pcd *OrbitCameraData) SetPanY(v float64) {

	memPackFloat32(pcd.MCB, OCD_PAN_Y, v)

}

func (pcd *OrbitCameraData) GetPanY() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_PAN_Y)

}

func (pcd *OrbitCameraData) SetAspect(v float64) {

	memPackFloat32(pcd.MCB, OCD_ASPECT, v)

}

func (pcd *OrbitCameraData) GetAspect() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_ASPECT)

}

func (pcd *OrbitCameraData) SetNear(v float64) {

	memPackFloat32(pcd.MCB, OCD_NEAR, v)

}

func (pcd *OrbitCameraData) GetNear() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_NEAR)

}

func (pcd *OrbitCameraData) SetFar(v float64) {

	memPackFloat32(pcd.MCB, OCD_FAR, v)

}

func (pcd *OrbitCameraData) GetFar() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_FAR)

}

func (pcd *OrbitCameraData) SetDollyRate(v float64) {

	memPackFloat32(pcd.MCB, OCD_DOLLYRATE, v)

}

func (pcd *OrbitCameraData) GetDollyRate() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_DOLLYRATE)

}

func (pcd *OrbitCameraData) SetZoom(v float64) {

	memPackFloat32(pcd.MCB, OCD_ZOOM, v*GFXMULT)

}

func (pcd *OrbitCameraData) GetZoom() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_ZOOM) / GFXMULT

}

func (pcd *OrbitCameraData) setDistance(v float64) {

	memPackFloat32(pcd.MCB, OCD_DISTANCE, v)

}

func (pcd *OrbitCameraData) SetDistance(v float64) {
	pcd.setDistance(v)
	pcd.computeMatrix()
}

func (pcd *OrbitCameraData) GetDistance() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_DISTANCE)

}

func (pcd *OrbitCameraData) SetShakeFrames(v int) {

	memPackInt(pcd.MCB, OCD_SHAKEFRAMES, v)

}

func (pcd *OrbitCameraData) GetShakeFrames() int {

	return memUnpackInt(pcd.MCB, OCD_SHAKEFRAMES)

}

func (pcd *OrbitCameraData) SetShakeMax(v float64) {

	memPackFloat32(pcd.MCB, OCD_SHAKEMAX, v)

}

func (pcd *OrbitCameraData) GetShakeMax() float64 {

	return memUnpackFloat32(pcd.MCB, OCD_SHAKEMAX)

}

func (pcd *OrbitCameraData) SetDPos(v bool) {

	memPackBool(pcd.MCB, OCD_DPOS, v)

}

func (pcd *OrbitCameraData) GetDPos() bool {

	return memUnpackBool(pcd.MCB, OCD_DPOS)

}

func (pcd *OrbitCameraData) SetDView(v bool) {

	memPackBool(pcd.MCB, OCD_DVIEW, v)

}

func (pcd *OrbitCameraData) GetDView() bool {

	return memUnpackBool(pcd.MCB, OCD_DVIEW)

}

func (pcd *OrbitCameraData) SetLockView(v bool) {

	memPackBool(pcd.MCB, OCD_LOCKVIEW, v)

}

func (pcd *OrbitCameraData) GetLockView() bool {

	return memUnpackBool(pcd.MCB, OCD_LOCKVIEW)

}

func (this *OrbitCameraData) Move(dx, dy, dz float64) {
	this.Advance(dz)
	this.Strafe(dx)
	this.Ascend(dy)
}

func (this *OrbitCameraData) Pan(dx, dy, dz float64) {
	// this.StrafeLock(dx)
	// this.AscendLock(dy)
	this.SetPanX(this.GetPanX() + dx)
	this.SetPanY(this.GetPanY() + dy)
}

// Manipulation functions
func (this *OrbitCameraData) Advance(distance float64) {
	this.setPosition(this.GetPosition().Add(this.GetForwardAxis().MulF(distance)))
	this.SetTarget(this.GetTarget().Add(this.GetForwardAxis().MulF(distance)))
}

func (this *OrbitCameraData) Ascend(distance float64) {
	this.setPosition(this.GetPosition().Add(this.GetUpAxis().MulF(distance)))
	this.SetTarget(this.GetTarget().Add(this.GetUpAxis().MulF(distance)))
}

func (this *OrbitCameraData) Strafe(distance float64) {
	this.setPosition(this.GetPosition().Add(this.GetLeftAxis().MulF(-distance)))
	this.SetTarget(this.GetTarget().Add(this.GetLeftAxis().MulF(-distance)))
}

func (this *OrbitCameraData) GetLeftAxis() *glmath.Vector3 {
	matrix := this.GetMatrix()
	return glmath.NewVector3(-matrix[0], -matrix[4], -matrix[8])
}

func (this *OrbitCameraData) GetUpAxis() *glmath.Vector3 {
	matrix := this.GetMatrix()
	return glmath.NewVector3(matrix[1], matrix[5], matrix[9])
}

func (this *OrbitCameraData) GetForwardAxis() *glmath.Vector3 {
	matrix := this.GetMatrix()
	return glmath.NewVector3(-matrix[2], -matrix[6], -matrix[10])
}

func (this *OrbitCameraData) AdvanceLock(distance float64) {
	this.setPosition(this.GetPosition().Add(this.GetForwardAxis().MulF(-distance)))
	this.setTarget(this.GetTarget())
	//this.LookAt.Add( this.ViewDir * -distance )
}

func (this *OrbitCameraData) AscendLock(distance float64) {
	this.setPosition(this.GetPosition().Add(this.GetUpAxis().MulF(distance)))
	this.setTarget(this.GetTarget())
	//this.LookAt.Add( this.UpDir * distance )
}

func (this *OrbitCameraData) StrafeLock(distance float64) {
	t := this.GetTarget()
	this.setPosition(this.GetPosition().Add(this.GetLeftAxis().MulF(distance)))
	this.setTarget(t)
	//this.LookAt.Add( this.RightDir * distance )
}

func (this *OrbitCameraData) angleToMatrix(angle *glmath.Vector3) *glmath.Matrix4 {
	var sx, sy, sz, cx, cy, cz, theta float64
	var left, up, forward glmath.Vector3

	// rotation angle about X-axis (pitch)
	theta = angle.X() * glmath.DEG2RAD
	sx = math.Sin(theta)
	cx = math.Cos(theta)

	// rotation angle about Y-axis (yaw)
	theta = -angle.Y() * glmath.DEG2RAD
	sy = math.Sin(theta)
	cy = math.Cos(theta)

	// rotation angle about Z-axis (roll)
	theta = angle.Z() * glmath.DEG2RAD
	sz = math.Sin(theta)
	cz = math.Cos(theta)

	// determine left axis
	left.SetX(cy * cz)
	left.SetY(sx*sy*cz + cx*sz)
	left.SetZ(-cx*sy*cz + sx*sz)

	// determine up axis
	up.SetX(-cy * sz)
	up.SetY(-sx*sy*sz + cx*cz)
	up.SetZ(cx*sy*sz + sx*cz)

	// determine forward axis
	forward.SetX(sy)
	forward.SetY(-sx * cy)
	forward.SetZ(cx * cy)

	// construct rotation matrix
	matrix := &glmath.Matrix4{}
	matrix.SetColumnV3(0, &left)
	matrix.SetColumnV3(1, &up)
	matrix.SetColumnV3(2, &forward)

	return matrix
}

func (this *OrbitCameraData) SetRotation(angle *glmath.Vector3) {
	this.SetAngle(angle)

	reversedAngle := glmath.NewVector3(angle.X(), -angle.Y(), angle.Z())
	quaternion := glmath.GetQuaternion(reversedAngle)

	matrixRotation := this.angleToMatrix(angle)

	this.SetQuaternion(quaternion)
	this.SetMatrixRotation(matrixRotation)

	this.computeMatrix()
}

func (this *OrbitCameraData) Pitch(angle float64) {
	this.pitch(angle)
}

func (this *OrbitCameraData) pitch(angle float64) {

	a := this.GetAngle()
	a.SetX(a.X() + angle)
	this.SetRotation(a)

}

func (this *OrbitCameraData) Yaw(angle float64) {
	this.yaw(angle)
}

func (this *OrbitCameraData) yaw(angle float64) {
	a := this.GetAngle()
	a.SetY(a.Y() + angle)
	this.SetRotation(a)
}

func (this *OrbitCameraData) Roll(angle float64) {
	this.roll(angle)
}

func (this *OrbitCameraData) roll(angle float64) {
	a := this.GetAngle()
	a.SetZ(a.Z() + angle)
	this.SetRotation(a)
}

func (pc *OrbitCameraData) Rotate3DX(a float64) {

	pc.pitch(a)

}

func (pc *OrbitCameraData) Rotate3DY(a float64) {

	pc.yaw(a)

}

func (pc *OrbitCameraData) Rotate3DZ(a float64) {

	pc.roll(a)

}

func (pc *OrbitCameraData) ResetPosition() {
	if def, ok := settings.CameraInitDefaults[pc.baseaddr]; ok {
		pc.setTarget(&glmath.Vector3{def.Eye[0], def.Eye[1], def.Eye[2]})
	} else {
		pc.setPosition(&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, CDIST * GFXMULT})
	}
}

func (pc *OrbitCameraData) ResetLookAt() {
	if def, ok := settings.CameraInitDefaults[pc.baseaddr]; ok {
		pc.setTarget(&glmath.Vector3{def.Target[0], def.Target[1], def.Target[2]})
	} else {
		pc.setTarget(&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, 0})
	}
}

func (this *OrbitCameraData) ResetOrientation() {
	if def, ok := settings.CameraInitDefaults[this.baseaddr]; ok {
		this.SetAngle(&glmath.Vector3{def.Angle[0], def.Angle[1], def.Angle[2]})
	} else {
		this.SetAngle(&glmath.Vector3{0, 0, 0})
	}
}

func (this *OrbitCameraData) ResetZoom() {
	if def, ok := settings.CameraInitDefaults[this.baseaddr]; ok {
		this.SetZoom(def.Zoom)
	} else {
		this.SetZoom(GFXMULT)
	}
	fmt.Println("ZOOM HAS BEEN RESET!!!!")
}

func (pc *OrbitCameraData) SetPos(x, y, z float64) {
	pc.SetPosV(&glmath.Vector3{x, y, z})
}

func (pc *OrbitCameraData) SetPosV(v *glmath.Vector3) {
	pc.SetPosition(v)
	if !pc.GetDPos() {
		pc.SetDPos(true)
		pc.SetDefaultPos(v)
	}
}

func (this *OrbitCameraData) ResetALL() {

	var eye, target, angle *glmath.Vector3
	var aspect, near, far, fov float64
	var lockview bool
	var zoom float64

	if def, ok := settings.CameraInitDefaults[this.baseaddr]; ok {
		//log2.Printf("Resetting camera with init defaults: %+v", def)
		eye = &glmath.Vector3{def.Eye[0], def.Eye[1], def.Eye[2]}
		target = &glmath.Vector3{def.Target[0], def.Target[1], def.Target[2]}
		angle = &glmath.Vector3{def.Angle[0], def.Angle[1], def.Angle[2]}
		aspect = float64(this.GetAspect())
		near = float64(def.Near)
		far = float64(def.Far)
		fov = float64(def.FOV)
		zoom = def.Zoom
		lockview = def.PivotLock
	} else {
		eye = &glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, CDIST * GFXMULT}
		target = &glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, 0}
		angle = &glmath.Vector3{0, 0, 0}
		aspect = float64(this.GetAspect())
		near = float64(12750)
		far = float64(15000)
		fov = float64(60)
		zoom = GFXMULT
		lockview = true
	}
	this.setPosition(eye)
	this.setTarget(target)
	d := target.Sub(eye)
	this.setDistance(d.Len())
	this.SetAngle(angle)

	this.SetFOV(fov)
	this.SetAspect(aspect)
	this.SetNear(near)
	this.SetFar(far)
	this.SetZoom(zoom)
	this.SetLockView(lockview)
	this.SetPanX(0)
	this.SetPanY(0)

	this.SetQuaternion(&glmath.Quaternion{1, 0, 0, 0})
	this.LookAtWithPosition(eye, target)
}

func (pc *OrbitCameraData) SetPivotLock(b bool) {
	pc.SetLockView(true)
}

// Fixed orbit function
func (pc *OrbitCameraData) CalculateOrbit(yaw, pitch, roll float64) *glmath.Vector3 {

	pc.roll(roll)
	pc.pitch(pitch)
	pc.yaw(yaw)

	return pc.GetPosition()
}

func (pc *OrbitCameraData) Orbit(yaw, pitch float64) {
	_ = pc.CalculateOrbit(yaw, pitch, 0)
	//pc.setPosition( p[0], p[1], p[2] )
	//pc.SetPosV(p)
}

func (pc *OrbitCameraData) Rotate(x, y, z float64) {
	pc.pitch(x)
	pc.yaw(y)
	pc.roll(z)
	//pc.setPosition( p[0], p[1], p[2] )
	//pc.SetPosV(p)
}

func (pc *OrbitCameraData) SetLookAtV(v *glmath.Vector3) {

	pc.SetTarget(v)

	if !pc.GetDView() {
		pc.SetDView(true)
		pc.SetDefaultTarget(pc.GetTarget())
	}
}

func (pc *OrbitCameraData) RotateAxis(a float64, axis mgl64.Vec3) {
	// r := mgl64.DegToRad(a)

	// m := mgl64.HomogRotate3D(r, axis)
	// //pc.Position = mgl64.TransformCoordinate(pc.Position, m)
	// pc.SetUpDir(mgl64.TransformNormal(pc.GetUpDir(), m).Normalize())
	// pc.SetRightDir(mgl64.TransformNormal(pc.GetRightDir(), m).Normalize())
	// pc.SetViewDir(mgl64.TransformNormal(pc.GetViewDir(), m).Normalize())

	//      //fmt.Printf("After rotate of %f, viewdir is %f, %f, %f (total %f)\n", a, pc.ViewDir[0], pc.ViewDir[1], pc.ViewDir[2], pc.ViewDir[0]+pc.ViewDir[1]+pc.ViewDir[2])
}

func (pc *OrbitCameraData) String() string {

	// var c CamData

	// c.P = pc.GetPosition()
	// c.U = pc.GetUpDir()
	// c.R = pc.GetRightDir()
	// c.V = pc.GetViewDir()
	// c.Z = pc.GetZoom()
	// c.RX, c.RY, c.RZ = pc.GetRotatedX(), pc.GetRotatedY(), pc.GetRotatedZ()
	// c.L = pc.GetLockView()

	// b, _ := json.Marshal(&c)

	b := []byte(nil)

	return string(b)
}

func (pc *OrbitCameraData) FromString(s string) {

	// var c CamData

	// _ = json.Unmarshal([]byte(s), &c)

	// pc.setPosition(c.P)
	// pc.SetUpDir(c.U)
	// pc.SetRightDir(c.R)
	// pc.SetViewDir(c.V)
	// pc.SetLockView(c.L)
	// pc.SetZoom(c.Z)
	// pc.SetRotatedX(c.RX)
	// pc.SetRotatedY(c.RY)
	// pc.SetRotatedZ(c.RZ)

}

func (pc *OrbitCameraData) GetState() ([]float64, []float64, []float64) {

	scale := glmath.Scale(pc.GetZoom(), pc.GetZoom(), 1)
	modelview := pc.GetMatrix()
	persp := glmath.Perspective(pc.GetFOV(), pc.GetAspect(), pc.GetNear(), pc.GetFar())
	//persp := mgl64.Perspective(pc.GetFOV(), pc.GetAspect(), pc.GetNear(), pc.GetFar())

	return scale[:], modelview[:], persp[:]
}

func (pc *OrbitCameraData) Update() {
	pc.computeMatrix()
	pc.pitch(0)
}
