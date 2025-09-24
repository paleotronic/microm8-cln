package video

import "paleotronic.com/glumby"
import "paleotronic.com/gl"
import "paleotronic.com/octalyzer/assets"
import "bytes"

type LED struct {
	State func() bool
	StateBool    bool
	OnImage string
	OffImage string
	OnTex *glumby.Texture
	OffTex *glumby.Texture
	X, Y, Z float32
	W, H float32
	m *glumby.Mesh
}

func assetBytes(path string) (*bytes.Buffer, error) {
	b, e := assets.Asset( path )
	buffer := bytes.NewBuffer(b)
	return buffer, e
}

func (l *LED) Init() {
	if l.OnImage != "" {
		buff, e := assetBytes( l.OnImage )
		if e != nil {
			panic(e)
		}
		l.OnTex, e = glumby.NewTextureFromBytes( buff )
		if e != nil {
			panic(e)
		}
	}
	if l.OffImage != "" {
		buff, e := assetBytes( l.OffImage )
		if e != nil {
			panic(e)
		}
		l.OffTex, e = glumby.NewTextureFromBytes( buff )
		if e != nil {
			panic(e)
		}
	}
	l.m = GetPlaneAsTriangles(l.W, l.H)
}

func (l *LED) Draw(x, y float32) {
	
	// If we have a custom func, run it to refresh the state
	if l.State != nil {
		l.StateBool = l.State()
	}
	
	if l.StateBool {
		if l.OnTex  != nil {
			l.OnTex.Bind()
			glumby.MeshBuffer_Begin(gl.TRIANGLES)
			l.m.DrawWithMeshBuffer( l.X+x, l.Y+y, l.Z )
			glumby.MeshBuffer_End()
			l.OnTex.Unbind()
		}
	} else {
		if l.OffTex != nil {
			l.OffTex.Bind()
			glumby.MeshBuffer_Begin(gl.TRIANGLES)
			l.m.DrawWithMeshBuffer( l.X+x, l.Y+y, l.Z )
			glumby.MeshBuffer_End()
			l.OffTex.Unbind()			
		}
	}
	
}

func NewLED( onimage string, offimage string, x, y, z, w, h float32, statebool bool, statefunc func()bool ) *LED {
	
	l := &LED{
		OnImage: onimage,
		OffImage: offimage, 
		X: x, Y: y, Z: z,
		W: w, H: h,
		StateBool: statebool,
		State: statefunc,
	}
	
	l.Init()
	
	return l
	
}
