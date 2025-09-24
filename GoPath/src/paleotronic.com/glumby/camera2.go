package glumby

import (
	"encoding/json"
	"math"
	"paleotronic.com/fmt"

	"paleotronic.com/log"
	//"os"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/types"
	"paleotronic.com/gl"
	"paleotronic.com/utils"
)

type CRect struct {
	L, R, T, B float64
}

type PRect struct {
	X0, Y0, X1, Y1 float64
}

func (p *PRect) ApplyTo(master *PRect, pos types.LayerPosMod) PRect {
	xd := master.X1 - master.X0
	yd := master.Y1 - master.Y0

	return PRect{
		(p.X0+pos.XPercent)*xd + master.X0,
		(p.Y0+pos.YPercent)*yd + master.Y0,
		(p.X1+pos.XPercent)*xd + master.X0,
		(p.Y1+pos.YPercent)*yd + master.Y0,
	}
}

type PCamera struct {
	Position mgl64.Vec3
	ViewDir  mgl64.Vec3
	UpDir    mgl64.Vec3
	RightDir mgl64.Vec3
	//
	FOV    float64
	Aspect float64
	Near   float64
	Far    float64
	// Defaults
	defaultPos  mgl64.Vec3
	defaultView mgl64.Vec3
	dPos, dView bool
	LockView    bool
	Zoom        float64
	DollyRate   float64
	DollyPos    []mgl64.Vec3
	DollyTarget []mgl64.Vec3
	ViewPort    PRect
	ShakeFrames int
	ShakeMax    float64
}

func NewPCamera(fov, aspect, near, far float64) *PCamera {
	this := &PCamera{}
	this.Position = mgl64.Vec3{0, 0, 0}
	this.ViewDir = mgl64.Vec3{0, 0, -1}
	this.UpDir = mgl64.Vec3{0, 1, 0}
	this.RightDir = mgl64.Vec3{1, 0, 0}

	this.FOV = fov
	this.Aspect = aspect
	this.Near = near
	this.Far = far
	this.Zoom = 1
	this.LockView = false

	this.DollyPos = make([]mgl64.Vec3, 0)
	this.DollyTarget = make([]mgl64.Vec3, 0)
	this.ViewPort = PRect{0, 0, 1, 1}

	return this
}

func (pc *PCamera) ToJSON() string {

	b, _ := json.Marshal(pc)

	return string(b)
}

func (pc *PCamera) FromJSON(s string) {

	oldvp := pc.ViewPort

	b := []byte(s)
	_ = json.Unmarshal(b, pc)

	pc.ViewPort = oldvp // keep this -- its important

}

func (pc *PCamera) Command(data []uint64) {

	if len(data) == 0 {
		return
	}

	switch types.CameraCommand(data[0]) {
	case types.CC_Shake:
		if len(data) < 3 {
			return
		}
		pc.ShakeFrames = int(data[1])
		pc.ShakeMax = float64(data[2])
	case types.CC_JSON:
		if len(data) < 2 {
			return
		}

		log.Printf("a %d, s %d\n", len(data), data[1])

		l := int(data[1])
		chunk := data[2 : 2+l]
		ba := make([]byte, len(chunk))
		for i, v := range chunk {
			ba[i] = byte(v)
		}
		log.Println(string(ba))
		pc.FromJSON(string(ba))
		//os.Exit(0)
	case types.CC_LookAt:
		if len(data) < 4 {
			return
		}
		x := types.Uint2Float64(data[1])
		y := types.Uint2Float64(data[2])
		z := types.Uint2Float64(data[3])
		pc.SetLookAt(x, y, z)
	case types.CC_AbsPos:
		if len(data) < 4 {
			return
		}
		x := types.Uint2Float64(data[1])
		y := types.Uint2Float64(data[2])
		z := types.Uint2Float64(data[3])
		pc.SetPos(x, y, z)
	case types.CC_RelPos:
		if len(data) < 4 {
			return
		}
		x := types.Uint2Float64(data[1])
		y := types.Uint2Float64(data[2])
		z := types.Uint2Float64(data[3])
		pc.Move(x, y, z)
	case types.CC_RotX:
		if len(data) < 2 {
			return
		}
		angle := types.Uint2Float64(data[1])
		pc.Rotate3DX(angle)
	case types.CC_RotY:
		if len(data) < 2 {
			return
		}
		angle := types.Uint2Float64(data[1])
		pc.Rotate3DY(angle)
	case types.CC_RotZ:
		if len(data) < 2 {
			return
		}
		angle := types.Uint2Float64(data[1])
		pc.Rotate3DZ(angle)
	case types.CC_RotateAxis:
		if len(data) < 5 {
			return
		}
		angle := types.Uint2Float64(data[1])
		x := types.Uint2Float64(data[2])
		y := types.Uint2Float64(data[3])
		z := types.Uint2Float64(data[4])
		pc.RotateAxis(angle, mgl64.Vec3{x, y, z})
	case types.CC_Orbit:
		if len(data) < 3 {
			return
		}
		pitch := types.Uint2Float64(data[1])
		yaw := types.Uint2Float64(data[2])
		pc.Orbit(yaw, pitch)
	case types.CC_ResetPos:
		pc.ResetPosition()
	case types.CC_ResetLookAt:
		pc.ResetLookAt()
	case types.CC_ResetAngle:
		pc.UpDir = mgl64.Vec3{0, 1, 0}
		pc.RightDir = mgl64.Vec3{1, 0, 0}
	case types.CC_ResetAll:
		pc.UpDir = mgl64.Vec3{0, 1, 0}
		pc.RightDir = mgl64.Vec3{1, 0, 0}
		pc.ResetLookAt()
		pc.ResetPosition()
		pc.SetZoom(1)
	case types.CC_PivotLock:
		pc.SetPivotLock(true)
	case types.CC_PivotUnlock:
		pc.SetPivotLock(false)
	case types.CC_Zoom:
		if len(data) < 2 {
			return
		}
		zoom := types.Uint2Float64(data[1])
		pc.SetZoom(zoom)
	case types.CC_Dolly:
		if len(data) < 2 {
			return
		}
		dolly := types.Uint2Float64(data[1])
		pc.SetDollyRate(dolly)
	}

}

func (pc *PCamera) ResetALL() {
	pc.UpDir = mgl64.Vec3{0, 1, 0}
	pc.RightDir = mgl64.Vec3{1, 0, 0}
	pc.ResetLookAt()
	pc.ResetPosition()
	pc.SetZoom(1)
}

func mat64to32(in mgl64.Mat4) mgl32.Mat4 {
	var out mgl32.Mat4
	for i, v := range in {
		out[i] = float32(v)
	}
	return out
}

func (pc *PCamera) Apply() {

	// Dolly handling
	if len(pc.DollyPos) > 0 {
		v := pc.DollyPos[0]
		pc.DollyPos = pc.DollyPos[1:]
		pc.Position = v
	}

	if len(pc.DollyTarget) > 0 {
		v := pc.DollyTarget[0]
		pc.DollyTarget = pc.DollyTarget[1:]
		pc.ViewDir = v
	}

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()

	// perspective
	pmat := mat64to32(mgl64.Perspective(mgl64.DegToRad(pc.FOV), pc.Aspect, pc.Near, pc.Far))
	gl.MultMatrixf(&pmat[0])

	smat := mat64to32(mgl64.Scale3D(pc.Zoom, pc.Zoom, 1.0))
	gl.MultMatrixf(&smat[0])

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	mvmat := mat64to32(mgl64.LookAtV(pc.Position, pc.ViewDir, pc.UpDir))
	gl.MultMatrixf(&mvmat[0])

}

func (pc *PCamera) ApplyWindow(master *PRect, pos types.LayerPosMod) {
	// pc.ViewPort is a percent based view of the full 16:9 space

	//log2.Printf("pcamera viewport = %+v", *master)

	actual := pc.ViewPort.ApplyTo(master, pos)

	if pc.ShakeFrames > 0 {
		variance := float64(utils.Random() * pc.ShakeMax)
		actual.X0 -= variance
		actual.Y0 -= variance
		actual.X1 -= variance
		actual.Y1 -= variance
		pc.ShakeFrames--
	}

	r := (actual.X1 - actual.X0) / (actual.Y1 - actual.Y0)
	fmt.Printf("===> ViewPort aspect is %.5f\n", r)

	gl.Viewport(int32(actual.X0), int32(actual.Y0), int32(actual.X1-actual.X0), int32(actual.Y1-actual.Y0))
	pc.Apply()
}

func (pc *PCamera) SetPos(x, y, z float64) {
	pc.SetPosV(mgl64.Vec3{x, y, z})
}

func (pc *PCamera) SetPosV(v mgl64.Vec3) {
	pc.Position = v
	if !pc.dPos {
		pc.dPos = true
		pc.defaultPos = v
	}
}

func (pc *PCamera) SetPosition(x, y, z float64) {
	target := mgl64.Vec3{x, y, z}

	if len(pc.DollyPos) == 0 {
		pc.CalculateMotion(pc.Position, target)
	} else {
		pc.CalculateMotion(pc.DollyPos[len(pc.DollyPos)-1], target)
	}
}

func (pc *PCamera) SetTargetPosition(x, y, z float64) {
	target := mgl64.Vec3{x, y, z}

	if len(pc.DollyTarget) == 0 {
		pc.CalculateTargetMotion(pc.ViewDir, target)
	} else {
		pc.CalculateTargetMotion(pc.DollyTarget[len(pc.DollyTarget)-1], target)
	}
}

func (pc *PCamera) SetLookAt(x, y, z float64) {
	pc.SetLookAtV(mgl64.Vec3{x, y, z})
}

func (pc *PCamera) SetLookAtV(v mgl64.Vec3) {
	pc.ViewDir = v
	if !pc.dView {
		pc.dView = true
		pc.defaultView = v
	}
}

func (pc *PCamera) Rotate3DX(a float64) {

	r := float64(float64(a) * (math.Pi / 180))

	m := mgl64.HomogRotate3DX(r)
	//pc.Position = mgl64.TransformCoordinate(pc.Position, m)
	pc.UpDir = mgl64.TransformNormal(pc.UpDir, m).Normalize()
	pc.RightDir = mgl64.TransformNormal(pc.RightDir, m).Normalize()
	//pc.ViewDir = mgl64.TransformNormal(pc.ViewDir, m)

	//pc.RotateAxis(a, pc.RightDir)

}

func (pc *PCamera) Rotate3DY(a float64) {

	r := float64(float64(a) * (math.Pi / 180))

	m := mgl64.HomogRotate3DY(r)
	//pc.Position = mgl64.TransformCoordinate(pc.Position, m)
	pc.UpDir = mgl64.TransformNormal(pc.UpDir, m).Normalize()
	pc.RightDir = mgl64.TransformNormal(pc.RightDir, m).Normalize()
	//pc.ViewDir = mgl64.TransformNormal(pc.ViewDir, m)

	//pc.RotateAxis(a, pc.UpDir)

}

func (pc *PCamera) Rotate3DZ(a float64) {

	r := float64(float64(a) * (math.Pi / 180))

	m := mgl64.HomogRotate3DZ(r)
	//pc.Position = mgl64.TransformCoordinate(pc.Position, m)
	pc.UpDir = mgl64.TransformNormal(pc.UpDir, m).Normalize()
	pc.RightDir = mgl64.TransformNormal(pc.RightDir, m).Normalize()
	//pc.ViewDir = mgl64.TransformNormal(pc.ViewDir, m)

	j := pc.ToJSON()
	log.Println(j)

	//pc.RotateAxis(a, pc.ViewDir)

}

func (pc *PCamera) RotateAxis(a float64, axis mgl64.Vec3) {
	r := mgl64.DegToRad(a)

	m := mgl64.HomogRotate3D(r, axis)
	//pc.Position = mgl64.TransformCoordinate(pc.Position, m)
	pc.UpDir = mgl64.TransformNormal(pc.UpDir, m).Normalize()
	pc.RightDir = mgl64.TransformNormal(pc.RightDir, m).Normalize()
	pc.ViewDir = mgl64.TransformNormal(pc.ViewDir, m).Normalize()

	//      //fmt.Printf("After rotate of %f, viewdir is %f, %f, %f (total %f)\n", a, pc.ViewDir[0], pc.ViewDir[1], pc.ViewDir[2], pc.ViewDir[0]+pc.ViewDir[1]+pc.ViewDir[2])
}

func (pc *PCamera) ResetPosition() {
	if pc.dPos {
		pc.Position = pc.defaultPos
	}
}

func (pc *PCamera) ResetLookAt() {
	if pc.dView {
		pc.ViewDir = pc.defaultView
	}
}

func (pc *PCamera) Move(dx, dy, dz float64) {

	m := mgl64.Translate3D(dx, dy, dz)

	if !pc.LockView {
		vtarget := mgl64.TransformCoordinate(pc.ViewDir, m)
		pc.SetTargetPosition(vtarget[0], vtarget[1], vtarget[2])
	}

	target := mgl64.TransformCoordinate(pc.Position, m)

	if len(pc.DollyPos) == 0 {
		pc.CalculateMotion(pc.Position, target)
	} else {
		pc.CalculateMotion(pc.DollyPos[len(pc.DollyPos)-1], target)
	}

}

func (pc *PCamera) SetZoom(v float64) {
	if v == 0 {
		v = 0.00001
	}
	pc.Zoom = v
}

func (pc *PCamera) SetDollyRate(v float64) {
	pc.DollyRate = v
}

func (pc *PCamera) SetPivotLock(b bool) {
	pc.LockView = true
}

func abs(v float64) float64 {
	return math.Abs(v)
}

func (pc *PCamera) CalculateMotion(start, end mgl64.Vec3) {

	if pc.DollyRate == 0 {
		pc.DollyPos = []mgl64.Vec3{end}
		return
	}

	if start == end {
		return
	}

	// differences on each axis
	dx, dy, dz := end[0]-start[0], end[1]-start[1], end[2]-start[2]

	// percent of a whole
	dxp, dyp, dzp := dx/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dy/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dz/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz)))

	xdiff, ydiff, zdiff := pc.DollyRate*dxp, pc.DollyRate*dyp, pc.DollyRate*dzp

	xx, yy, zz := start[0], start[1], start[2]

	for (xx != end[0]) || (yy != end[1]) || (zz != end[2]) {

		xa := xdiff
		if abs(xa) > abs(end[0]-xx) {
			xa = end[0] - xx
		}

		ya := ydiff
		if abs(ya) > abs(end[1]-yy) {
			ya = end[1] - yy
		}

		za := zdiff
		if abs(za) > abs(end[2]-zz) {
			za = end[2] - zz
		}

		xx += xa
		yy += ya
		zz += za

		pc.DollyPos = append(pc.DollyPos, mgl64.Vec3{xx, yy, zz})

	}

}

func (pc *PCamera) CalculateTargetMotion(start, end mgl64.Vec3) {

	if pc.DollyRate == 0 {
		pc.DollyTarget = []mgl64.Vec3{end}
		return
	}

	if start == end {
		return
	}

	// differences on each axis
	dx, dy, dz := end[0]-start[0], end[1]-start[1], end[2]-start[2]

	// percent of a whole
	dxp, dyp, dzp := dx/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dy/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dz/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz)))

	xdiff, ydiff, zdiff := pc.DollyRate*dxp, pc.DollyRate*dyp, pc.DollyRate*dzp

	xx, yy, zz := start[0], start[1], start[2]

	for (xx != end[0]) || (yy != end[1]) || (zz != end[2]) {

		xa := xdiff
		if abs(xa) > abs(end[0]-xx) {
			xa = end[0] - xx
		}

		ya := ydiff
		if abs(ya) > abs(end[1]-yy) {
			ya = end[1] - yy
		}

		za := zdiff
		if abs(za) > abs(end[2]-zz) {
			za = end[2] - zz
		}

		xx += xa
		yy += ya
		zz += za

		pc.DollyTarget = append(pc.DollyTarget, mgl64.Vec3{xx, yy, zz})

	}

}

// Fixed orbit function
func (pc *PCamera) CalculateOrbit(yaw, pitch float64) mgl64.Vec3 {

	src := pc.Position
	target := pc.ViewDir

	camFocusVector := src.Sub(target)

	r1 := mgl64.HomogRotate3D(mgl64.DegToRad(yaw), pc.UpDir)
	r2 := mgl64.HomogRotate3D(mgl64.DegToRad(pitch), pc.RightDir)

	camFocusVector = mgl64.TransformCoordinate(camFocusVector, r1)
	camFocusVector = mgl64.TransformCoordinate(camFocusVector, r2)

	p := pc.ViewDir.Add(camFocusVector)

	pc.UpDir = mgl64.TransformNormal(pc.UpDir, r1)
	pc.UpDir = mgl64.TransformNormal(pc.UpDir, r2)

	pc.RightDir = mgl64.TransformNormal(pc.RightDir, r1)
	pc.RightDir = mgl64.TransformNormal(pc.RightDir, r2)

	pc.ViewDir = mgl64.TransformNormal(pc.ViewDir, r1)
	pc.ViewDir = mgl64.TransformNormal(pc.ViewDir, r2)

	return p
}

func (pc *PCamera) Orbit(yaw, pitch float64) {
	p := pc.CalculateOrbit(yaw, pitch)
	//pc.SetPosition( p[0], p[1], p[2] )
	pc.SetPosV(p)
}
