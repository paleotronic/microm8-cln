package common

import "log"

type SerialPrinterEmu struct {
	emu Parallel
	bufferSize int
}

func NewSerialPrinterEmu(p Parallel, bufferSize int) *SerialPrinterEmu {
	log.Printf("SerialPrinterEmu: initialized")
	return &SerialPrinterEmu{emu: p, bufferSize: bufferSize}
}

func (sp *SerialPrinterEmu) CanSend() bool {
	return sp.emu.BufferCount() < (sp.bufferSize - 8)
}


func (sp *SerialPrinterEmu) IsConnected() bool {
	return true
}

func (sp *SerialPrinterEmu) InputAvailable() bool {
	return true
}

func (sp *SerialPrinterEmu) GetInputByte() int {
	return 0x00
}

func (sp *SerialPrinterEmu) SendOutputByte(value int) {
	//if value > 0x7f {
	//	log.Printf("HIGH BYTE!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! %.2x !!!!", value)
	//}
	if !sp.CanSend() {
		log.Printf("SerialPrinterEmu: write of byte %.2x failed due to full buffer", value)
		return
	}
	// log.Printf("SerialPrinterEmu: buffer %d/%d", sp.emu.BufferCount(), sp.bufferSize)
	sp.emu.Write([]byte{byte(value)})
}

func (sp *SerialPrinterEmu) Start() {
}

func (sp *SerialPrinterEmu) Stop() {
	sp.emu.Close()
}
