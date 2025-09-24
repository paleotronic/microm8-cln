package snapshot

type Z80State struct {
	RegA, RegB, RegC, RegD, RegE, RegH, RegL byte
	RegF                                     byte
	FlagQ                                    bool
	RegAx                                    byte
	RegFx                                    byte
	RegBx, RegCx, RegDx, RegEx, RegHx, RegLx byte
	RegPC                                    uint16
	RegIX                                    uint16
	RegIY                                    uint16
	RegSP                                    uint16
	RegI                                     byte
	RegR                                     byte
	IFF1                                     byte
	IFF2                                     byte
	PendingEI                                bool
	ActiveNMI                                bool
	ActiveINT                                bool
	ModeINT                                  byte
	Halted                                   bool
	Tstates                                  int
}
