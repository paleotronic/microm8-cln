package settings

const (
	CHEIGHT   = 960
	CDIST     = 854
	GFXMULT   = 16
	PASPECT   = 1.777777778
	CRTASPECT = 1.458
	CWIDTH    = CRTASPECT * CHEIGHT
)

type CameraDefaults struct {
	FOV       float64
	Near      float64
	Far       float64
	Aspect    float64
	Eye       [3]float64
	Target    [3]float64
	Angle     [3]float64
	Zoom      float64
	PivotLock bool
}

var CameraTextDefaults = CameraDefaults{
	FOV:    60,
	Near:   800,
	Far:    5000,
	Aspect: CRTASPECT,
	Eye:    [3]float64{CWIDTH / 2, CHEIGHT / 2, CDIST},
	Target: [3]float64{CWIDTH / 2, CHEIGHT / 2, 0},
	Angle:  [3]float64{0, 0, 0},
	Zoom:   1,
}

var CameraOSDDefaults = CameraDefaults{
	FOV:    60,
	Near:   800,
	Far:    5000,
	Aspect: 1.77778,
	Eye:    [3]float64{CWIDTH / 2, CHEIGHT / 2, CDIST},
	Target: [3]float64{CWIDTH / 2, CHEIGHT / 2, 0},
	Angle:  [3]float64{0, 0, 0},
	Zoom:   1,
}

var CameraGFXDefaults = CameraDefaults{
	FOV:       60,
	Near:      12750,
	Far:       16000,
	Aspect:    CRTASPECT,
	Eye:       [3]float64{CWIDTH / 2, CHEIGHT / 2, CDIST * GFXMULT},
	Target:    [3]float64{CWIDTH / 2, CHEIGHT / 2, 0},
	Angle:     [3]float64{0, 0, 0},
	Zoom:      GFXMULT,
	PivotLock: true,
}

var CameraGFXBGDefaults = CameraDefaults{
	FOV:       60,
	Near:      12750,
	Far:       14750,
	Aspect:    CRTASPECT,
	Eye:       [3]float64{CWIDTH / 2, CHEIGHT / 2, CDIST * GFXMULT},
	Target:    [3]float64{CWIDTH / 2, CHEIGHT / 2, 0},
	Angle:     [3]float64{0, 0, 0},
	Zoom:      GFXMULT,
	PivotLock: true,
}

var CameraVectorDefaults = CameraDefaults{
	FOV:       60,
	Near:      800,
	Far:       30000,
	Aspect:    CRTASPECT,
	Eye:       [3]float64{CWIDTH / 2, CHEIGHT / 2, CDIST * GFXMULT},
	Target:    [3]float64{CWIDTH / 2, CHEIGHT / 2, 0},
	Angle:     [3]float64{0, 0, 0},
	Zoom:      GFXMULT,
	PivotLock: true,
}

var GFXCameraDefaults = [8]CameraDefaults{
	CameraGFXDefaults,
	CameraVectorDefaults,
	CameraGFXDefaults,
	CameraGFXDefaults,
	CameraGFXDefaults,
	CameraGFXDefaults,
	CameraGFXDefaults,
	CameraGFXBGDefaults,
}

var CameraInitDefaults = map[int]CameraDefaults{}
