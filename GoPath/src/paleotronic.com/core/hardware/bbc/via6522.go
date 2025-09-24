package bbc

type Interruptable interface {
	PullIRQLine()
}

type VIAState struct {
	ora, orb         byte
	ira, irb         byte
	ddra, ddrb       byte
	acr, pcr         byte
	ifr, ier         byte
	timer1c, timer2c int  /* NOTE: Timers descrement at 2MHz and values are */
	timer1l, timer2l int  /*   fixed up on read/write - latches hold 1MHz values*/
	timer1hasshot    bool /* True if we have already caused an interrupt for one shot mode */
	timer2hasshot    bool /* True if we have already caused an interrupt for one shot mode */
	timer1adjust     int  // Adjustment for 1.5 cycle counts, every other interrupt, it becomes 2 cycles instead of one
	timer2adjust     int  // Adjustment for 1.5 cycle counts, every other interrupt, it becomes 2 cycles instead of on
}

func (v *VIAState) Reset() {
	v.ora = 0xff
	v.orb = 0xff
	v.ira = 0xff
	v.irb = 0xff
	v.ddra = 0
	v.ddrb = 0   /* All inputs */
	v.acr = 0    /* Timed ints on t1, t2, no pb7 hacking, no latching, no shifting */
	v.pcr = 0    /* Neg edge inputs for cb2,ca2 and CA1 and CB1 */
	v.ifr = 0    /* No interrupts presently interrupting */
	v.ier = 0x80 /* No interrupts enabled */
	v.timer1l = 0xffff
	v.timer2l = 0xffff /*0xffff; */
	v.timer1c = 0xffff
	v.timer2c = 0xffff /*0x1ffff; */
	v.timer1hasshot = false
	v.timer2hasshot = false
	v.timer1adjust = 0
	v.timer2adjust = 0
}
