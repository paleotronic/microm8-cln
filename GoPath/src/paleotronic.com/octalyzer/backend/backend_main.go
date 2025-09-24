package backend

import (
	"runtime"
	"time"

	"paleotronic.com/api"
	"paleotronic.com/core"
	"paleotronic.com/core/memory" //"os"
	"paleotronic.com/core/settings"
	"paleotronic.com/filerecord"
	"paleotronic.com/fmt"
	"paleotronic.com/panic"
	"paleotronic.com/utils"
	//	"paleotronic.com/octalyzer/video"
)

// ProducerMain is the public instance of the producer
var ProducerMain *core.Producer
var RUNNING bool
var STOPPED bool
var REBOOT_NEEDED bool

// HandleException is called when a panic is found executing the interpreters
// capture stacktrace and state and create a Bug Ticket
func HandleException(r interface{}) {

	if settings.SystemType != "nox" {

	//panic(r)

	b := make([]byte, 8192)
	i := runtime.Stack(b, false)
	// Stack trace
	stackstr := string(b[0:i])
	slotid := ProducerMain.GetContext()

	fmt.Println(stackstr)

	ent := ProducerMain.GetInterpreter(slotid)
	for ent.GetChild() != nil {
		ent = ent.GetChild()
	}

	// Construct record
	bug := filerecord.BugReport{}
	bug.Summary = "System crashed"
	bug.Body = `
System crash  occurred in Slot#` + utils.IntToStr(slotid) + `
` + fmt.Sprintf("%v", r) + `

Stack trace:
` + stackstr + `
    `
	// Add compressed stuff
	att := filerecord.BugAttachment{}
	att.Name = "Compressed Runstate"
	att.Created = time.Now()
	tmp, _ := ent.FreezeBytes()
	att.Content = utils.GZIPBytes(tmp)
	bug.Attachments = []filerecord.BugAttachment{att}
	bug.Comments = []filerecord.BugComment{
		filerecord.BugComment{
			User:    "system",
			Content: "Logged automatically by runtime system.",
			Created: time.Now(),
		},
	}
	bug.Creator = s8webclient.CONN.Username
	bug.Filename = ent.GetFileRecord().FileName
	bug.Filepath = ent.GetFileRecord().FilePath
	bug.Created = time.Now()

	if bug.Filename != "" {
		bug.Summary = "Program crash: " + bug.Filepath + "/" + bug.Filename
	}

	//fmt.Println(bug.Body)

	_ = s8webclient.CONN.CreateUpdateBug(bug)

}

	REBOOT_NEEDED = true
}

func Run(r *memory.MemoryMap, callback func(slotid int)) {

	RUNNING = true
	STOPPED = false

	for RUNNING {

		panic.Do(
			func() {
				RunCore(r, callback)
			},
			HandleException,
		)

	}

	STOPPED = true

}

func Stop() {

	RUNNING = false

	for !STOPPED {
		time.Sleep(1 * time.Millisecond)
	}

}

func RunCore(r *memory.MemoryMap, callback func(slotid int)) {

	runtime.LockOSThread()

	//fmt.Println("Creating producer")
	ProducerMain = core.NewProducer(r, bootstrap)
	//ProducerMain = core.NewProducer(r, "fp")
	core.SetInstance(ProducerMain)
	//fmt.Println("Created producer")

	ProducerMain.NeedsPrompt = true

	//fmt.Println("Executing producer")

	for i := 0; i < memory.OCTALYZER_NUM_INTERPRETERS; i++ {
		ProducerMain.PostCallback[i] = callback
	}
	ProducerMain.Run()
	//fmt.Println("done producer")

}

type PP struct {
}

func (p *PP) StopTheWorld(slotid int) {
	for ProducerMain == nil {
		time.Sleep(50 * time.Millisecond)
	}
	ProducerMain.StopTheWorld(slotid)
}

func (p *PP) ResumeTheWorld(slotid int) {
	for ProducerMain == nil {
		time.Sleep(50 * time.Millisecond)
	}
	ProducerMain.ResumeTheWorld(slotid)
}

func (p *PP) InjectCommands(slotid int, command string) {
	for ProducerMain == nil {
		time.Sleep(50 * time.Millisecond)
	}
	ProducerMain.InjectCommands(slotid, command)
}

func (p *PP) Halt(slotid int) {
	for ProducerMain == nil {
		time.Sleep(50 * time.Millisecond)
	}
	ProducerMain.Halt(slotid)
}

func (p *PP) Freeze(slotid int) []byte {
	return ProducerMain.Freeze(slotid)
}

var VPP *PP = &PP{}
