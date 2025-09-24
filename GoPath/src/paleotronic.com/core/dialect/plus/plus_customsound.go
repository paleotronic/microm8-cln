package plus

import (
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/utils"
)

type PlusCustomSound struct {
	dialect.CoreFunction
}

func SendSongStop(ent interfaces.Interpretable, r *memory.MemoryMap) {
	// wait for channel
	for r.ReadGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND) != 0 {
		time.Sleep(50 * time.Microsecond)
	}
	// Send instrument data
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER_COUNT, 0)
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND, uint64(types.RS_StopSong))
}

func SendSongPause(ent interfaces.Interpretable, r *memory.MemoryMap) {
	// wait for channel
	for r.ReadGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND) != 0 {
		time.Sleep(50 * time.Microsecond)
	}
	// Send instrument data
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER_COUNT, 0)
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND, uint64(types.RS_PauseSong))
}

func SendSongResume(ent interfaces.Interpretable, r *memory.MemoryMap) {
	// wait for channel
	for r.ReadGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND) != 0 {
		time.Sleep(50 * time.Microsecond)
	}
	// Send instrument data
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER_COUNT, 0)
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND, uint64(types.RS_ResumeSong))
}

func SendInstrument(ent interfaces.Interpretable, r *memory.MemoryMap, instrument string) {
	// wait for channel
	for r.ReadGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND) != 0 {
		time.Sleep(50 * time.Microsecond)
	}
	// Send instrument data
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER_COUNT, uint64(len(instrument)))
	for i, v := range instrument {
		r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER+i, uint64(v))
	}
	// Finalise command
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND, uint64(types.RS_Instrument))
}

func SendCustomTone(ent interfaces.Interpretable, r *memory.MemoryMap, p, d int) {
	// wait for channel
	for r.ReadGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND) != 0 {
		time.Sleep(50 * time.Microsecond)
	}
	// Send instrument data
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER_COUNT, 2)
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER+0, uint64(p))
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER+1, uint64(d))
	// Finalise command
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND, uint64(types.RS_Sound))
}

func SendSong(ent interfaces.Interpretable, r *memory.MemoryMap, songdata []byte) {
	// wait for channel
	for r.ReadGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND) != 0 {
		time.Sleep(50 * time.Microsecond)
	}
	// Send instrument data
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER_COUNT, uint64(len(songdata)))
	for i, v := range songdata {
		r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_BUFFER+i, uint64(v))
	}
	// Finalise command
	r.WriteGlobal(ent.GetMemIndex(), r.MEMBASE(ent.GetMemIndex())+memory.OCTALYZER_MUSIC_COMMAND, uint64(types.RS_PlaySong))
}

func SendNoteStream(ent interfaces.Interpretable, r *memory.MemoryMap, instrument string) {
	// wait for channel
	cmd := fmt.Sprintf(`
use mixer.voices.tone
set instrument "WAVE=SAWTOOTH:ADSR=0,0,500,500"
set notes "%s"
set volume 0.5
	`, instrument)
	ent.PassRestBufferNB(cmd)
}

func (this *PlusCustomSound) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	pitch := 0
	duration := 0
	instrument := ""

	pitch = this.Stack.Shift().AsInteger()
	duration = this.Stack.Shift().AsInteger()
	instrument = this.Stack.Shift().Content

	r := this.Interpreter.GetMemoryMap()
	if instrument != "" {
		SendInstrument(this.Interpreter, r, instrument)
	}
	SendCustomTone(this.Interpreter, r, pitch, duration)

	time.Sleep(time.Millisecond * time.Duration(duration))

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusCustomSound) Syntax() string {

	/* vars */
	var result string

	result = "SOUND{f,delay,instrument}"

	/* enforce non void return */
	return result

}

func (this *PlusCustomSound) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusCustomSound(a int, b int, params types.TokenList) *PlusCustomSound {
	this := &PlusCustomSound{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SOUND"

	return this
}
