package logo

import (
	"errors"
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
)

type StandardCommandTOOT struct {
	dialect.Command
}

func (this *StandardCommandTOOT) Syntax() string {

	/* vars */
	var result string

	result = "TOOT"

	/* enforce non void return */
	return result

}

func (this *StandardCommandTOOT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	t, _ := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)

	if t.Type != types.LIST || t.List.Size() != 2 {
		return result, errors.New("I NEED 2 VALUES")
	}

	freq := t.List.Shift().AsInteger()
	dur := 1000 / 60 * t.List.Shift().AsInteger()

	SendCustomTone(caller, caller.GetMemoryMap(), freq, dur)

	time.Sleep(time.Duration(dur) * time.Millisecond)

	/* enforce non void return */
	return result, nil

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
