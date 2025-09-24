// +build remint

package main

/*
 * REMINT SERVICE LAYER
 * ====================
 */

import "paleotronic.com/ducktape/server"
import "paleotronic.com/ducktape"
import "paleotronic.com/core/memory"
import "paleotronic.com/encoding/mempak"
import "paleotronic.com/utils"
import "paleotronic.com/octalyzer/backend"
import "paleotronic.com/fmt"
import "paleotronic.com/log"
import "time"

var s *server.DuckTapeServer
var VALID_CONTROL map[string]map[int]int

func DefineMatrix() {
	VALID_CONTROL = make(map[string]map[int]int)

	VALID_CONTROL["april"] = make(map[int]int)
	VALID_CONTROL["april"][memory.PDL0] = memory.PDL0
	VALID_CONTROL["april"][memory.PDL1] = -1
	VALID_CONTROL["melody"] = make(map[int]int)
	VALID_CONTROL["melody"][memory.PDL0] = memory.PDL1
	VALID_CONTROL["melody"][memory.PDL1] = -1
}

func init() {
	//DefinePresets()
	DefineMatrix()
}

func RunServer() {

	m := make(server.DuckHandlerMap)
	//m["TOS"] = TransferHandler
	//m["TRM"] = TerminateHandler
	//m["SCM"] = ControlsHandler
	//m["CAM"] = CameraHandler
	//m["QIM"] = QueryInputHandler
	m["CBY"] = MemHandler
	m["FUL"] = FullHandler

	s = server.NewDuckTapeServer(":"+utils.IntToStr(*useport), m)
	s.OnDisconnect = OnChange
	go s.Run()

	fmt.Println("=================================================================")
	fmt.Println(" REMOTE SERVICE RUNNING ON", utils.IntToStr(*useport))
	fmt.Println("=================================================================")

	time.Sleep(50 * time.Millisecond)

	//~ for i := 0; i<len(c); i++ {
	//~ fmt.Println("Spinning up ORAC"+utils.IntToStr(i)+"...")
	//~ c[i] = client.NewDuckTapeClient( "localhost", ":"+utils.IntToStr(*useport), "ORAC"+utils.IntToStr(i), "tcp" )
	//~ e := c[i].Connect()
	//~ if e != nil {
	//~ panic(e)
	//~ }
	//~ c[i].SubscribeChannel( "cliupd"+utils.IntToStr(i) )
	//~ c[i].SendChannel = "memupd"+utils.IntToStr(i)
	//~ time.Sleep(1*time.Millisecond)
	//~ }

}

func OnChange(s *server.DuckTapeServer, count int, ocount int) {

	p := backend.ProducerMain

	if p == nil {
		return
	}

	fmt.Printf("%d clients left...\n", count)

	counts := s.ChannelCounts()

	for k, v := range counts {
		fmt.Printf("%20s %d\n", k, v)
	}

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {

		if counts["memupd"+utils.IntToStr(i)] == 0 && !*benchmode {
			if p.GetInterpreter(i) != nil {
				p.GetInterpreter(i).StopTheWorld()
				fmt.Println("Suspending", i)
				RAM.Track[i] = false
				_ = RAM.GetRemoteLoggedChanges(i)
				RAM.MemCapMode[i] = memory.MEMCAP_REMOTE // <--- remote capture mode
				p.GetInterpreter(i).SetClientSync(false)
			}
		} else {
			if p.GetInterpreter(i) != nil {
				p.GetInterpreter(i).ResumeTheWorld()
				fmt.Println("Resuming", i)
				RAM.Track[i] = true
				RAM.MemCapMode[i] = memory.MEMCAP_REMOTE // <--- remote capture mode
				p.GetInterpreter(i).SetClientSync(false)
			}
		}

	}
}

func PostCycleCallback(slotid int) {

	//fmt.Print(slotid)

	data := RAM.GetRemoteLoggedChanges(slotid)
	count := 0
	if len(data) > 0 {

		payload := make([]byte, 3)

		for _, mc := range data {

			if len(mc.Value) == 0 {
				payload = append(payload, mempak.Encode(0, mc.Global, 0, true)...)
				count++
			} else {

				if mc.Global >= memory.OCTALYZER_SIM_SIZE || RAM.IsMappedAddress(slotid, mc.Global) {

					for i, v := range mc.Value {
						payload = append(payload, mempak.Encode(0, mc.Global+i, v, false)...)
						count++
					}

				}
			}

		}

		if count > 0 {
			payload[0] = byte((count / 65536) % 256)
			payload[1] = byte((count / 256) % 256)
			payload[2] = byte(count % 256)
			s.DirectSend("BMU", payload, true, "memupd"+utils.IntToStr(slotid))

			//fmt.Printf("*** %d memchanges shipped for slot %d\n", count, slotid)
			//ProcessOutgoing(slotid)
			//ProcessIncoming(slotid)
			return
		}
	}

	//ProcessOutgoing(slotid)
	//ProcessIncoming(slotid)

	return
}

// MemHandler handles single byte updates from the client
func MemHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	slotid, addr, value, read, _, _ := mempak.Decode(msg.Payload)

	user := c.Name

	if read {
		return nil
	}

	index := int(slotid)

	redirector, ok := VALID_CONTROL[user]
	if ok {
		raddr, exists := redirector[addr]
		if exists {
			addr = raddr
		}
	}

	if addr != -1 {
		RAM.WriteInterpreterMemorySilent(index, addr, value)
	}

	return nil
}

// Full Handler handles full memory update request from the client
func FullHandler(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {

	slotid := int(msg.Payload[0] - 48)

	fmt.Printf("Memory sync requested by %s for slot %d\n", c.Name, slotid)

	p := backend.ProducerMain

	if p == nil {
		return nil
	}

	data, _ := p.GetInterpreter(slotid).FreezeBytes()
	cdata := utils.GZIPBytes(data)
	c.SendMessageEx("RAM", cdata, true)
	log.Printf("Sending RAM snapshot of %d bytes\n", len(cdata))

	return nil
}
