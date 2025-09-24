package apple2rom

import "paleotronic.com/core/hardware/cpu/mos6502"

/*
 * WAIT
 * 
 * Description: This subroutine delays for a specific amount of time.
 * 
 * Input      : cpu.A
 * Output     : cpu.A <- 0
 */
var ROMCALL_FCA8 mos6502.Func6502 = func( cpu *mos6502.Core6502 ) int64 {
	
	var microseconds int64
	
	microseconds = int64( 0.5102 * float64(26 + (27*cpu.A) + (5 * cpu.A * cpu.A) ) )
	
	cpu.A = 0 // zero A register
	
	return microseconds
}
