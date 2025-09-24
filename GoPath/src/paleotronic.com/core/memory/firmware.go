package memory

type Firmware interface {
	FirmwareRead(offset int) uint64
	FirmwareWrite(offset int, value uint64)
	FirmwareExec(offset int, PC, A, X, Y, SP, P *int) int64
	GetID() int
}
