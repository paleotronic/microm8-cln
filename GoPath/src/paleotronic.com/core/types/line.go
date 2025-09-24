// line.go
package types

type Line []Statement

func NewLine() Line {
	return make(Line, 0)
}

func (this *Line) Push(st Statement) {
	a := *this
	a = append(a, st)
	*this = a
}

func (this Line) String() string {

     out := ""

     for _, st := range this {
         if out != "" {
            out = out + " : ";
         }
         out = out + st.AsString()
     }

     return out
}
