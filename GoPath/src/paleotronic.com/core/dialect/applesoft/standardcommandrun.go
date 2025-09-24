package applesoft

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

type StandardCommandRUN struct {
	dialect.Command
}

func (this *StandardCommandRUN) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	apple2helpers.SetSPEED(caller, 255)
	caller.SetMemory(49168, 0)

	result = 0

	if caller.IsRunningDirect() && (tokens.Size() > 0) {
		// collapse tokens
		out := ""

		for _, t := range tokens.Content {
			out = out + t.Content
		}

		tokens.Clear()
		tokens.Push(types.NewToken(types.STRING, out))
	}

	if caller.IsEmpty() {
		return result, exception.NewESyntaxError("Current entity has no code")
	}

	//_ = caller.GetMemory(49152 + 89) // reset them

	//caller.GetMemoryMap().PaddleMap[caller.GetMemIndex()] = make(map[int]int)
	caller.GetMemoryMap().IntSetPaddleValue(caller.GetMemIndex(), 0, 127)
	caller.GetMemoryMap().IntSetPaddleValue(caller.GetMemIndex(), 1, 127)
	caller.GetMemoryMap().IntSetPaddleValue(caller.GetMemIndex(), 2, 127)
	caller.GetMemoryMap().IntSetPaddleValue(caller.GetMemIndex(), 3, 127)

	caller.GetMemoryMap().IntSetZeroPageState(caller.GetMemIndex(), 0)
	caller.GetMemoryMap().IntSetPDState(caller.GetMemIndex(), 0)

	apple2helpers.SpriteReset(caller)

	caller.SetBreakable(true)

	apple2.PDL_TARGET = [4]int64{0, 0, 0, 0}

	//apple2helpers.SetTextFull(caller)

	startLine := 0
	//boolean keepvars = (caller.State == RUNNING);
	keepvars := false
	locals := caller.GetLocal()

	purgeVideo := true

	if tokens.Size() > 0 {
		tl := types.NewTokenList()
		if (tokens.Size() == 1) && (tokens.LPeek().Type == types.VARIABLE) && (!caller.ExistsVar(tokens.LPeek().Content)) {
			tokens.LPeek().Type = types.STRING
		}
		b := caller.ParseTokensForResult(tokens)
		tl.Push(&b)
		if tl.LPeek().Type == types.STRING {
			empty := types.NewTokenList()

			// if caller.IsRunningDirect() {
			// 	resetState(caller)
			// }

			_, e := caller.GetDialect().GetCommands()["load"].Execute(env, caller, *tl, Scope, LPC)
			if e != nil {
				index := caller.GetMemIndex()
				mm := caller.GetMemoryMap()
				cindex := mm.GetCameraConfigure(index)
				for i := 0; i < 8; i++ {
					control := types.NewOrbitController(mm, index, cindex)
					control.ResetALL()
					control.SetZoom(types.GFXMULT)
				}
				return 0, e
			}
			caller.GetDialect().GetCommands()["text"].Execute(env, caller, *empty, Scope, LPC)
		} else {
			startLine = tl.LPeek().AsInteger()
			purgeVideo = false
		}
	}

	apple2helpers.TrashCPU(caller)
	fmt.Println("About to run")
	caller.Run(false)
	fmt.Printf("curent run state: %v\n", caller.GetState())
	caller.GetDialect().BeforeRun(caller)

	if keepvars {
		caller.SetLocal(locals)
	}

	if startLine > 0 {
		caller.GetPC().Line = startLine
	}

	// Purge keys
	caller.GetMemoryMap().KeyBufferEmpty(caller.GetMemIndex())

	// flush HGR buffers
	if purgeVideo {
		//AppleHiRES.HgrFill( ((AppleVDU)caller.GetVDU()).BitmapMemory[0], 0 )
		//AppleHiRES.HgrFill( ((AppleVDU)caller.GetVDU()).BitmapMemory[1], 0 )

		//((AppleVDU)caller.GetVDU()).BitmapMemory[0].ClearTransactions()
		//((AppleVDU)caller.GetVDU()).BitmapMemory[1].ClearTransactions()

		caller.GetLoopStack().Clear()
		//((AppleVDU)(caller.GetVDU())).CameraReset()
	}
	/* enforce non void return */

	caller.SetFirstString("")

	memory.WarmStart = true
	caller.LoadSpec(caller.GetSpec())
	memory.WarmStart = false

	configureCPUMode(caller)

	if settings.IsRemInt {
		data := caller.GetMemoryMap().BlockRead(caller.GetMemIndex(), caller.GetMemoryMap().MEMBASE(caller.GetMemIndex())+8192, 16384)
		caller.GetMemoryMap().BlockWrite(caller.GetMemIndex(), caller.GetMemoryMap().MEMBASE(caller.GetMemIndex())+8192, data)
	}

	empty := types.NewTokenList()
	caller.GetDialect().GetCommands()["text"].Execute(env, caller, *empty, Scope, LPC)
	caller.SetMemory(0xc05f, 0)

	index := caller.GetMemIndex()
	mm := caller.GetMemoryMap()
	cindex := mm.GetCameraConfigure(index)
	//log2.Printf("*** cindex = %d", cindex)
	for i := 0; i < 8; i++ {
		control := types.NewOrbitController(mm, index, cindex)
		control.ResetALL()
		control.SetZoom(types.GFXMULT)
		control.Update()
	}

	apple2helpers.PixelTextX, apple2helpers.PixelTextY = 0, 0
	apple2helpers.PixelTextColor = 3
	apple2helpers.PixelTextWidth = 1
	apple2helpers.PixelTextHeight = 1
	apple2helpers.PixelTextFont = apple2helpers.LoadNormalFontGlyphs()

	//
	caller.GetMemoryMap().IntSetTargetSlot(caller.GetMemIndex(), 0)

	return result, nil

}

func (this *StandardCommandRUN) Syntax() string {

	/* vars */
	var result string

	result = "RUN"

	/* enforce non void return */
	return result

}

func configureCPUMode(caller interfaces.Interpretable) {

	code := caller.GetCode()

	h := code.GetHighIndex()
	l := code.C[h]

	var lastkw bool

	for _, st := range l {
		// st == Statement
		t := st.LPeek()
		fmt.Printf("token = %s\n", t.Content)
		lastkw = false
		if strings.ToLower(t.Content) == "call" && t.Type == types.KEYWORD {
			lastkw = true
		}
	}

	if lastkw {
		fmt.Println("configureCPUMode: BasicMode = false")
		cpu := apple2helpers.GetCPU(caller)
		cpu.BasicMode = false
	}

}
