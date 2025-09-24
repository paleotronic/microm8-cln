package logo

func (d *LogoDriver) SetREM(s string) {
	d.lastRem = s
}

func (d *LogoDriver) LastRem() string {
	return d.lastRem
}
