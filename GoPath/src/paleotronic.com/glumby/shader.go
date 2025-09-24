package glumby

import (
	"reflect"
	"strings"
	"unsafe"

	"paleotronic.com/gl"
)

// #include <stdlib.h>
import "C"

type Shader struct {
	id     uint32
	source string
}

type Program struct {
	id uint32
}

func NewSimpleProgram(vertex, fragment *Shader) *Program {
	this := &Program{}

	this.id = gl.CreateProgram()

	if vertex != nil {
		gl.AttachShader(this.id, vertex.id)
	}

	if fragment != nil {
		gl.AttachShader(this.id, fragment.id)
	}

	gl.LinkProgram(this.id)

	return this
}

func (p *Program) Use() {
	gl.UseProgram(p.id)
}

func (p *Program) UseFixedPipeline() {
	gl.UseProgram(0)
}

func Str(str string) *uint8 {
	if !strings.HasSuffix(str, "\x00") {
		panic("str argument missing null terminator: " + str)
	}
	header := (*reflect.StringHeader)(unsafe.Pointer(&str))
	return (*uint8)(unsafe.Pointer(header.Data))
}

func Strs(strs ...string) (cstrs **uint8, free func()) {
	if len(strs) == 0 {
		panic("Strs: expected at least 1 string")
	}

	// Allocate a contiguous array large enough to hold all the strings' contents.
	n := 0
	for i := range strs {
		n += len(strs[i])
	}
	data := C.malloc(C.size_t(n))

	// Copy all the strings into data.
	dataSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(data),
		Len:  n,
		Cap:  n,
	}))
	css := make([]*uint8, len(strs)) // Populated with pointers to each string.
	offset := 0
	for i := range strs {
		copy(dataSlice[offset:offset+len(strs[i])], strs[i][:]) // Copy strs[i] into proper data location.
		css[i] = (*uint8)(unsafe.Pointer(&dataSlice[offset]))   // Set a pointer to it.
		offset += len(strs[i])
	}

	return (**uint8)(&css[0]), func() { C.free(data) }
}

func NewShader(source string, kind uint32) (*Shader, error) {
	this := &Shader{}

	this.id = gl.CreateShader(kind)

	if !strings.HasSuffix(source, "\x00") {
		source += "\x00"
	}

	this.source = source

	//data := []uint8(source)

	csources, free := Strs(source)

	gl.ShaderSource(this.id, 1, csources, nil)
	free()

	// compile shader
	gl.CompileShader(this.id)

	return this, nil
}

func NewFragmentShader(source string) (*Shader, error) {
	return NewShader(source, gl.FRAGMENT_SHADER)
}

func NewVertexShader(source string) (*Shader, error) {
	return NewShader(source, gl.VERTEX_SHADER)
}
