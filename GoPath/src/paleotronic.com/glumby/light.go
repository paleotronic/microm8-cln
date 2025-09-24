package glumby

import (
	"paleotronic.com/gl"
)

type LightID uint32

const (
	Light0 LightID = LightID(gl.LIGHT0)
	Light1 LightID = LightID(gl.LIGHT1)
	Light2 LightID = LightID(gl.LIGHT2)
	Light3 LightID = LightID(gl.LIGHT3)
	Light4 LightID = LightID(gl.LIGHT4)
	Light5 LightID = LightID(gl.LIGHT5)
	Light6 LightID = LightID(gl.LIGHT6)
	Light7 LightID = LightID(gl.LIGHT7)
)

type LightSource struct {
	Id          LightID
	Ambient     []float32
	Diffuse     []float32
	Position    []float32
	DiffuseKnob float32
	AmbientKnob float32
}

func NewLightSource(id LightID, a []float32, d []float32, p []float32) *LightSource {
	this := &LightSource{Id: id, Ambient: a, Diffuse: d, Position: p, DiffuseKnob: 1, AmbientKnob: 1}
	return this
}

func (l *LightSource) On() {
	l.Update()
	gl.Enable(uint32(l.Id))
}

func (l *LightSource) Off() {
	gl.Disable(uint32(l.Id))
}

func knob(amount float32, values []float32) []float32 {
	out := make([]float32, len(values))
	for i, v := range values {
		out[i] = amount * v
		if out[i] > 1 {
			out[i] = 1
		}
	}
	return out
}

func (l *LightSource) Update() {
	ambient := knob(l.AmbientKnob, l.Ambient)
	diffuse := knob(l.DiffuseKnob, l.Diffuse)
	lightPosition := l.Position

	gl.Lightfv(uint32(l.Id), gl.AMBIENT, &ambient[0])
	gl.Lightfv(uint32(l.Id), gl.DIFFUSE, &diffuse[0])
	gl.Lightfv(uint32(l.Id), gl.POSITION, &lightPosition[0])
}

func (l *LightSource) SetAmbientLevel(f float32) {
	l.AmbientKnob = f
	l.Update()
}

func (l *LightSource) SetDiffuseLevel(f float32) {
	l.DiffuseKnob = f
	l.Update()
}
