package debugger

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fmt"

	"paleotronic.com/log"

	"paleotronic.com/octalyzer/assets"

	"paleotronic.com/core/hardware/cpu/mos6502"
	"paleotronic.com/core/hardware/cpu/mos6502/asm"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/settings"
	"paleotronic.com/debugger/debugtypes"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func serveFunc(w http.ResponseWriter, r *http.Request) {

	//log.Println("in serveFunc")

	vars := mux.Vars(r)

	file := "/" + vars["file"]
	if file == "/" {
		file = "/index.html"
	}

	ext := filepath.Ext(file)
	if ext == "" {
		file = "/" + strings.Trim(file, "/") + "/index.html"
	}

	file = "debugger" + file

	log.Printf("Request for file: %s", file)

	var data []byte
	var err error

	if os.Getenv("LOCAL_DEBUG_FILES") != "" {
		data, err = ioutil.ReadFile(file)
	} else {
		data, err = assets.Asset(file)
		if err != nil {
			log.Println("Not found")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("File not found"))
			return
		}
	}

	log.Printf("ext = %s", ext)

	switch ext {
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "text/html")
	}

	w.Write(data)

}

func (d *Debugger) Serve(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/api/debug/screen/", apiScreen)
	r.HandleFunc("/api/debug/asm/upload", apiASMUpload)
	r.HandleFunc("/api/debug/upload", apiUpload)
	r.HandleFunc("/api/debug/download/{mode}/{addr}/{size}", apiDownload)
	r.HandleFunc("/api/websocket/debug", apiWSDebug)
	r.HandleFunc("/{file:.*}", serveFunc)
	r.HandleFunc("/", serveFunc)
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
}

type errStruct struct {
	Error string
}

func response(data interface{}, w http.ResponseWriter, r *http.Request, status int) {
	j, _ := json.Marshal(data)
	w.Write(j)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

func errResponse(err error, w http.ResponseWriter, r *http.Request, status int) {
	e := &errStruct{Error: err.Error()}
	j, _ := json.Marshal(e)
	w.Write(j)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("Not found"))
}

func formatDump(addr int, data []byte) []byte {
	out := ""
	for i, v := range data {
		if i%8 == 0 {
			if out != "" {
				out += "\r\n"
			}
			out += fmt.Sprintf("%.4x: ", addr+i)
		}
		out += fmt.Sprintf("%.2x ", v)
	}
	out += "\r\n"
	return []byte(out)
}

func formatDasm(addr int, length int, cpu *mos6502.Core6502) []byte {
	out := ""
	i := addr
	for i < addr+length {
		bytdata, desc, _ := cpu.DecodeInstruction(i)
		bstr := ""
		for _, k := range bytdata {
			if bstr != "" {
				bstr += " "
			}
			bstr += fmt.Sprintf("%.2x", k)
		}
		out += fmt.Sprintf("%.4x: %-9s  %-20s\r\n", i, bstr, desc)
		i += len(bytdata)
	}
	out += "\r\n"
	return []byte(out)
}

func apiDownload(w http.ResponseWriter, r *http.Request) {

	//log.Println("in apiDownload")

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Bad method")
		return
	}
	vars := mux.Vars(r)
	mode := vars["mode"]
	if mode == "" {
		mode = "bin"
	}
	addr, ok := parseNumber(vars["addr"])
	if !ok {
		w.Write([]byte("Save address invalid " + vars["addr"]))
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad save address %s", vars["addr"])
		return
	}
	size, ok := parseNumber(vars["size"])
	if !ok {
		w.Write([]byte("Save length invalid " + vars["size"]))
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad save length %s", vars["size"])
		return
	}

	data := DebuggerInstance.ReadBlob(addr, size)

	switch mode {
	case "bin":
		w.Header().Set("Content-type", "application/octet-stream")
		w.Header().Set("Content-disposition", fmt.Sprintf("inline; filename=\"debug-save#0x%.4x.bin\"", addr))
		w.Write(data)
	case "txt":
		w.Header().Set("Content-type", "application/octet-stream")
		w.Header().Set("Content-disposition", fmt.Sprintf("inline; filename=\"debug-dump#0x%.4x.txt\"", addr))
		w.Write(formatDump(addr, data))
	case "dasm":
		w.Header().Set("Content-type", "application/octet-stream")
		w.Header().Set("Content-disposition", fmt.Sprintf("inline; filename=\"debug-dasm#0x%.4x.asm\"", addr))
		w.Write(formatDasm(addr, size, apple2helpers.GetCPU(DebuggerInstance.ent())))
	}

}

func apiUpload(w http.ResponseWriter, r *http.Request) {

	//log.Println("in apiUpload")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Bad method")
		return
	}
	loadAddrStr := r.Header.Get("X-LoadAddress")
	if loadAddrStr == "" {
		w.Write([]byte("Load address missing"))
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Load address missing")
		return
	}
	addr, ok := parseNumber(loadAddrStr)
	if !ok {
		w.Write([]byte("Load address invalid " + loadAddrStr))
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad load address %s", loadAddrStr)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Failed reading body"))
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed reading body: %v", err)
		return
	}
	log.Printf("Binary size: %d bytes", len(data))
	if len(data) > 65536 {
		w.Write([]byte("File data too large"))
		w.WriteHeader(http.StatusInsufficientStorage)
		return
	}
	DebuggerInstance.WriteBlob(addr, data)
}

func apiASMUpload(w http.ResponseWriter, r *http.Request) {

	//log.Println("in apiASMUpload")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Bad method")
		return
	}
	loadAddrStr := r.Header.Get("X-LoadAddress")
	if loadAddrStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Load address missing"))
		log.Printf("Load address missing")
		return
	}
	addr, ok := parseNumber(loadAddrStr)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Load address invalid " + loadAddrStr))
		log.Printf("Bad load address %s", loadAddrStr)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed reading body"))
		log.Printf("Failed reading body: %v", err)
		return
	}
	// at this point we have the bytes of the source file so need
	// to assemble them
	lines := strings.Split(string(data), "\n")

	a := asm.NewAsm65C02()
	blocks, _, msg, err := a.AssembleMultipass(lines, addr)
	log.Printf("asm result is %v", err)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("asm error: " + err.Error() + "\n" + msg))
		return
	}
	for newaddr, newdata := range blocks {
		log.Printf("Binary size: %d bytes", len(data))
		if len(newdata) > 65536 {
			w.WriteHeader(http.StatusInsufficientStorage)
			w.Write([]byte("File data too large"))
			return
		}
		DebuggerInstance.WriteBlob(newaddr, newdata)
	}
}

func apiWSDebug(w http.ResponseWriter, r *http.Request) {

	//log.Println("in apiWSDebug")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	DebuggerInstance.socket = c
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			DebuggerInstance.ContinueCPU()
			DebuggerInstance.Detach()
			break
		}

		log.Printf("recv: %s", message)

		// assume json
		var msg debugtypes.WebSocketCommand
		err = json.Unmarshal(message, &msg)
		if err == nil {
			resp := handleCommand(msg)
			// j, err := json.Marshal(&resp)
			if err == nil {
				//err = c.WriteMessage(mt, j)
				err = DebuggerInstance.SendMessage(resp.Type, resp.Payload, resp.Ok)
				if err != nil {
					log.Println("write:", err)
					break
				} else {
					log.Println("Sent response...")
				}
			} else {
				log.Printf("Failed to marshal json response: %v", err)
			}
		} else {
			log.Printf("Failed to unmarshal json payload: %v", err)
		}
	}
}

func apiScreen(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	if len(settings.ScreenShotJPEGData) > 0 {
		w.Write(settings.ScreenShotJPEGData)
	} else {
		data, err := assets.Asset("images/octasplash.png")
		if err != nil {
			panic("failed to load splash")
		}
		w.Write(data)
	}
}

func apiAttach(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slotid, err := strconv.ParseInt(vars["slotid"], 10, 64)
	if err != nil {
		errResponse(err, w, r, http.StatusBadRequest)
		return
	}
	if slotid < 0 || slotid > settings.NUMSLOTS-1 {
		errResponse(fmt.Errorf("Invalid slot %d", slotid), w, r, http.StatusBadRequest)
		return
	}
	go DebuggerInstance.Run(int(slotid), nil)
	w.WriteHeader(http.StatusOK)
}

func apiCPUState(w http.ResponseWriter, r *http.Request) {
	if DebuggerInstance.CPU == nil {
		errResponse(
			fmt.Errorf("No cpu state"),
			w,
			r,
			http.StatusNotFound,
		)
		return
	}

	response(
		DebuggerInstance.CPU,
		w,
		r,
		http.StatusOK,
	)
}

func apiCPUStep(w http.ResponseWriter, r *http.Request) {
	if !DebuggerInstance.attached {
		errResponse(
			fmt.Errorf("No cpu state"),
			w,
			r,
			http.StatusNotFound,
		)
		return
	}

	DebuggerInstance.StepCPU()

	response(
		DebuggerInstance.CPU,
		w,
		r,
		http.StatusOK,
	)
}

func apiCPUPause(w http.ResponseWriter, r *http.Request) {
	if !DebuggerInstance.attached {
		errResponse(
			fmt.Errorf("No cpu state"),
			w,
			r,
			http.StatusNotFound,
		)
		return
	}

	DebuggerInstance.PauseCPU()

	response(
		DebuggerInstance.CPU,
		w,
		r,
		http.StatusOK,
	)
}

func apiCPUContinue(w http.ResponseWriter, r *http.Request) {
	if !DebuggerInstance.attached {
		errResponse(
			fmt.Errorf("No cpu state"),
			w,
			r,
			http.StatusNotFound,
		)
		return
	}

	DebuggerInstance.ContinueCPU()

	response(
		DebuggerInstance.CPU,
		w,
		r,
		http.StatusOK,
	)
}

func apiCPUStatus(w http.ResponseWriter, r *http.Request) {
	if !DebuggerInstance.attached {
		errResponse(
			fmt.Errorf("No cpu status"),
			w,
			r,
			http.StatusNotFound,
		)
		return
	}

	m := &debugtypes.CPUMode{}
	cpu := apple2helpers.GetCPU(DebuggerInstance.ent())
	switch cpu.RunState {
	case mos6502.CrsFreeRun:
		m.State = "freerun"
	case mos6502.CrsPaused:
		m.State = "paused"
	case mos6502.CrsHalted:
		m.State = "halted"
	case mos6502.CrsSingleStep:
		m.State = "step"
	}

	response(
		m,
		w,
		r,
		http.StatusOK,
	)

}

func apiDecode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address, err := strconv.ParseInt(vars["address"], 10, 64)
	if err != nil {
		errResponse(err, w, r, http.StatusBadRequest)
		return
	}
	count, err := strconv.ParseInt(vars["count"], 10, 64)
	if err != nil {
		errResponse(err, w, r, http.StatusBadRequest)
		return
	}
	if address < 0 || address > 65536 {
		errResponse(fmt.Errorf("Invalid address %d", address), w, r, http.StatusBadRequest)
		return
	}
	cpu := apple2helpers.GetCPU(DebuggerInstance.ent())

	instr := make([]debugtypes.CPUInstructionDecode, int(count))
	for i, _ := range instr {
		code, desc, cycles := cpu.DecodeInstruction(int(address))
		instr[i].Address = int(address)
		instr[i].Bytes = code
		instr[i].Instruction = desc
		instr[i].Cycles = cycles
		address += int64(len(code)) % 65536
	}
	response(
		&debugtypes.CPUInstructions{Instructions: instr},
		w,
		r,
		http.StatusOK,
	)
}
