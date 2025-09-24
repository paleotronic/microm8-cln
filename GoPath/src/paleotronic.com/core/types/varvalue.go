package types

import (
//	"paleotronic.com/fmt"
)

type VVContext struct {
	Value string
	Published bool
}

type VarValue struct {
	Value string
	Published bool
	Content []*VVContext
}

func NewVarValue( v string ) *VarValue {
	return &VarValue{ Content: make([]*VVContext, 0), Value: v }
}

func (vv *VarValue) Pending() int {
	return len(vv.Content)
}

func (vv *VarValue) String() string {
	return vv.GetValue()
}

func (vv *VarValue) GetValue() string {
	////fmt.Printf("Before GetValue() = %v\n", vv.Content)
	if len(vv.Content) > 0 {
		vv.Value = vv.Content[0].Value
		vv.Published = vv.Content[0].Published
		vv.Content = vv.Content[1:]
	}
	////fmt.Printf("After GetValue() = %v\n", vv.Content)
	return vv.Value
}

func (vv *VarValue) AssignStackedRemote( v string ) {
	vv.Content = append(vv.Content, &VVContext{Value: v, Published: true})
	////fmt.Printf("After AssignStackedRemote() = %v\n", vv.Content)
}

func (vv *VarValue) AssignStacked( v string ) {
	vv.Content = append(vv.Content, &VVContext{Value: v, Published: false})
	////fmt.Printf("After AssignStacked() = %v\n", vv.Content)
}

func (vv *VarValue) Assign( v string, stacked bool ) {
	if stacked {
		vv.AssignStacked(v)
	} else {
		vv.Value = v
		vv.Published = false
	}
}

