package restalgia

import (
	"strings"

	"paleotronic.com/fmt"
)

type RQueryable interface {
	GetLabel() string
	SetLabel(label string)
	GetKind() string
	SetKind(kind string)
	GetAttributes() map[string]*RAttribute
	GetObjects() map[string]*RObject
}

type RAttribute struct {
	label   string
	kind    string
	indexed bool
	get     func(index int) interface{}
	set     func(index int, v interface{})
	length  func() int
}

type RObject struct {
	label   string
	kind    string
	indexed bool
	get     func(index int) RQueryable
	set     func(index int, v RQueryable)
	length  func() int
}

type RInfo struct {
	label      string
	kind       string
	attributes map[string]*RAttribute
	objects    map[string]*RObject
}

func (a *RAttribute) Set(index int, value interface{}) {
	if a.set != nil {
		a.set(index, value)
	}
}

func (a *RAttribute) Get(index int) interface{} {
	if a.get != nil {
		a.get(index)
	}
	return nil
}

func NewRInfo(label string, kind string) *RInfo {
	return &RInfo{
		label:      label,
		kind:       kind,
		attributes: make(map[string]*RAttribute),
		objects:    make(map[string]*RObject),
	}
}

func (ri *RInfo) GetAttributes() map[string]*RAttribute {
	return ri.attributes
}

func (ri *RInfo) GetObjects() map[string]*RObject {
	return ri.objects
}

func (ri *RInfo) GetLabel() string {
	return ri.label
}

func (ri *RInfo) SetLabel(v string) {
	ri.label = v
}

func (ri *RInfo) GetKind() string {
	return ri.kind
}

func (ri *RInfo) SetKind(v string) {
	ri.kind = v
}

func (ri *RInfo) AddAttribute(label string, kind string, indexed bool, getter func(index int) interface{}, setter func(index int, v interface{}), length func() int) {
	ri.attributes[label] = &RAttribute{
		label:   label,
		kind:    kind,
		indexed: indexed,
		get:     getter,
		set:     setter,
		length:  length,
	}
}

func (ri *RInfo) AddObject(label string, kind string, indexed bool, getter func(index int) RQueryable, setter func(index int, v RQueryable), length func() int) {
	ri.objects[label] = &RObject{
		label:   label,
		kind:    kind,
		indexed: indexed,
		get:     getter,
		set:     setter,
		length:  length,
	}
}

func lastSegment(path string) string {
	p := strings.Split(path, ".")
	return p[len(p)-1]
}

func ResolveField(path string, o RQueryable) (interface{}, string, *RAttribute, bool) {

	/// query = mixer.voices.beeper.volume
	thing := o
	segments := strings.Split(path, ".")

	if segments[0] != o.GetLabel() {
		return nil, "", nil, false
	}

	segments = segments[1:]

	for len(segments) > 0 {
		target := segments[0]
		segments = segments[1:]

		if len(segments) > 0 {
			objs := thing.GetObjects()
			if newthing, ok := objs[target]; ok {

				if newthing.indexed {
					starget := segments[0]
					segments = segments[1:]
					found := false
					for i := 0; i < newthing.length(); i++ {
						if newthing.get(i).GetLabel() == starget {
							fmt.Printf("subobject ok: %s\n", target+"."+starget)
							thing = newthing.get(i)
							found = true
							break
						}
					}
					if found {
						if len(segments) == 0 {
							return thing, thing.GetKind(), nil, true
						}
						continue
					}
					fmt.Printf("subobject not found: %s\n", target+"."+starget)
					return nil, "", nil, false
				}

				fmt.Printf("object ok: %s\n", target)
				thing = newthing.get(0)
				continue
			}
			fmt.Printf("object not found: %s\n", target)
			return nil, "", nil, false
		}

		attr := thing.GetAttributes()
		//fmt.Println("Attrdump:", attr)
		if a, ok := attr[target]; ok {
			fmt.Printf("attr ok: %s\n", target)
			return a.get(0), a.kind, a, true
		}

		fmt.Printf("attr not found: %s\n", target)
	}

	return nil, "", nil, false

}

func QueryObjectTree(base string, o RQueryable) map[string]interface{} {

	if base != "" {
		base += "."
	}

	var out = make(map[string]interface{})

	if o == nil {
		return out
	}

	// attributes first
	attrs := o.GetAttributes()
	for name, attr := range attrs {
		if attr.indexed {
			for i := 0; i < attr.length(); i++ {
				out[base+name+fmt.Sprintf("[%d]", i)] = attr.Get(i)
			}
		} else {
			out[base+name] = attr.Get(0)
		}
	}

	// objects
	objs := o.GetObjects()
	for name, attr := range objs {
		if attr.indexed {
			for i := 0; i < attr.length(); i++ {
				tobj := attr.get(i)
				out[base+name] = tobj
				tmp := QueryObjectTree(base+name+"."+tobj.GetLabel(), tobj)
				for k, v := range tmp {
					out[k] = v
				}
			}
		} else {
			tobj := attr.get(0)
			out[base+name] = tobj
			tmp := QueryObjectTree(base+name, tobj)
			for k, v := range tmp {
				out[k] = v
			}
		}
	}

	return out

}
