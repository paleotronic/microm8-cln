package memory

/*
 * SHARE SERVICE
 *
 */

import "time"
import "strings"
import "sort"
import "sync"
import "paleotronic.com/fmt"
import "paleotronic.com/utils"
import "paleotronic.com/fastserv"
import "paleotronic.com/encoding/mempak"
import "paleotronic.com/log"

const MAX_SHARES = 12288

type ShareService struct {
	s           *fastserv.Server
	port        int
	slotid      int
	q           []*MemoryChange
	mm          *MemoryMap
	start       time.Time
	counter     int64
	sc          ShareControl
	connections int
	InChannel   chan *MemoryChange

	// throughPut measurement
	lastSecond int64 // last time network was measured
	byteCount  int64

	memchanges sync.Mutex

	// Mapper
	mapper map[string]string
}

type ShareControl interface {
	StopTheWorld(slotid int)
	ResumeTheWorld(slotid int)
	InjectCommands(slotid int, command string)
	Halt(slotid int)
	Freeze(slotid int) []byte
}

func (ss *ShareService) OnConnect(name string) {
	ss.connections++

	ss.mapper[name] = "JOY0"

	fmt.Printf("Connection on slot #%d - now %d users\n", ss.slotid, ss.connections)

	if ss.connections > 0 {
		for ss.sc == nil {
			time.Sleep(50 * time.Millisecond)
		}
		ss.sc.ResumeTheWorld(ss.slotid)
	}
}

func (ss *ShareService) OnDisconnect() {
	ss.connections--

	fmt.Printf("Disconnection from slot #%d - now %d users\n", ss.slotid, ss.connections)

	if ss.connections == 0 {
		ss.sc.StopTheWorld(ss.slotid)
	}
}

// Mask decides if an update needs to pass through...
func (ss *ShareService) Mask(addr int) bool {

	addr = addr % OCTALYZER_INTERPRETER_SIZE

	return true

}

func (ss *ShareService) Start(sc ShareControl, mm *MemoryMap, slotid int, port int) {

	m := make(fastserv.HandlerMap)

	m[fastserv.FS_CLIENTMEM] = ss.MemHandler
	m[fastserv.FS_MEMSYNC_REQUEST] = ss.FullHandler
	m[fastserv.FS_REQUEST_TRANSFER_OWNERSHIP] = ss.OwnershipHandler
	m[fastserv.FS_ALLOCATE_CONTROL] = ss.ControlHandler
	m[fastserv.FS_REMOTE_PARSE] = ss.CommandHandler
	m[fastserv.FS_REMOTE_EXEC] = ss.ExecHandler

	ss.port = port
	ss.slotid = slotid
	ss.mm = mm
	ss.start = time.Now()
	ss.sc = sc

	if ss.s == nil {
		ss.s = fastserv.NewServer(":"+utils.IntToStr(port), m)
		//ss.s.OnDisconnect = ss.OnChange
	}

	ss.s.OnConnect = ss.OnConnect
	ss.s.OnDisconnect = ss.OnDisconnect

	if ss.s.Running {
		ss.s.Stop()
	}

	ss.InChannel = make(chan *MemoryChange, MAX_SHARES)

	go ss.s.Run()
	time.Sleep(time.Millisecond)
	go ss.BusyBee()

	// Add hook
	ss.mm.SetCustomLogger(ss.slotid, ss.Post)
	ss.mm.CustomAudioLogger[ss.slotid] = ss.Audio
	ss.mm.SetRestCallback(ss.slotid, ss.RestCommand)
}

func (ss *ShareService) Stop() {

	if ss.s == nil {
		return
	}

	ss.mm.SetCustomLogger(ss.slotid, nil)
	ss.mm.CustomAudioLogger[ss.slotid] = nil

	ss.s.Stop()
}

// Post will post a change to the network...  called by MemoryMap.DoCustomLog()
func (ss *ShareService) Post(mci *MemoryChange) {

	//~ for len(ss.InChannel) > 12000 {

	//~ time.Sleep(250*time.Microsecond)

	//~ }

	// if mci.Global%OCTALYZER_INTERPRETER_SIZE >= 0x3e000 && mci.Global%OCTALYZER_INTERPRETER_SIZE < 0x3f000 {
	// 	fmt.Printf("Memory event sending %d -> %d\n", mci.Value[0], mci.Global%OCTALYZER_INTERPRETER_SIZE)
	// }

	ss.InChannel <- mci

}

func (ss *ShareService) RestCommand(index int, command string) {
	msg := []byte{byte(fastserv.FS_RESTALGIA_COMMAND)}
	msg = append(msg, []byte(command)...)
	ss.s.InChannel <- msg
}

func (ss *ShareService) Audio(c, rate int, bytepacked bool, indata []uint64) {

	//fmt.Println("ss.Audio")

	data := []uint64{uint64(rate)}
	data = append(data, indata...)
	b := append([]byte{byte(fastserv.FS_CLIENTAUDIO)}, mempak.PackSliceUints(data)...)
	ss.s.InChannel <- b
	//ss.byteCount += len(b)

}

func (ss *ShareService) shouldFilterAddress(addr int) bool {

	if addr >= MICROM8_2ND_DISKII_BASE && addr <= MICROM8_2ND_DISKII_BASE+2*MICROM8_2ND_DISKII_SIZE {
		return true
	}

	if addr >= OCTALYZER_DISKII_BASE && addr <= OCTALYZER_DISKII_BASE+2*OCTALYZER_DISKII_SIZE {
		return true
	}

	if addr >= OCTALYZER_MAPPED_CAM_BASE && addr <= OCTALYZER_MAPPED_CAM_END {
		return true
	}

	if addr >= OCTALYZER_DISKII_CARD_STATE && addr < OCTALYZER_VOXEL_DEPTH {
		return true
	}

	if addr >= MICROM8_R6522_BASE && addr < MICROM8_RESTALGIA_PATH_LOOP {
		return true
	}

	return false

}

func (ss *ShareService) BusyBee() {

	for !ss.s.Running {
		time.Sleep(1 * time.Millisecond)
	}

	count := 0
	payload := make([]byte, 4)

	ticker := time.NewTicker(5 * time.Millisecond)

	for ss.s.Running {

		select {
		case mc := <-ss.InChannel:
			if len(mc.Value) > 0 {

				a := mc.Global % OCTALYZER_INTERPRETER_SIZE

				if (a >= OCTALYZER_SIM_SIZE || ss.mm.IsMappedAddress(ss.slotid, a)) && !ss.shouldFilterAddress(a) {

					if len(mc.Value) > 3 {
						data := mempak.EncodeBlock(0, a, mc.Value)
						l := len(data)
						payload = append(payload, 0x18) // special flag byte
						payload = append(payload, byte((l/65536)%256))
						payload = append(payload, byte((l/256)%256))
						payload = append(payload, byte(l%256))
						payload = append(payload, data...)
						count++
					} else {
						for i, v := range mc.Value {
							payload = append(payload, mempak.Encode(0, a+i, v, false)...)
							count++
						}
					}

				}
			}
		case _ = <-ticker.C:
			/* build message */
			/* nothing to do */
			if count > 0 {
				payload[0] = byte(fastserv.FS_BULKMEM)
				payload[1] = byte((count / 65536) % 256)
				payload[2] = byte((count / 256) % 256)
				payload[3] = byte(count % 256)

				ss.byteCount += int64(len(payload))

				fmt.Printf("Sending payload of %d bytes\n", len(payload))
				ss.s.InChannel <- payload

				// re-init
				count = 0
				payload = make([]byte, 4)
			}
		}

	}

}

func (ss *ShareService) BusyBeeNew() {

	for !ss.s.Running {
		time.Sleep(1 * time.Millisecond)
	}

	count := 0
	payload := make([]byte, 4)

	ticker := time.NewTicker(2 * time.Millisecond)

	zoot := make(map[int]uint64)

	for ss.s.Running {

		select {
		case mc := <-ss.InChannel:
			if len(mc.Value) > 0 {

				a := mc.Global % OCTALYZER_INTERPRETER_SIZE

				if a >= OCTALYZER_SIM_SIZE || ss.mm.IsMappedAddress(ss.slotid, a) {

					for i, v := range mc.Value {
						zoot[a+i] = v
					}

				}
			}
		case _ = <-ticker.C:
			/* build message */
			if len(zoot) == 0 {
				continue
			}

			count = 0

			zootkeys := make([]int, len(zoot))
			idx := 0
			for addr, _ := range zoot {
				zootkeys[idx] = addr
			}
			sort.Ints(zootkeys)
			for addr, _ := range zoot {
				payload = append(payload, mempak.Encode(0, addr, zoot[addr], false)...)
				count++
			}
			/* nothing to do */
			if count > 0 {
				payload[0] = byte(fastserv.FS_BULKMEM)
				payload[1] = byte((count / 65536) % 256)
				payload[2] = byte((count / 256) % 256)
				payload[3] = byte(count % 256)

				ss.byteCount += int64(len(payload))

				ss.s.InChannel <- payload

				// re-init
				count = 0
				payload = make([]byte, 4)

				zoot = make(map[int]uint64)
			}

		}

	}

}

func (ss *ShareService) MemHandler(c *fastserv.Client, s *fastserv.Server, msg []byte) error {

	//~ ss.memchanges.Lock()
	//~ defer ss.memchanges.Unlock()

	_, addr, value, read, _, _ := mempak.Decode(msg[1:])

	log.Printf("MEM slot %d event from client: addr=%d, value=%d, read=%v", ss.slotid, addr, value, read)

	if read {
		return nil
	}

	index := ss.slotid

	p, _ := ss.mapper[c.Name]
	if p == "" {
		p = "NONE"
	}
	redir, exists := PROFILES[p][addr]
	if exists {
		addr = redir
	}

	if addr != -1 && addr >= 500000 {
		ss.mm.WriteInterpreterMemorySilent(index, addr%OCTALYZER_INTERPRETER_SIZE, value)
		//ss.mm.AddIncoming( index, addr % OCTALYZER_INTERPRETER_SIZE, value )
	}

	return nil
}

func (ss *ShareService) FullHandler(c *fastserv.Client, s *fastserv.Server, msg []byte) error {

	slotid := ss.slotid

	cdata := utils.GZIPBytes(ss.sc.Freeze(slotid))

	msgm := []byte{byte(fastserv.FS_MEMSYNC_RESPONSE)}
	msgm = append(msgm, cdata...)
	c.SendMessage(msgm)
	log.Printf("Sending RAM snapshot of %d bytes from slot %d\n", len(cdata), ss.slotid)

	return nil
}

func NewShareService() *ShareService {
	m := make(map[string]string)
	m["april"] = "PDL0"
	m["melody"] = "PDL1"
	return &ShareService{
		mapper: m,
		q:      make([]*MemoryChange, 0),
	}
}

func (ss *ShareService) GetUsers() []string {
	if ss.s == nil || !ss.s.Running {
		return []string{}
	}
	return ss.s.GetUsers()
}

// Ownership handler for server
func (ss *ShareService) OwnershipHandler(c *fastserv.Client, s *fastserv.Server, msg []byte) error {

	// Msg format: []byte{targetname}
	// Rsp format: []byte{1} = ok, []byte{0} = error

	name := string(msg[1:])
	if s.IsOwner(c) {
		ok := s.TransferOwnership(c, name)
		if ok {
			c.SendMessage([]byte{byte(fastserv.FS_TRANSFER_OWNERSHIP_OK)})
			return nil
		}
	}

	c.SendMessage([]byte{0})

	return nil
}

// Control map handler
func (ss *ShareService) ControlHandler(c *fastserv.Client, s *fastserv.Server, msg []byte) error {

	// Msg format: []byte{ name : profile }
	// Rsp format: []byte{1} = ok, []byte{0} = error

	if !s.IsOwner(c) {
		c.SendMessage([]byte{0})
		return nil
	}

	tmp := string(msg[1:])
	parts := strings.SplitN(tmp, ":", 2)

	if len(parts) == 2 {
		name := strings.Trim(strings.ToLower(parts[0]), " ")
		profile := strings.Trim(strings.ToUpper(parts[1]), " ")
		_, valid := PROFILES[profile]
		if valid {
			// try find connection with name
			cc := s.Find(name)
			if cc != nil {
				ss.mapper[name] = profile
				c.SendMessage([]byte{byte(fastserv.FS_ALLOCATE_CONTROL_OK)})
			}
		}
	}

	c.SendMessage([]byte{0})

	return nil
}

func (ss *ShareService) CommandHandler(c *fastserv.Client, s *fastserv.Server, msg []byte) error {

	if !s.IsOwner(c) {
		c.SendMessage([]byte{0})
		return nil
	}

	command := string(msg[1:])

	// Now we need to feed the command back into the system
	ss.sc.InjectCommands(ss.slotid, command)

	return nil
}

func (ss *ShareService) ExecHandler(c *fastserv.Client, s *fastserv.Server, msg []byte) error {

	if !s.IsOwner(c) {
		c.SendMessage([]byte{0})
		return nil
	}

	command := string(msg[1:])

	// Now we need to feed the command back into the system
	ss.sc.Halt(ss.slotid)
	ss.sc.InjectCommands(ss.slotid, command)

	return nil
}

var PADDLE0 map[int]int = map[int]int{
	PDL0: PDL0,
	PDL1: 0,
	PDL2: 0,
	PDL3: 0,
}

var PADDLE1 map[int]int = map[int]int{
	PDL0: PDL1,
	PDL1: 0,
	PDL2: 0,
	PDL3: 0,
}

var PADDLE2 map[int]int = map[int]int{
	PDL0: PDL2,
	PDL1: 0,
	PDL2: 0,
	PDL3: 0,
}

var PADDLE3 map[int]int = map[int]int{
	PDL0: PDL3,
	PDL1: 0,
	PDL2: 0,
	PDL3: 0,
}

var JOYSTICK0 map[int]int = map[int]int{
	PDL0: PDL0,
	PDL1: PDL1,
	PDL2: 0,
	PDL3: 0,
}

var JOYSTICK1 map[int]int = map[int]int{
	PDL0: PDL2,
	PDL1: PDL3,
	PDL2: 0,
	PDL3: 0,
}

var PADDLENONE map[int]int = map[int]int{
	PDL0: 0,
	PDL1: 0,
	PDL2: 0,
	PDL3: 0,
}

var PROFILES map[string]map[int]int = map[string]map[int]int{
	"PDL0": PADDLE0,
	"PDL1": PADDLE1,
	"PDL2": PADDLE2,
	"PDL3": PADDLE3,
	"JOY0": JOYSTICK0,
	"JOY1": JOYSTICK1,
	"NONE": PADDLENONE,
}
