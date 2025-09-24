// +build remint

package main

import (
	"flag"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof" //"github.com/pkg/profile"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"paleotronic.com/api"
	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/backend"
	"paleotronic.com/server/remoteapi"
)

const (
	PASPECT   = 1.777777778
	CRTASPECT = 1.333333333
)

var (
	ramActiveState [memory.OCTALYZER_NUM_INTERPRETERS]uint
	RAM            *memory.MemoryMap
	distance_z     float32
	cx, cy         int
	modeChanged    bool
	lastP0, lastP1 int
	filecache      map[[16]byte]*files.FilePack
	modifier       bool
	buffer         int

	pcam  [memory.OCTALYZER_NUM_INTERPRETERS]*types.PerspCameraData
	fxcam [memory.OCTALYZER_NUM_INTERPRETERS][memory.OCTALYZER_MAPPED_CAM_GFXCOUNT]*types.PerspCameraData

	FlipCase        bool
	ignoreKeyEvents bool
	reboot          bool
)

var targetHost = flag.String("backend", "localhost", "Host to connect to") // blah
var useport = flag.Int("port", 8580, "Remint service port")
var dataHost = flag.String("server", "paleotronic.com", "Server to connect to")
var versionDisplay = flag.Bool("version", false, "Show version and exit.")
var noUpdate = flag.Bool("no-update", false, "Dont check for updates.")
var trace6502 = flag.Bool("trace-cpu", false, "Trace 6502 access")
var prof = flag.Bool("profile-cpu", false, "Profile 6502 performance")
var heatmap = flag.Bool("heatmap", false, "Trace ROM access")
var benchmode = flag.Bool("bench", false, "Benchmark all slots")
var share [memory.OCTALYZER_NUM_INTERPRETERS]*bool
var shareport [memory.OCTALYZER_NUM_INTERPRETERS]*int
var ourIP string

func init() {

	settings.IsRemInt = true

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

		share[i] = flag.Bool(fmt.Sprintf("share-%d", i), true, "Share slot enabled")
		shareport[i] = flag.Int(fmt.Sprintf("sport-%d", i), 8580+i, "Share slot port")

	}

	runtime.LockOSThread()
	time.Sleep(50 * time.Millisecond)
}

func SetSlotAspect(index int, aspect float64) {
	pcam[index].SetAspect(aspect)
	for i := 0; i < 8; i++ {
		fxcam[index][i].SetAspect(aspect)
	}
}

func initBackend(r *memory.MemoryMap, nu bool) {
	backend.Run(r, PostCycleCallback)
}

func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func RoundTo5(v int64) int64 {

	z := float64(v)

	z = Round(z/5) * 5

	return int64(z)
}

func networkInit(host string) {
	s8webclient.NetworkInit(host)
}

// WaveMonkey is a callback...
func WaveMonkey(index int, data []float32, rate int) {

	d := time.Duration(500000000 * float32(len(data)) / float32(rate))
	time.Sleep(d)

}

func startSharing() {

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		if *share[i] {
			RAM.Share(i, backend.VPP, *shareport[i])
		}
	}

}

func stopBackend() {

}

func initCameras() {
	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

		pcam[i] = types.NewPerspCamera(
			RAM,
			i,
			memory.OCTALYZER_MAPPED_CAM_BASE,
			60,
			float64(CRTASPECT),
			1,
			8000,
			mgl64.Vec3{types.CWIDTH / 2, types.CHEIGHT / 2, 0},
		)

		pcam[i].SetPos(types.CWIDTH/2, types.CHEIGHT/2, types.CDIST)
		pcam[i].SetLookAt(mgl64.Vec3{types.CWIDTH / 2, types.CHEIGHT / 2, 0})

		for z := 0; z < memory.OCTALYZER_MAPPED_CAM_GFXCOUNT; z++ {
			fxcam[i][z] = types.NewPerspCamera(
				RAM,
				i,
				memory.OCTALYZER_MAPPED_CAM_BASE+(z+1)*memory.OCTALYZER_MAPPED_CAM_SIZE,
				60,
				float64(CRTASPECT),
				100,
				15000,
				mgl64.Vec3{types.CWIDTH / 2, types.CHEIGHT / 2, 0},
			)
			fxcam[i][z].SetPos(types.CWIDTH/2, types.CHEIGHT/2, types.CDIST*types.GFXMULT)
			fxcam[i][z].SetLookAt(mgl64.Vec3{types.CWIDTH / 2, types.CHEIGHT / 2, 0})
			fxcam[i][z].SetPivotLock(true)
			fxcam[i][z].SetZoom(types.GFXMULT)
		}

	}
}

func memoryInit() {
	RAM = memory.NewMemoryMap()

	RAM.InputToggle(-1)

	startSharing()

	initCameras()

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		RAM.IntSetDHGRRender(i, settings.VM_FLAT)
		RAM.IntSetHGRRender(i, settings.VM_FLAT)
		RAM.IntSetVideoTint(i, settings.VPT_NONE)
	}

	//go statusReporter()

	initBackend(RAM, *noUpdate)
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func getSlotStatus(slotid int) *remoteapi.SlotStatus {

	e := backend.ProducerMain.GetInterpreter(slotid)
	if e == nil {
		return nil
	}

	ss := e.GetMemoryMap().SlotShares[slotid]
	if ss == nil {
		return nil
	}

	s := &remoteapi.SlotStatus{
		Slotid:           int32(slotid),
		WorkingDirectory: e.GetWorkDir(),
		Host:             ourIP,
		Port:             int32(*shareport[slotid]),
		Profile:          settings.SpecFile[slotid],
		State:            e.GetState().String(),
		ActiveUsers:      ss.GetUsers(),
	}

	resources := make([]string, 0)
	if e.GetFileRecord().FileName != "" {
		resources = append(resources, e.GetFileRecord().FileName)
	}
	if settings.PureBootVolume[slotid] != "" {
		resources = append(resources, settings.PureBootVolume[slotid])
	}
	if settings.PureBootVolume2[slotid] != "" {
		resources = append(resources, settings.PureBootVolume2[slotid])
	}

	return s

}

func statusReporter() {

	time.Sleep(60 * time.Second)

	for {

		st := getSlotStatus(0)

		if st != nil {
			req := &remoteapi.RemoteStatusRequest{
				Status: st,
			}
			_, err := s8webclient.CONN.ReportStatus(req)
			if err != nil {
				log.Printf("Report status failed: %v", err)
			}
		}

		<-time.After(60 * time.Second)
	}
}

func main() {

	ourIP = GetOutboundIP().String()

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", 7070), nil)
	}()

	FlipCase = false

	flag.Parse()

	mos6502.TRACE = *trace6502
	mos6502.PROFILE = *prof
	mos6502.HEATMAP = *heatmap

	if *versionDisplay {
		os.Exit(0)
	}

	if *dataHost != "" {
		s8webclient.CONN = s8webclient.New(*dataHost, ":6581")
	}

	networkInit(*dataHost)
	memoryInit()

}

func SendPaddleButton(pdl int, v uint64) {
	for i := 0; i < 8; i++ {
		RAM.IntSetPaddleButton(i, pdl, v)
	}
}

func SendPaddleModify(pdl int, mv int) {
	for i := 0; i < 8; i++ {
		v := int(RAM.IntGetPaddleValue(i, pdl)) + mv
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}

		RAM.IntSetPaddleValue(i, pdl, uint64(v))
	}
}

func SendPaddleValue(pdl int, f float32) {
	for i := 0; i < 8; i++ {
		RAM.IntSetPaddleValue(i, pdl, uint64(127+127*f))
	}
}
