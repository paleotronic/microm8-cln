package mock

import (
	"math"

	"paleotronic.com/core/memory"
	"paleotronic.com/log"
)

const (
	ORB1               = 0x00 // 6522 registers
	ORB2               = 0x80
	ORA1               = 0x01
	ORA2               = 0x81
	DDRB1              = 0x02
	DDRB2              = 0x82
	DDRA1              = 0x03
	DDRA2              = 0x83
	PSGFreqFineA       = 0x00 // PSG Registers
	PSGFreqCourseA     = 0x01
	PSGFreqFineB       = 0x02
	PSGFreqCourseB     = 0x03
	PSGFreqFineC       = 0x04
	PSGFreqCourseC     = 0x05
	PSGFreqNG          = 0x06
	PSGEnableControl   = 0x07
	PSGLevelAndEnvA    = 0x08
	PSGLevelAndEnvB    = 0x09
	PSGLevelAndEnvC    = 0x0a
	PSGEnvPeriodFine   = 0x0b
	PSGEnvPeriodCourse = 0x0c
	PSGEnvShape        = 0x0d
)

const (
	PSGEnableNoiseOnC = 1 << 5
	PSGEnableNoiseOnB = 1 << 4
	PSGEnableNoiseOnA = 1 << 3
	PSGEnableToneOnC  = 1 << 2
	PSGEnableToneOnB  = 1 << 1
	PSGEnableToneOnA  = 1 << 0
)

type MemoryAccess interface {
	GetMemoryMap() *memory.MemoryMap
	GetMemIndex() int
}

type MockDriver struct {
	Base int
	Int  MemoryAccess
	regs [2][16]byte
}

func New(e MemoryAccess, base int) *MockDriver {
	m := &MockDriver{
		Base: base,
		Int:  e,
	}
	m.Reset(0)
	m.Reset(1)
	m.DDRA1(0xff)
	m.DDRA2(0xff)
	m.DDRB1(0xff)
	m.DDRB2(0xff)
	return m
}

func (m *MockDriver) ORB1(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+ORB1,
		uint64(value),
	)
}

func (m *MockDriver) ORB2(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+ORB2,
		uint64(value),
	)
}

func (m *MockDriver) ORA1(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+ORA1,
		uint64(value),
	)
}

func (m *MockDriver) GetORA1() byte {
	return byte(m.Int.GetMemoryMap().ReadInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+ORA1,
	))
}

func (m *MockDriver) ORA2(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+ORA2,
		uint64(value),
	)
}

func (m *MockDriver) GetORA2() byte {
	return byte(m.Int.GetMemoryMap().ReadInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+ORA2,
	))
}

func (m *MockDriver) DDRB1(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+DDRB1,
		uint64(value),
	)
}

func (m *MockDriver) DDRB2(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+DDRB2,
		uint64(value),
	)
}

func (m *MockDriver) DDRA1(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+DDRA1,
		uint64(value),
	)
}

func (m *MockDriver) DDRA2(value byte) {
	m.Int.GetMemoryMap().WriteInterpreterMemory(
		m.Int.GetMemIndex(),
		m.Base+DDRA2,
		uint64(value),
	)
}

func (m *MockDriver) SetRegNumber(chip byte) {
	switch chip {
	case 0:
		m.ORB1(0x07)
	case 1:
		m.ORB2(0x07)
	}
}

func (m *MockDriver) SetInactive(chip byte) {
	switch chip {
	case 0:
		m.ORB1(0x04)
	case 1:
		m.ORB2(0x04)
	}
}

func (m *MockDriver) WriteData(chip byte) {
	switch chip {
	case 0:
		m.ORB1(0x06)
	case 1:
		m.ORB2(0x06)
	}
}

func (m *MockDriver) ReadData(chip byte) {
	switch chip {
	case 0:
		m.ORB1(0x05)
	case 1:
		m.ORB2(0x05)
	}
}

func (m *MockDriver) Reset(chip byte) {
	switch chip {
	case 0:
		m.ORB1(0x00)
	case 1:
		m.ORB2(0x00)
	}
	m.ResetRegs(chip)
}

func (m *MockDriver) ResetRegs(chip byte) {
	m.regs[chip] = [16]byte{0, 0, 0, 0, 0, 0, 0, 255, 0, 0, 0, 0, 0, 0, 0, 0}
	for i, v := range m.regs[chip] {
		m.WritePSGRegister(chip, byte(i), v)
	}
}

func (m *MockDriver) SetA(chip byte, value byte) {
	switch chip {
	case 0:
		m.ORA1(value)
	case 1:
		m.ORA2(value)
	}
}

func (m *MockDriver) GetA(chip byte) byte {
	switch chip {
	case 0:
		return m.GetORA1()
	case 1:
		return m.GetORA2()
	}
	return 0x00
}

func (m *MockDriver) WritePSGRegister(chip byte, psgReg byte, psgValue byte) {
	m.SetA(chip, psgReg)
	m.SetRegNumber(chip)
	m.SetA(chip, psgValue)
	m.WriteData(chip)
	m.regs[int(chip)][int(psgReg)] = psgValue

	// readBack := m.ReadPSGRegister(chip, psgReg)
	// log.Printf("Read back of reg %d gives %d", psgReg, readBack)
}

func (m *MockDriver) ReadPSGRegister(chip byte, psgReg byte) byte {
	m.SetA(chip, psgReg)
	m.SetRegNumber(chip)
	m.ReadData(chip)
	switch chip {
	case 0:
		m.DDRA1(0x00)
	case 1:
		m.DDRA2(0x00)
	}
	v := m.GetA(chip)
	switch chip {
	case 0:
		m.DDRA1(0xff)
	case 1:
		m.DDRA2(0xff)
	}
	return v
}

func (m *MockDriver) SetToneACoarse(chip byte, freq byte) {
	clip := freq & 0x0f
	m.WritePSGRegister(chip, PSGFreqCourseA, byte(clip))
}

func (m *MockDriver) SetToneBCoarse(chip byte, freq byte) {
	clip := freq & 0x0f
	m.WritePSGRegister(chip, PSGFreqCourseB, byte(clip))
}

func (m *MockDriver) SetToneCCoarse(chip byte, freq byte) {
	clip := freq & 0x0f
	m.WritePSGRegister(chip, PSGFreqCourseC, byte(clip))
}

func (m *MockDriver) SetToneAFine(chip byte, freq byte) {
	clip := freq & 0xff
	m.WritePSGRegister(chip, PSGFreqFineA, byte(clip))
}

func (m *MockDriver) SetToneBFine(chip byte, freq byte) {
	clip := freq & 0xff
	m.WritePSGRegister(chip, PSGFreqFineB, byte(clip))
}

func (m *MockDriver) SetToneCFine(chip byte, freq byte) {
	clip := freq & 0xff
	m.WritePSGRegister(chip, PSGFreqFineC, byte(clip))
}

func (m *MockDriver) SetToneAFreq(chip byte, freq uint16) {
	clip := freq & 0x0fff
	m.WritePSGRegister(chip, PSGFreqFineA, byte(clip&0xff))
	m.WritePSGRegister(chip, PSGFreqCourseA, byte(clip>>8))
}

func (m *MockDriver) SetToneBFreq(chip byte, freq uint16) {
	clip := freq & 0x0fff
	m.WritePSGRegister(chip, PSGFreqFineB, byte(clip&0xff))
	m.WritePSGRegister(chip, PSGFreqCourseB, byte(clip>>8))
}

func (m *MockDriver) SetToneCFreq(chip byte, freq uint16) {
	clip := freq & 0x0fff
	m.WritePSGRegister(chip, PSGFreqFineC, byte(clip&0xff))
	m.WritePSGRegister(chip, PSGFreqCourseC, byte(clip>>8))
}

func (m *MockDriver) SetNoiseFreq(chip, freq byte) {
	m.WritePSGRegister(chip, PSGFreqNG, freq&0x1f)
}

func (m *MockDriver) SetVoiceEnableState(voice int, param byte) {
	chip := voice / 3
	var tone, noise byte
	switch voice % 3 {
	case 0:
		tone = PSGEnableToneOnA
		noise = PSGEnableNoiseOnA
	case 1:
		tone = PSGEnableToneOnB
		noise = PSGEnableNoiseOnB
	case 2:
		tone = PSGEnableToneOnC
		noise = PSGEnableNoiseOnC
	default:
		return
	}
	switch param {
	case 0x00:
		m.SetVoiceEnable(byte(chip), tone, false)
		m.SetVoiceEnable(byte(chip), noise, false)
	case 0x01:
		m.SetVoiceEnable(byte(chip), tone, false)
		m.SetVoiceEnable(byte(chip), noise, true)
	case 0x10:
		m.SetVoiceEnable(byte(chip), tone, true)
		m.SetVoiceEnable(byte(chip), noise, false)
	case 0x11:
		m.SetVoiceEnable(byte(chip), tone, true)
		m.SetVoiceEnable(byte(chip), noise, true)
	}
}

func (m *MockDriver) ResetEnv(track int) {
	chip := byte(track / 3)
	cval := m.ReadPSGRegister(chip, PSGEnvShape)
	m.WritePSGRegister(chip, PSGEnvShape, cval)
}

func (m *MockDriver) SetVoiceEnable(chip, params byte, enabled bool) {
	current := m.regs[int(chip)][PSGEnableControl]
	if enabled {
		// clear the bit
		current &= ^params
	} else {
		// set the bit
		current |= params
	}
	m.WritePSGRegister(chip, PSGEnableControl, current)
}

func (m *MockDriver) SetLevelA(chip byte, level byte) {
	plevel := m.regs[chip][PSGLevelAndEnvA]
	m.WritePSGRegister(chip, PSGLevelAndEnvA, (plevel&16)|(level&15))
}

func (m *MockDriver) SetLevelB(chip byte, level byte) {
	plevel := m.regs[chip][PSGLevelAndEnvB]
	m.WritePSGRegister(chip, PSGLevelAndEnvB, (plevel&16)|(level&15))
}

func (m *MockDriver) SetLevelC(chip byte, level byte) {
	plevel := m.regs[chip][PSGLevelAndEnvC]
	m.WritePSGRegister(chip, PSGLevelAndEnvC, (plevel&16)|(level&15))
}

func (m *MockDriver) SetEnvEnableA(chip byte, b bool) {
	level := m.regs[chip][PSGLevelAndEnvA]
	if b {
		level |= 16
	} else {
		level &= 15
	}
	m.WritePSGRegister(chip, PSGLevelAndEnvA, level)
}

func (m *MockDriver) SetEnvEnableB(chip byte, b bool) {
	level := m.regs[chip][PSGLevelAndEnvB]
	if b {
		level |= 16
	} else {
		level &= 15
	}
	m.WritePSGRegister(chip, PSGLevelAndEnvB, level)
}

func (m *MockDriver) SetEnvEnableC(chip byte, b bool) {
	level := m.regs[chip][PSGLevelAndEnvC]
	if b {
		level |= 16
	} else {
		level &= 15
	}
	m.WritePSGRegister(chip, PSGLevelAndEnvC, level)
}

func (m *MockDriver) SetEnvelopePeriod(chip byte, period uint16) {
	m.WritePSGRegister(chip, PSGEnvPeriodFine, byte(period&0xff))
	m.WritePSGRegister(chip, PSGEnvPeriodCourse, byte(period>>8))
}

func (m *MockDriver) SetEnvelopeShape(chip byte, shape byte) {
	m.WritePSGRegister(chip, PSGEnvShape, byte(shape&0xf))
}

func (m *MockDriver) Squelch(voice int) {
	chip := byte(voice / 3)
	switch voice % 3 {
	case 0:
		m.SetVoiceEnable(chip, PSGEnableToneOnA, false)
	case 1:
		m.SetVoiceEnable(chip, PSGEnableToneOnB, false)
	case 2:
		m.SetVoiceEnable(chip, PSGEnableToneOnC, false)
	}
	switch voice % 3 {
	case 0:
		m.SetVoiceEnable(chip, PSGEnableNoiseOnA, false)
	case 1:
		m.SetVoiceEnable(chip, PSGEnableNoiseOnB, false)
	case 2:
		m.SetVoiceEnable(chip, PSGEnableNoiseOnC, false)
	}
	m.ApplyVoiceVolume(voice, 0)
}

func (m *MockDriver) AdjustTonePeriodCoarse(voice int, diff int) {
	chip := byte(voice / 3)

	var c int
	switch voice % 3 {
	case 0:
		c = int(m.regs[chip][PSGFreqCourseA])
	case 1:
		c = int(m.regs[chip][PSGFreqCourseB])
	case 2:
		c = int(m.regs[chip][PSGFreqCourseC])
	}

	c += diff
	if c < 0 {
		c = 0
	}
	if c > 15 {
		c = 15
	}

	switch voice % 3 {
	case 0:
		m.SetToneACoarse(chip, byte(c))
	case 1:
		m.SetToneBCoarse(chip, byte(c))
	case 2:
		m.SetToneCCoarse(chip, byte(c))
	}
}

func (m *MockDriver) SetTrackTonePeriodFine(voice int, param byte) {
	chip := byte(voice / 3)
	switch voice % 3 {
	case 0:
		m.SetToneAFine(chip, param)
	case 1:
		m.SetToneBFine(chip, param)
	case 2:
		m.SetToneCFine(chip, param)
	}
}

func (m *MockDriver) AdjustTonePeriodFine(voice int, diff int) {
	chip := byte(voice / 3)

	var c int
	switch voice % 3 {
	case 0:
		c = int(m.regs[chip][PSGFreqFineA])
	case 1:
		c = int(m.regs[chip][PSGFreqFineB])
	case 2:
		c = int(m.regs[chip][PSGFreqFineC])
	}

	c += diff
	if c < 0 {
		c = 0
	}
	if c > 255 {
		c = 255
	}

	switch voice % 3 {
	case 0:
		m.SetToneAFine(chip, byte(c))
	case 1:
		m.SetToneBFine(chip, byte(c))
	case 2:
		m.SetToneCFine(chip, byte(c))
	}
}

func (m *MockDriver) AdjustEnvPeriodCoarse(voice int, diff int) {
	chip := byte(voice / 3)

	var c int

	c = int(m.regs[chip][PSGEnvPeriodCourse])

	c += diff
	if c < 0 {
		c = 0
	}
	if c > 255 {
		c = 255
	}

	//m.SetEnvelopePeriod(chip, (uint16(c)<<8)|uint16(m.regs[chip][PSGEnvPeriodFine]))
	m.WritePSGRegister(chip, PSGEnvPeriodCourse, byte(c&0xff))
}

func (m *MockDriver) AdjustEnvPeriodFine(voice int, diff int) {
	chip := byte(voice / 3)

	var c int

	c = int(m.regs[chip][PSGEnvPeriodFine])

	log.Printf("Current Env fine value for voice %d is $%.2x", voice, c)

	c += diff
	if c < 0 {
		c = 0
	}
	if c > 255 {
		c = 255
	}

	m.WritePSGRegister(chip, PSGEnvPeriodFine, byte(c&0xff))

	//m.SetEnvelopePeriod(chip, uint16(c)|(uint16(m.regs[chip][PSGEnvPeriodCourse])<<8))

	log.Printf("Updated Env fine value for voice %d is $%.2x", voice, c)
}

func (m *MockDriver) AdjustVoiceVolume(voice int, diff int) {
	chip := byte(voice / 3)

	var c int
	switch voice % 3 {
	case 0:
		c = int(m.regs[chip][PSGLevelAndEnvA])
	case 1:
		c = int(m.regs[chip][PSGLevelAndEnvB])
	case 2:
		c = int(m.regs[chip][PSGLevelAndEnvC])
	}

	c += diff
	if c < 0 {
		c = 0
	}
	if c > 15 {
		c = 15
	}

	switch voice % 3 {
	case 0:
		m.SetLevelA(chip, byte(c))
	case 1:
		m.SetLevelB(chip, byte(c))
	case 2:
		m.SetLevelC(chip, byte(c))
	}
}

func (m *MockDriver) ApplyVoiceVolume(voice int, level byte) {
	chip := byte(voice / 3)
	switch voice % 3 {
	case 0:
		m.SetLevelA(chip, byte(level))
	case 1:
		m.SetLevelB(chip, byte(level))
	case 2:
		m.SetLevelC(chip, byte(level))
	}
}

func (m *MockDriver) SetVoiceNoisePeriod(voice int, period byte) {
	chip := byte(voice / 3)
	m.SetNoiseFreq(chip, period)
}

func (m *MockDriver) SetVoiceEnvPeriodCoarse(voice int, period byte) {
	chip := byte(voice / 3)
	m.WritePSGRegister(chip, PSGEnvPeriodCourse, period)
}

func (m *MockDriver) SetVoiceEnvPeriodFine(voice int, period byte) {
	chip := byte(voice / 3)
	m.WritePSGRegister(chip, PSGEnvPeriodFine, period)
}

func (m *MockDriver) AppleVoiceEnvelope(voice int, enabled bool) {
	chip := byte(voice / 3)
	switch voice % 3 {
	case 0:
		m.SetEnvEnableA(chip, enabled)
	case 1:
		m.SetEnvEnableB(chip, enabled)
	case 2:
		m.SetEnvEnableC(chip, enabled)
	}
}

func (m *MockDriver) ApplyVoice(
	voice int,
	useTone bool,
	toneFreq uint16,
	useNoise bool,
	noiseFreq byte,
	level int,
	useEnv bool,
	envFreq uint16,
	envShape int,
) {

	log.Printf("Applying voice %d, %v", voice, useNoise)

	chip := byte(voice / 3)
	if useTone {
		switch voice % 3 {
		case 0:
			m.SetVoiceEnable(chip, PSGEnableToneOnA, true)
			m.SetToneAFreq(chip, toneFreq)
		case 1:
			m.SetVoiceEnable(chip, PSGEnableToneOnB, true)
			m.SetToneBFreq(chip, toneFreq)
		case 2:
			m.SetVoiceEnable(chip, PSGEnableToneOnC, true)
			m.SetToneCFreq(chip, toneFreq)
		}
	} else {
		switch voice % 3 {
		case 0:
			m.SetVoiceEnable(chip, PSGEnableToneOnA, false)
		case 1:
			m.SetVoiceEnable(chip, PSGEnableToneOnB, false)
		case 2:
			m.SetVoiceEnable(chip, PSGEnableToneOnC, false)
		}
	}
	if useNoise {
		switch voice % 3 {
		case 0:
			m.SetVoiceEnable(chip, PSGEnableNoiseOnA, true)
			m.SetNoiseFreq(chip, noiseFreq)
		case 1:
			m.SetVoiceEnable(chip, PSGEnableNoiseOnB, true)
			m.SetNoiseFreq(chip, noiseFreq)
		case 2:
			m.SetVoiceEnable(chip, PSGEnableNoiseOnC, true)
			m.SetNoiseFreq(chip, noiseFreq)
		}
	} else {
		switch voice % 3 {
		case 0:
			m.SetVoiceEnable(chip, PSGEnableNoiseOnA, false)
		case 1:
			m.SetVoiceEnable(chip, PSGEnableNoiseOnB, false)
		case 2:
			m.SetVoiceEnable(chip, PSGEnableNoiseOnC, false)
		}
	}
	switch voice % 3 {
	case 0:
		m.SetLevelA(chip, byte(level))
	case 1:
		m.SetLevelB(chip, byte(level))
	case 2:
		m.SetLevelC(chip, byte(level))
	}
	if useEnv {
		m.SetEnvelopePeriod(chip, envFreq)
		m.SetEnvelopeShape(chip, byte(envShape))
		switch voice % 3 {
		case 0:
			m.SetEnvEnableA(chip, true)
		case 1:
			m.SetEnvEnableB(chip, true)
		case 2:
			m.SetEnvEnableC(chip, true)
		}
	} else {
		switch voice % 3 {
		case 0:
			m.SetEnvEnableA(chip, false)
		case 1:
			m.SetEnvEnableB(chip, false)
		case 2:
			m.SetEnvEnableC(chip, false)
		}
	}
}

// FreqToEnvPeriod
// **Env Freq = A2 Clock Freq/ [ (65536 x Coarse) + (256 x Fine) ]
// freq * (65536 x coarse + 256 x fine) = a2 clock freq
// (65536 x coarse + 256 x fine) = a2 clock freq / freq
// 256 x course + fine = a2 clock freq / (freq * 256)

const a2ClockFreq = 1020484

func round(f float64) uint16 {
	return uint16(math.Floor(f + 0.5))
}

func EnvPeriodToFreqHz(period uint16) float64 {
	fr := a2ClockFreq / (256 * float64(period))
	return fr
}

func FreqHzToEnvPeriod(fr float64) uint16 {
	period := round((a2ClockFreq / fr) / 256)
	return period
}

func TonePeriodToFreqHz(period uint16) float64 {
	fr := a2ClockFreq / (16 * float64(period))
	return fr
}

func FreqHzToTonePeriod(fr float64) uint16 {
	period := round((a2ClockFreq / fr) / 16)
	return period
}

func NoisePeriodToFreqHz(period byte) float64 {
	fr := a2ClockFreq / (16 * float64(period))
	return fr
}

func FreqHzToNoisePeriod(fr float64) byte {
	period := byte(round((a2ClockFreq / fr) / 16))
	return period
}
