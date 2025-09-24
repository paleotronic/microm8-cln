package mos6502

/*
 * This package defines a special shim interface for defining calls that make changes to the CPU
 * state.
 * 
 * It allows us for example to simulate a ROM code call for the processor, modifying the registers,
 * returning an appropriate cycle penalty etc.
 * 
 */

// Func6502 is a simple to implement shim interface... return value is cycle penalty for mimicking 
// the correct timings.

type Func6502 func (cpu *Core6502) int64

