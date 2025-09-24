package types

import (
	"paleotronic.com/fmt"
	"testing"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types/glmath"
)

func dumpMat(l string, f []float64) {
	fmt.Println(l)
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			fmt.Printf(" %.6f", f[c*4+r])
		}
		fmt.Println()
	}
}

func dumpState(oc *OrbitCameraData, pc *PerspCameraData) {

	os, om, op := oc.GetState()
	ps, pm, pp := pc.GetState()

	fmt.Println("===========================================================================")
	fmt.Println("OrbitCamera:")
	dumpMat("Scale", os)
	dumpMat("ModelView", om)
	dumpMat("Perspective", op)
	fmt.Println("---------------------------------------------------------------------------")
	fmt.Println("PerspCamera:")
	dumpMat("Scale", ps)
	dumpMat("ModelView", pm)
	dumpMat("Perspective", pp)

}

func TestCurrentCamera(t *testing.T) {

	mm := memory.NewMemoryMap()

	pc := NewPerspCamera(
		mm,
		0,
		memory.OCTALYZER_MAPPED_CAM_BASE+(0+1)*memory.OCTALYZER_MAPPED_CAM_SIZE,
		60,
		float64(CRTASPECT),
		12750,
		14750,
		mgl64.Vec3{CWIDTH / 2, CHEIGHT / 2, 0},
	)
	//fxcam[i].ViewDir = mgl64.Vec3{CWIDTH / 2, CHEIGHT / 2, 0}
	pc.SetPos(CWIDTH/2, CHEIGHT/2, CDIST*GFXMULT)
	pc.SetLookAt(mgl64.Vec3{CWIDTH / 2, CHEIGHT / 2, 0})
	pc.SetZoom(GFXMULT)

	oc := NewOrbitCamera(mm, 0, memory.OCTALYZER_MAPPED_CAM_BASE, 60, CRTASPECT, 12750, 14750,
		&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, CDIST * GFXMULT},
		&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, 0},
		&glmath.Vector3{0, 0, 0},
	)
	oc.SetPos(CWIDTH/2, CHEIGHT/2, CDIST*GFXMULT)
	oc.SetTarget(&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, 0})
	oc.SetZoom(GFXMULT)

	dumpState(oc, pc)

}

func TestNewCamera(t *testing.T) {

	mm := memory.NewMemoryMap()

	pc := NewOrbitCamera(mm, 0, memory.OCTALYZER_MAPPED_CAM_BASE, 60, 1.33, 12000, 15000,
		&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, CDIST * GFXMULT},
		&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, 0},
		&glmath.Vector3{0, 0, 0},
	)
	pc.computeMatrix()
	pc.SetPos(CWIDTH/2, CHEIGHT/2, CDIST*GFXMULT)
	pc.SetTarget(&glmath.Vector3{CWIDTH / 2, CHEIGHT / 2, 0})
	pc.Move(-50, 10, 100)
	pc.Orbit(5, 0)

	pc.ResetALL()
	pc.computeMatrix()

	fmt.Println(pc.GetMatrix())
}
