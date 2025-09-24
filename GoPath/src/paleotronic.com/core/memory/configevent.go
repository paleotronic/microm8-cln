package memory

type ConfigureSubSystem int

const (
	CSS_HGR ConfigureSubSystem = 1 << iota
	CSS_DHGR
	CSS_LAYERS
	CSS_RESTALGIA
	//
	CSS_MAX
)

type ConfigureEvent struct {
	Index     int
	SubSystem ConfigureSubSystem
}

func (ce *ConfigureEvent) GetItems() []ConfigureSubSystem {
	out := make([]ConfigureSubSystem, 0)
	for i := 0; 1<<uint(i) < int(CSS_MAX); i++ {
		if int(ce.SubSystem)&1<<uint(CSS_MAX) != 0 {
			out = append(out, ConfigureSubSystem(1<<uint(i)))
		}
	}
	return out
}

func (ce *ConfigureEvent) AddItem(css ConfigureSubSystem) {
	ce.SubSystem |= css
}

func (ce *ConfigureEvent) DelItem(css ConfigureSubSystem) {
	ce.SubSystem &= (0xffffff ^ css)
}

func NewConfigureEvent(index int, css ConfigureSubSystem) *ConfigureEvent {
	return &ConfigureEvent{
		Index:     index,
		SubSystem: css,
	}
}
