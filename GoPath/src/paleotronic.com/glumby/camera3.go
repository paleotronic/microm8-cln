package glumby

import (
	"encoding/json"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/core/types/glmath"
	"paleotronic.com/gl"
	"paleotronic.com/utils"
)

type PerspCamera struct {
	*types.OrbitCameraData
	DollyPos      []glmath.Vector3 `json:"DP,omitempty"`
	DollyTarget   []glmath.Vector3 `json:"DT,omitempty"`
	ViewPort      PRect            `json:"-"`
	PixelViewPort PRect            `json:"-"`
	ZoomViewport  bool
}

func NewPerspCamera(RAM *memory.MemoryMap, slotid int, baseaddr int, fov, aspect, near, far float64, eye, target, angle *glmath.Vector3) *PerspCamera {
	this := &PerspCamera{}

	mcb := memory.NewMemoryControlBlock(RAM, slotid, false)

	baseglobal := RAM.MEMBASE(slotid) + baseaddr
	baseglobalend := baseglobal + memory.OCTALYZER_MAPPED_CAM_SIZE
	mcb.Add(RAM.Data[slotid][baseglobal:baseglobalend], baseglobal)

	this.OrbitCameraData = types.NewOrbitCamera(RAM, slotid, baseaddr, fov, aspect, near, far, eye, target, angle)

	this.SetPosition(eye)
	// this.SetViewDir(&glmath.Vector3{0, 0, 1})
	// this.SetUpDir(&glmath.Vector3{0, 1, 0})
	// this.SetRightDir(&glmath.Vector3{1, 0, 0})
	this.SetTarget(target)

	this.SetFOV(fov)
	this.SetAspect(aspect)
	this.SetNear(near)
	this.SetFar(far)
	this.SetZoom(1)
	this.SetLockView(false)

	this.DollyPos = make([]glmath.Vector3, 0)
	this.DollyTarget = make([]glmath.Vector3, 0)
	this.ViewPort = this.CalcReferenceViewport(types.PASPECT, aspect)

	return this
}

func NewPerspCameraWithConfig(RAM *memory.MemoryMap, slotid int, baseaddr int, config settings.CameraDefaults) *PerspCamera {
	this := &PerspCamera{}

	mcb := memory.NewMemoryControlBlock(RAM, slotid, false)

	baseglobal := RAM.MEMBASE(slotid) + baseaddr
	baseglobalend := baseglobal + memory.OCTALYZER_MAPPED_CAM_SIZE
	mcb.Add(RAM.Data[slotid][baseglobal:baseglobalend], baseglobal)

	this.OrbitCameraData = types.NewOrbitCamera(
		RAM,
		slotid,
		baseaddr,
		config.FOV,
		config.Aspect,
		config.Near,
		config.Far,
		&glmath.Vector3{config.Eye[0], config.Eye[1], config.Eye[2]},
		&glmath.Vector3{config.Target[0], config.Target[1], config.Target[2]},
		&glmath.Vector3{config.Angle[0], config.Angle[1], config.Angle[2]},
	)

	this.SetPosition(&glmath.Vector3{config.Eye[0], config.Eye[1], config.Eye[2]})
	this.SetTarget(&glmath.Vector3{config.Target[0], config.Target[1], config.Target[2]})

	this.SetFOV(config.FOV)
	this.SetAspect(config.Aspect)
	this.SetNear(config.Near)
	this.SetFar(config.Far)
	this.SetZoom(config.Zoom)
	this.SetLockView(config.PivotLock)

	// save config for later
	settings.CameraInitDefaults[baseaddr] = config

	this.DollyPos = make([]glmath.Vector3, 0)
	this.DollyTarget = make([]glmath.Vector3, 0)
	this.ViewPort = this.CalcReferenceViewport(types.PASPECT, config.Aspect)

	return this
}

func (pc *PerspCamera) CalcReferenceViewport(wide, crt float64) PRect {

	var rh float64 = 900
	widthWide := rh * wide
	widthCrt := rh * crt
	sidepad := (widthWide - widthCrt) / 2
	padprop := sidepad / widthWide

	return PRect{padprop, 0, 1 - padprop, 1}
}

func (pc *PerspCamera) ToJSON() string {

	b, _ := json.Marshal(pc)

	return string(b)
}

func (pc *PerspCamera) FromJSON(s string) {

	oldvp := pc.ViewPort

	b := []byte(s)
	_ = json.Unmarshal(b, pc)

	pc.ViewPort = oldvp // keep this -- its important

}

func (pc *PerspCamera) Command(data []uint64, index int, ram *memory.MemoryMap) {

	if len(data) == 0 {
		return
	}

	switch types.CameraCommand(data[0]) {
	case types.CC_Shake:
		if len(data) < 3 {
			return
		}
		pc.SetShakeFrames(int(data[1]))
		pc.SetShakeMax(float64(data[2]))
	case types.CC_GetJSONR:
		return
	case types.CC_GetJSON:

		s := pc.ToJSON()

		ram.WriteInterpreterMemorySilent(index, memory.OCTALYZER_CAMERA_GFX_BASE+1, uint64(len(s)))

		for i, ch := range s {
			ram.WriteInterpreterMemorySilent(index, memory.OCTALYZER_CAMERA_GFX_BASE+2+i, uint64(ch))
		}

		ram.WriteInterpreterMemorySilent(index, memory.OCTALYZER_CAMERA_GFX_BASE+0, uint64(types.CC_GetJSONR))

	case types.CC_JSON:
		if len(data) < 2 {
			return
		}

		//		log.Printf("a %d, s %d\n", len(data), data[1])

		l := int(data[1])
		chunk := data[2 : 2+l]
		ba := make([]byte, len(chunk))
		for i, v := range chunk {
			ba[i] = byte(v)
		}
		//		log.Println(string(ba))
		pc.FromJSON(string(ba))
		//os.Exit(0)
	case types.CC_LookAt:
		if len(data) < 4 {
			return
		}
		x := types.Uint2Float64(data[1])
		y := types.Uint2Float64(data[2])
		z := types.Uint2Float64(data[3])
		pc.SetTarget(&glmath.Vector3{x, y, z})
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
		// angle := types.Uint2Float64(data[1])
		// x := types.Uint2Float64(data[2])
		// y := types.Uint2Float64(data[3])
		// z := types.Uint2Float64(data[4])
		// pc.SetRotation(angle, glmath.Vector3{x, y, z})
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
		pc.SetRotation(&glmath.Vector3{0, 0, 0})
	case types.CC_ResetAll:
		pc.ResetALL()
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

func (this *PerspCamera) ResetALL() {
	this.ResetPosition()
	this.ResetOrientation()
	this.ResetLookAt()
	this.ResetZoom()

	this.DollyPos = make([]glmath.Vector3, 0)
	this.DollyTarget = make([]glmath.Vector3, 0)
	this.SetAspect(types.CRTASPECT)
	this.ViewPort = this.CalcReferenceViewport(types.PASPECT, this.GetAspect())
}

func getClipPlanesFromMVP(mat mgl32.Mat4) [6]mgl32.Vec4 {

	var mPlanes [6]mgl32.Vec4

	//left
	mPlanes[LEFT_PLANE][X] = mat.Row(0)[3] + mat.Row(0)[0]
	mPlanes[LEFT_PLANE][Y] = mat.Row(1)[3] + mat.Row(1)[0]
	mPlanes[LEFT_PLANE][Z] = mat.Row(2)[3] + mat.Row(2)[0]
	mPlanes[LEFT_PLANE][W] = mat.Row(3)[3] + mat.Row(3)[0]

	//right
	mPlanes[RIGHT_PLANE][X] = mat.Row(0)[3] - mat.Row(0)[0]
	mPlanes[RIGHT_PLANE][Y] = mat.Row(1)[3] - mat.Row(1)[0]
	mPlanes[RIGHT_PLANE][Z] = mat.Row(2)[3] - mat.Row(2)[0]
	mPlanes[RIGHT_PLANE][W] = mat.Row(3)[3] - mat.Row(3)[0]

	//bottom
	mPlanes[BOTTOM_PLANE][X] = mat.Row(0)[3] + mat.Row(0)[1]
	mPlanes[BOTTOM_PLANE][Y] = mat.Row(1)[3] + mat.Row(1)[1]
	mPlanes[BOTTOM_PLANE][Z] = mat.Row(2)[3] + mat.Row(2)[1]
	mPlanes[BOTTOM_PLANE][W] = mat.Row(3)[3] + mat.Row(3)[1]

	//top
	mPlanes[TOP_PLANE][X] = mat.Row(0)[3] - mat.Row(0)[1]
	mPlanes[TOP_PLANE][Y] = mat.Row(1)[3] - mat.Row(1)[1]
	mPlanes[TOP_PLANE][Z] = mat.Row(2)[3] - mat.Row(2)[1]
	mPlanes[TOP_PLANE][W] = mat.Row(3)[3] - mat.Row(3)[1]

	return mPlanes

}

const (
	LEFT_PLANE int = iota
	RIGHT_PLANE
	TOP_PLANE
	BOTTOM_PLANE
	NEAR_PLANE
	FAR_PLANE
)

const (
	X int = iota
	Y
	Z
	W
)

func (pc *PerspCamera) ScreenPosToWorldPos(win *glmath.Vector3, waspect float64, fov float64, initialX, initialY, width, height int) (*glmath.Vector3, error) {

	projection := glmath.Perspective(fov, waspect, pc.GetNear(), pc.GetFar())
	modelview := pc.OrbitCameraData.GetMatrix()

	return glmath.UnProject(win, modelview, projection, initialX, initialY, width, height)

}

func (pc *PerspCamera) Apply(cr CRect, waspect float64, fov float64, pos types.LayerPosMod, mult float64) {

	// Update viewport
	pc.ViewPort = pc.CalcReferenceViewport(types.PASPECT, pc.GetAspect())

	// Dolly handling
	if len(pc.DollyPos) > 0 {
		v := pc.DollyPos[0]
		pc.DollyPos = pc.DollyPos[1:]
		pc.SetPosition(&v)
	}

	if len(pc.DollyTarget) > 0 {
		v := pc.DollyTarget[0]
		pc.DollyTarget = pc.DollyTarget[1:]
		pc.SetTarget(&v)
	}

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()

	// perspective
	pmat := glmath.Perspective(fov, waspect, pc.GetNear(), pc.GetFar()).ToGLFloat()
	gl.MultMatrixf(&pmat[0])

	if pos.XPercent != 0 || pos.YPercent != 0 {

		MW := (cr.R - cr.L) * mult
		MH := (cr.T - cr.B) * mult

		gl.Translatef(float32(MW*pos.XPercent), float32(MH*pos.YPercent), 0)
	}

	px, py := pc.GetPanX(), pc.GetPanY()
	if px != 0 || py != 0 {
		gl.Translatef(float32(px), float32(py), 0)
	}

	//fmt.Printf("Calculated frustrum: %v\n", cr)
	// p := getClipPlanesFromMVP(pmat)
	// fmt.Printf("L %f, R %f, T %f, B %f\n", p[LEFT_PLANE][X]*p[LEFT_PLANE][W], p[RIGHT_PLANE][X], p[TOP_PLANE][Y], p[BOTTOM_PLANE][Y])

	// gl.Frustum(
	// 	cr.L,
	// 	cr.R,
	// 	cr.B,
	// 	cr.T,
	// 	pc.GetNear(),
	// 	pc.GetFar(),
	// )

	smat := glmath.Scale(pc.GetZoom(), pc.GetZoom(), 1).ToGLFloat()
	gl.MultMatrixf(&smat[0])

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	mvmat := pc.OrbitCameraData.GetMatrix().ToGLFloat()
	gl.MultMatrixf(&mvmat[0])

}

func (pc *PerspCamera) ApplyWindow(master *PRect, pos types.LayerPosMod, cr CRect, waspect float64, fov float64, mult float64) {
	// pc.ViewPort is a percent based view of the full 16:9 space

	// Update viewport
	pc.ViewPort = pc.CalcReferenceViewport(types.PASPECT, pc.GetAspect())

	actual := pc.ViewPort.ApplyTo(master, pos)

	var variance float64

	if pc.GetShakeFrames() > 0 {
		variance = float64(utils.Random() * float64(pc.GetShakeMax()))
		// actual.X0 -= variance
		// actual.Y0 -= variance
		// actual.X1 -= variance
		// actual.Y1 -= variance
		pc.SetShakeFrames(pc.GetShakeFrames() - 1)
	}

	// w := actual.X1 - actual.X0
	// h := actual.Y1 - actual.Y0
	// if pc.GetZoom() > types.GFXMULT /*&& pc.ZoomViewport*/ {
	// 	z := (pc.GetZoom() / types.GFXMULT) * 1.3
	// 	cx := actual.X0 + (w / 2)
	// 	cy := actual.Y0 + (h / 2)

	// 	w = w * z
	// 	h = h * z

	// 	actual.X0 = cx - (w / 2)
	// 	actual.Y0 = cy - (h / 2)
	// }

	// if pos.XPercent != 0 || pos.YPercent != 0 {
	// 	fmt.Println("Vp")
	// 	gl.Viewport(int32(actual.X0), int32(actual.Y0), int32(w), int32(h))
	// }
	pos.XPercent += variance
	pos.YPercent += variance

	pc.Apply(cr, waspect, fov, pos, mult)

	pc.PixelViewPort = actual
}

func (pc *PerspCamera) SetTargetPosition(x, y, z float64) {
	target := glmath.Vector3{x, y, z}

	if len(pc.DollyTarget) == 0 {
		pc.CalculateTargetMotion(*pc.OrbitCameraData.GetForwardAxis(), target)
	} else {
		pc.CalculateTargetMotion(pc.DollyTarget[len(pc.DollyTarget)-1], target)
	}
}

func (pc *PerspCamera) Move(dx, dy, dz float64) {

	m := glmath.Translate(dx, dy, dz)

	if !pc.GetLockView() {
		vtarget := glmath.TransformVector3(pc.GetForwardAxis(), m)
		pc.SetTargetPosition(vtarget[0], vtarget[1], vtarget[2])
	}

	target := glmath.TransformVector3(pc.GetPosition(), m)

	if len(pc.DollyPos) == 0 {
		pc.CalculateMotion(*pc.GetPosition(), *target)
	} else {
		pc.CalculateMotion(pc.DollyPos[len(pc.DollyPos)-1], *target)
	}

}

func (pc *PerspCamera) CalculateMotion(start, end glmath.Vector3) {

	if pc.GetDollyRate() == 0 {
		pc.DollyPos = []glmath.Vector3{end}
		return
	}

	if start == end {
		return
	}

	// differences on each axis
	dx, dy, dz := end[0]-start[0], end[1]-start[1], end[2]-start[2]

	// percent of a whole
	dxp, dyp, dzp := dx/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dy/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dz/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz)))

	dr := pc.GetDollyRate()

	xdiff, ydiff, zdiff := dr*dxp, dr*dyp, dr*dzp

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

		pc.DollyPos = append(pc.DollyPos, glmath.Vector3{xx, yy, zz})

	}

}

func (pc *PerspCamera) CalculateTargetMotion(start, end glmath.Vector3) {

	if pc.GetDollyRate() == 0 {
		pc.DollyTarget = []glmath.Vector3{end}
		return
	}

	if start == end {
		return
	}

	// differences on each axis
	dx, dy, dz := end[0]-start[0], end[1]-start[1], end[2]-start[2]

	// percent of a whole
	dxp, dyp, dzp := dx/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dy/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz))), dz/float64(math.Abs(float64(dx))+math.Abs(float64(dy))+math.Abs(float64(dz)))

	dr := pc.GetDollyRate()

	xdiff, ydiff, zdiff := dr*dxp, dr*dyp, dr*dzp

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

		pc.DollyTarget = append(pc.DollyTarget, glmath.Vector3{xx, yy, zz})

	}

}

func cosf(r float64) float64 {
	return float64(math.Cos(float64(r)))
}

func sinf(r float64) float64 {
	return float64(math.Sin(float64(r)))
}
