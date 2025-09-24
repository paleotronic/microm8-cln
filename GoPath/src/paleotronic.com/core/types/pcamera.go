package types

import (
	"encoding/json"
	"math"

	"paleotronic.com/fmt"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/memory"
)

const (
	CHEIGHT   = 960
	CDIST     = 854
	GFXMULT   = 16
	PASPECT   = 1.777777778
	CRTASPECT = 1.458
	CWIDTH    = CRTASPECT * CHEIGHT
)

const (
	PCD_POSITION = 0
	PCD_VIEWDIR  = 3
	PCD_UPDIR    = 6
	PCD_RIGHTDIR = 9
	PCD_LOOKAT   = 12

	PCD_FOV    = 15
	PCD_ASPECT = 16
	PCD_NEAR   = 17
	PCD_FAR    = 18

	PCD_DEFAULTPOS  = 19
	PCD_DEFAULTVIEW = 22

	PCD_DPOS     = 25
	PCD_DVIEW    = 26
	PCD_LOCKVIEW = 27
	PCD_DIRTY    = 28

	PCD_DOLLYRATE   = 29
	PCD_SHAKEFRAMES = 30
	PCD_SHAKEMAX    = 31

	PCD_ROTATEDX = 32
	PCD_ROTATEDY = 33
	PCD_ROTATEDZ = 34

	PCD_ZOOM = 35

	PCD_SIZE = 36
)

type PerspCameraData struct {
	// Needed for memory mapping
	// Index int
	// Base  int // Base Address
	// Mm    *memory.MemoryMap

	MCB *memory.MemoryControlBlock

	//~ Position mgl64.Vec3
	//~ ViewDir  mgl64.Vec3
	//~ UpDir    mgl64.Vec3
	//~ RightDir mgl64.Vec3
	//~ LookAt   mgl64.Vec3
	//~ //
	//~ FOV    float64
	//~ Aspect float64
	//~ Near   float64
	//~ Far    float64
	//~ // Defaults
	//~ defaultPos  mgl64.Vec3
	//~ defaultView mgl64.Vec3
	//~ dPos  bool
	//~ dView bool
	//~ LockView    bool
	//~ Zoom        float64
	//~ DollyRate   float64
	//~ DollyPos    []mgl64.Vec3
	//~ DollyTarget []mgl64.Vec3
	//~ ViewPort    PRect
	//~ PixelViewPort PRect
	//~ ShakeFrames int
	//~ ShakeMax    float64
	//~ RotatedX float64
	//~ RotatedY float64
	//~ RotatedZ float64
}

func memUnpackFloat32(mm *memory.MemoryControlBlock, addr int) float64 {
	return Uint2Float64(mm.Read(addr))
}

func memPackFloat32(mm *memory.MemoryControlBlock, addr int, f float64) {
	mm.Write(addr, Float642uint(f))
}

func memUnpackInt(mm *memory.MemoryControlBlock, addr int) int {
	return int(mm.Read(addr))
}

func memPackInt(mm *memory.MemoryControlBlock, addr int, f int) {
	mm.Write(addr, uint64(f))
}

func memUnpackBool(mm *memory.MemoryControlBlock, addr int) bool {
	return (mm.Read(addr) != 0)
}

func memPackBool(mm *memory.MemoryControlBlock, addr int, b bool) {
	var v uint64
	if b {
		v = 1
	}
	mm.Write(addr, v)
}

func memUnpackVec3(mm *memory.MemoryControlBlock, addr int) mgl64.Vec3 {

	return mgl64.Vec3{
		Uint2Float64(mm.Read(addr + 0)),
		Uint2Float64(mm.Read(addr + 1)),
		Uint2Float64(mm.Read(addr + 2)),
	}

}

func memPackVec3(mm *memory.MemoryControlBlock, addr int, v mgl64.Vec3) {
	mm.Write(addr+0, Float642uint(v[0]))
	mm.Write(addr+1, Float642uint(v[1]))
	mm.Write(addr+2, Float642uint(v[2]))
}

func NewPerspCamera(RAM *memory.MemoryMap, slotid int, baseaddr int, fov, aspect, near, far float64, eye mgl64.Vec3) *PerspCameraData {

	mcb := memory.NewMemoryControlBlock(RAM, slotid, false)

	baseglobal := RAM.MEMBASE(slotid) + baseaddr
	baseglobalend := baseglobal + memory.OCTALYZER_MAPPED_CAM_SIZE
	mcb.Add(RAM.Data[slotid][baseglobal:baseglobalend], baseglobal)

	this := &PerspCameraData{MCB: mcb}

	this.SetPosition(mgl64.Vec3{0, 0, 0})
	this.SetViewDir(mgl64.Vec3{0, 0, 1})
	this.SetUpDir(mgl64.Vec3{0, 1, 0})
	this.SetRightDir(mgl64.Vec3{1, 0, 0})
	this.SetLookAt(eye)

	this.SetFOV(fov)
	this.SetAspect(aspect)
	this.SetNear(near)
	this.SetFar(far)
	this.SetZoom(1)
	this.SetLockView(false)

	return this

}

func NewPerspController(RAM *memory.MemoryMap, slotid int, camid int) *PerspCameraData {

	baseaddr := memory.OCTALYZER_MAPPED_CAM_BASE + (camid+1)*memory.OCTALYZER_MAPPED_CAM_SIZE

	baseglobal := RAM.MEMBASE(slotid) + baseaddr
	baseglobalend := baseglobal + memory.OCTALYZER_MAPPED_CAM_SIZE
	mcb := memory.NewMemoryControlBlock(RAM, slotid, false)
	mcb.Add(RAM.Data[slotid][baseglobal:baseglobalend], baseglobal)

	this := &PerspCameraData{MCB: mcb}

	//~ this.SetPosition( mgl64.Vec3{0, 0, 0} )
	//~ this.SetViewDir( mgl64.Vec3{0, 0, 1} )
	//~ this.SetUpDir( mgl64.Vec3{0, 1, 0} )
	//~ this.SetRightDir( mgl64.Vec3{1, 0, 0} )
	//~ this.SetLookAt(eye)

	//~ this.SetFOV(fov)
	//~ this.SetAspect(aspect)
	//~ this.SetNear(near)
	//~ this.SetFar(far)
	//~ this.SetZoom(1)
	//~ this.SetLockView(false)

	return this

}

func (pcd *PerspCameraData) GetPosition() mgl64.Vec3 {

	return memUnpackVec3(pcd.MCB, PCD_POSITION)

}

func (pcd *PerspCameraData) SetPosition(v mgl64.Vec3) {

	memPackVec3(pcd.MCB, PCD_POSITION, v)

}

func (pcd *PerspCameraData) GetViewDir() mgl64.Vec3 {

	return memUnpackVec3(pcd.MCB, PCD_VIEWDIR)

}

func (pcd *PerspCameraData) SetViewDir(v mgl64.Vec3) {

	memPackVec3(pcd.MCB, PCD_VIEWDIR, v)

}

func (pcd *PerspCameraData) GetUpDir() mgl64.Vec3 {

	return memUnpackVec3(pcd.MCB, PCD_UPDIR)

}

func (pcd *PerspCameraData) SetUpDir(v mgl64.Vec3) {

	memPackVec3(pcd.MCB, PCD_UPDIR, v)

}

func (pcd *PerspCameraData) GetRightDir() mgl64.Vec3 {

	return memUnpackVec3(pcd.MCB, PCD_RIGHTDIR)

}

func (pcd *PerspCameraData) SetRightDir(v mgl64.Vec3) {

	memPackVec3(pcd.MCB, PCD_RIGHTDIR, v)

}

func (pcd *PerspCameraData) GetLookAt() mgl64.Vec3 {

	return memUnpackVec3(pcd.MCB, PCD_LOOKAT)

}

func (pcd *PerspCameraData) SetLookAt(v mgl64.Vec3) {

	memPackVec3(pcd.MCB, PCD_LOOKAT, v)

}

func (pcd *PerspCameraData) SetDefaultPos(v mgl64.Vec3) {

	memPackVec3(pcd.MCB, PCD_DEFAULTPOS, v)

}

func (pcd *PerspCameraData) GetDefaultPos() mgl64.Vec3 {

	return memUnpackVec3(pcd.MCB, PCD_DEFAULTPOS)

}

func (pcd *PerspCameraData) SetDefaultView(v mgl64.Vec3) {

	memPackVec3(pcd.MCB, PCD_DEFAULTVIEW, v)

}

func (pcd *PerspCameraData) GetDefaultView() mgl64.Vec3 {

	return memUnpackVec3(pcd.MCB, PCD_DEFAULTVIEW)

}

func (pcd *PerspCameraData) SetFOV(v float64) {

	memPackFloat32(pcd.MCB, PCD_FOV, v)

}

func (pcd *PerspCameraData) GetFOV() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_FOV)

}

func (pcd *PerspCameraData) SetAspect(v float64) {

	memPackFloat32(pcd.MCB, PCD_ASPECT, v)

}

func (pcd *PerspCameraData) GetAspect() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_ASPECT)

}

func (pcd *PerspCameraData) SetNear(v float64) {

	memPackFloat32(pcd.MCB, PCD_NEAR, v)

}

func (pcd *PerspCameraData) GetNear() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_NEAR)

}

func (pcd *PerspCameraData) SetFar(v float64) {

	memPackFloat32(pcd.MCB, PCD_FAR, v)

}

func (pcd *PerspCameraData) GetFar() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_FAR)

}

func (pcd *PerspCameraData) SetRotatedX(v float64) {

	memPackFloat32(pcd.MCB, PCD_ROTATEDX, v)

}

func (pcd *PerspCameraData) GetRotatedX() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_ROTATEDX)

}

func (pcd *PerspCameraData) SetRotatedY(v float64) {

	memPackFloat32(pcd.MCB, PCD_ROTATEDY, v)

}

func (pcd *PerspCameraData) GetRotatedY() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_ROTATEDY)

}

func (pcd *PerspCameraData) SetRotatedZ(v float64) {

	memPackFloat32(pcd.MCB, PCD_ROTATEDZ, v)

}

func (pcd *PerspCameraData) GetRotatedZ() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_ROTATEDZ)

}

func (pcd *PerspCameraData) SetDollyRate(v float64) {

	memPackFloat32(pcd.MCB, PCD_DOLLYRATE, v)

}

func (pcd *PerspCameraData) GetDollyRate() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_DOLLYRATE)

}

func (pcd *PerspCameraData) SetZoom(v float64) {

	memPackFloat32(pcd.MCB, PCD_ZOOM, v*GFXMULT)

}

func (pcd *PerspCameraData) GetZoom() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_ZOOM) / GFXMULT

}

func (pcd *PerspCameraData) SetShakeFrames(v int) {

	memPackInt(pcd.MCB, PCD_SHAKEFRAMES, v)

}

func (pcd *PerspCameraData) GetShakeFrames() int {

	return memUnpackInt(pcd.MCB, PCD_SHAKEFRAMES)

}

func (pcd *PerspCameraData) SetShakeMax(v float64) {

	memPackFloat32(pcd.MCB, PCD_SHAKEMAX, v)

}

func (pcd *PerspCameraData) GetShakeMax() float64 {

	return memUnpackFloat32(pcd.MCB, PCD_SHAKEMAX)

}

func (pcd *PerspCameraData) SetDPos(v bool) {

	memPackBool(pcd.MCB, PCD_DPOS, v)

}

func (pcd *PerspCameraData) GetDPos() bool {

	return memUnpackBool(pcd.MCB, PCD_DPOS)

}

func (pcd *PerspCameraData) SetDView(v bool) {

	memPackBool(pcd.MCB, PCD_DVIEW, v)

}

func (pcd *PerspCameraData) GetDView() bool {

	return memUnpackBool(pcd.MCB, PCD_DVIEW)

}

func (pcd *PerspCameraData) SetLockView(v bool) {

	memPackBool(pcd.MCB, PCD_LOCKVIEW, v)

}

func (pcd *PerspCameraData) GetLockView() bool {

	return memUnpackBool(pcd.MCB, PCD_LOCKVIEW)

}

func (this *PerspCameraData) Move(dx, dy, dz float64) {
	this.Advance(dz)
	this.Strafe(dx)
	this.Ascend(dy)
}

// Manipulation functions
func (this *PerspCameraData) Advance(distance float64) {
	this.SetPosition(this.GetPosition().Add(this.GetViewDir().Mul(-distance)))
	this.SetLookAt(this.GetLookAt().Add(this.GetViewDir().Mul(-distance)))
}

func (this *PerspCameraData) Ascend(distance float64) {
	this.SetPosition(this.GetPosition().Add(this.GetUpDir().Mul(distance)))
	this.SetLookAt(this.GetLookAt().Add(this.GetUpDir().Mul(distance)))
}

func (this *PerspCameraData) Strafe(distance float64) {
	this.SetPosition(this.GetPosition().Add(this.GetRightDir().Mul(distance)))
	this.SetLookAt(this.GetLookAt().Add(this.GetRightDir().Mul(distance)))
}

func (this *PerspCameraData) AdvanceLock(distance float64) {
	this.SetPosition(this.GetPosition().Add(this.GetViewDir().Mul(-distance)))
	//this.LookAt.Add( this.ViewDir * -distance )
}

func (this *PerspCameraData) AscendLock(distance float64) {
	this.SetPosition(this.GetPosition().Add(this.GetUpDir().Mul(distance)))
	//this.LookAt.Add( this.UpDir * distance )
}

func (this *PerspCameraData) StrafeLock(distance float64) {
	this.SetPosition(this.GetPosition().Add(this.GetRightDir().Mul(distance)))
	//this.LookAt.Add( this.RightDir * distance )
}

func cosf(r float64) float64 {
	return float64(math.Cos(float64(r)))
}

func sinf(r float64) float64 {
	return float64(math.Sin(float64(r)))
}

func (this *PerspCameraData) pitch(angle float64) {
	// keep track of how far we've gone around the axis
	this.SetRotatedX(this.GetRotatedX() + angle)

	// calculate the new forward vector
	this.SetViewDir(this.GetViewDir().Mul(cosf(mgl64.DegToRad(angle))).Add(
		this.GetUpDir().Mul(sinf(mgl64.DegToRad(angle))),
	))

	// calculate the new up vector
	this.SetUpDir(this.GetViewDir().Cross(this.GetRightDir()))

	// invert so that positive goes down
	this.SetUpDir(this.GetUpDir().Mul(-1))
}

func (this *PerspCameraData) yaw(angle float64) {
	// keep track of how far we've gone around the axis
	this.SetRotatedY(this.GetRotatedY() + angle)

	// calculate the new forward vector
	this.SetViewDir(this.GetViewDir().Mul(cosf(mgl64.DegToRad(angle))).Add(
		this.GetRightDir().Mul(-1 * sinf(mgl64.DegToRad(angle))),
	))

	// calculate the new up vector
	this.SetRightDir(this.GetViewDir().Cross(this.GetUpDir()))
}

func (this *PerspCameraData) roll(angle float64) {
	// keep track of how far we've gone around the axis
	this.SetRotatedZ(this.GetRotatedZ() + angle)

	// calculate the new forward vector
	this.SetRightDir(this.GetRightDir().Mul(cosf(mgl64.DegToRad(angle))).Add(
		this.GetUpDir().Mul(sinf(mgl64.DegToRad(angle))),
	))

	// calculate the new up vector
	this.SetUpDir(this.GetViewDir().Cross(this.GetRightDir()))

	// invert so that positive goes down
	//this.SetUpDir( this.GetUpDir().Mul( -1 ) )
}

func (pc *PerspCameraData) Rotate3DX(a float64) {

	pc.pitch(a)

}

func (pc *PerspCameraData) Rotate3DY(a float64) {

	pc.yaw(a)

}

func (pc *PerspCameraData) Rotate3DZ(a float64) {

	pc.roll(a)

}

func (pc *PerspCameraData) ResetPosition() {
	pc.SetPos(CWIDTH/2, CHEIGHT/2, CDIST*GFXMULT)
}

func (pc *PerspCameraData) ResetLookAt() {
	pc.SetLookAt(mgl64.Vec3{CWIDTH / 2, CHEIGHT / 2, 0})
}

func (this *PerspCameraData) ResetOrientation() {
	this.SetViewDir(mgl64.Vec3{0, 0, 1})
	this.SetUpDir(mgl64.Vec3{0, 1, 0})
	this.SetRightDir(mgl64.Vec3{1, 0, 0})
}

func (this *PerspCameraData) ResetZoom() {
	this.SetZoom(GFXMULT)
	fmt.Println("ZOOM HAS BEEN RESET!!!!")
}

func (pc *PerspCameraData) SetPos(x, y, z float64) {
	pc.SetPosV(mgl64.Vec3{x, y, z})
}

func (pc *PerspCameraData) SetPosV(v mgl64.Vec3) {
	pc.SetPosition(v)
	if !pc.GetDPos() {
		pc.SetDPos(true)
		pc.SetDefaultPos(v)
	}
}

func (this *PerspCameraData) ResetALL() {
	this.SetFOV(60)
	this.SetNear(12750)
	this.SetFar(15000)
	this.SetPos(CWIDTH/2, CHEIGHT/2, CDIST*GFXMULT)
	this.SetLookAt(mgl64.Vec3{CWIDTH / 2, CHEIGHT / 2, 0})
	this.SetPivotLock(true)
	this.SetZoom(GFXMULT)
	this.ResetLookAt()
	this.ResetOrientation()
}

func (pc *PerspCameraData) SetPivotLock(b bool) {
	pc.SetLockView(true)
}

// Fixed orbit function
func (pc *PerspCameraData) CalculateOrbit(yaw, pitch float64) mgl64.Vec3 {

	src := pc.GetPosition()
	target := pc.GetLookAt()

	camFocusVector := src.Sub(target)

	r1 := mgl64.HomogRotate3D(mgl64.DegToRad(yaw), pc.GetUpDir())
	r2 := mgl64.HomogRotate3D(mgl64.DegToRad(pitch), pc.GetRightDir())

	camFocusVector = mgl64.TransformCoordinate(camFocusVector, r1)
	camFocusVector = mgl64.TransformCoordinate(camFocusVector, r2)

	p := pc.GetLookAt().Add(camFocusVector)

	pc.SetUpDir(mgl64.TransformNormal(pc.GetUpDir(), r1))
	pc.SetUpDir(mgl64.TransformNormal(pc.GetUpDir(), r2))

	pc.SetRightDir(mgl64.TransformNormal(pc.GetRightDir(), r1))
	pc.SetRightDir(mgl64.TransformNormal(pc.GetRightDir(), r2))

	pc.SetViewDir(mgl64.TransformNormal(pc.GetViewDir(), r1))
	pc.SetViewDir(mgl64.TransformNormal(pc.GetViewDir(), r2))

	return p
}

func (pc *PerspCameraData) Orbit(yaw, pitch float64) {
	p := pc.CalculateOrbit(yaw, pitch)
	//pc.SetPosition( p[0], p[1], p[2] )
	pc.SetPosV(p)
}

func (pc *PerspCameraData) SetLookAtV(v mgl64.Vec3) {

	pc.SetLookAt(v)

	caor := pc.GetViewDir().Dot(v)
	aor := mgl64.RadToDeg(float64(math.Acos(float64(caor))))
	axis := pc.GetViewDir().Cross(v).Normalize()

	if pc.GetDView() {
		pc.RotateAxis(aor, axis)
	}

	//	//fmt.Println("Rotate by %f on axis %v\n", aor, axis)
	//os.Exit(1)

	if !pc.GetDView() {
		pc.SetDView(true)
		pc.SetDefaultView(pc.GetLookAt())
	}
}

func (pc *PerspCameraData) RotateAxis(a float64, axis mgl64.Vec3) {
	r := mgl64.DegToRad(a)

	m := mgl64.HomogRotate3D(r, axis)
	//pc.Position = mgl64.TransformCoordinate(pc.Position, m)
	pc.SetUpDir(mgl64.TransformNormal(pc.GetUpDir(), m).Normalize())
	pc.SetRightDir(mgl64.TransformNormal(pc.GetRightDir(), m).Normalize())
	pc.SetViewDir(mgl64.TransformNormal(pc.GetViewDir(), m).Normalize())

	//      //fmt.Printf("After rotate of %f, viewdir is %f, %f, %f (total %f)\n", a, pc.ViewDir[0], pc.ViewDir[1], pc.ViewDir[2], pc.ViewDir[0]+pc.ViewDir[1]+pc.ViewDir[2])
}

type CamData struct {
	P          mgl64.Vec3
	U          mgl64.Vec3
	R          mgl64.Vec3
	V          mgl64.Vec3
	L          bool
	Z          float64
	RX, RY, RZ float64
}

func (pc *PerspCameraData) String() string {

	var c CamData

	c.P = pc.GetPosition()
	c.U = pc.GetUpDir()
	c.R = pc.GetRightDir()
	c.V = pc.GetViewDir()
	c.Z = pc.GetZoom()
	c.RX, c.RY, c.RZ = pc.GetRotatedX(), pc.GetRotatedY(), pc.GetRotatedZ()
	c.L = pc.GetLockView()

	b, _ := json.Marshal(&c)

	return string(b)
}

func (pc *PerspCameraData) FromString(s string) {

	var c CamData

	_ = json.Unmarshal([]byte(s), &c)

	pc.SetPosition(c.P)
	pc.SetUpDir(c.U)
	pc.SetRightDir(c.R)
	pc.SetViewDir(c.V)
	pc.SetLockView(c.L)
	pc.SetZoom(c.Z)
	pc.SetRotatedX(c.RX)
	pc.SetRotatedY(c.RY)
	pc.SetRotatedZ(c.RZ)

}

func (pc *PerspCameraData) GetState() ([]float64, []float64, []float64) {

	scale := mgl64.Scale3D(pc.GetZoom(), pc.GetZoom(), 1)
	modelview := mgl64.LookAtV(pc.GetPosition(), pc.GetLookAt(), pc.GetUpDir())
	persp := mgl64.Perspective(pc.GetFOV(), pc.GetAspect(), pc.GetNear(), pc.GetFar())

	return scale[:], modelview[:], persp[:]
}
