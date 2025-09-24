package main

import (
	"paleotronic.com/fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"image/color"
	"paleotronic.com/log"
	"paleotronic.com/glumby"
	"runtime"
	"time"
)

var (
	t1, t2        *glumby.Texture
	light, light2 *glumby.LightSource
	rotationX     float32
	rotationY     float32
	meshy         *glumby.Mesh
	cammy         *glumby.Camera
	distance_z    float32
)

func init() {
	runtime.LockOSThread()
}

func main() {

	props := glumby.NewWindowProperties()

	props.Title = "How absolutely boring, eh?"
	props.Width = 800

	w := glumby.NewWindow(props)

	w.OnCreate = OnCreateWindow
	w.OnDestroy = OnDestroyWindow
	w.OnRender = OnRenderWindow
	w.OnEvent = OnEventWindow
	w.OnKey = OnKeyEvent
	w.OnChar = OnCharEvent
	w.OnMouseMove = OnMouseMoveEvent
	w.OnMouseButton = OnMouseButtonEvent
	w.OnScroll = OnScrollEvent

	w.Run()
}

func OnCreateWindow(w *glumby.Window) {

	// Create a texture
	t1 = glumby.NewSolidColorTexture(color.RGBA{255, 128, 255, 255})
	t2 = glumby.NewSolidColorTexture(color.RGBA{255, 255, 255, 128})

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.LIGHTING)
	gl.Enable(gl.COLOR_MATERIAL)
	gl.ColorMaterial(gl.FRONT, gl.AMBIENT)
	gl.ClearColor(0.9, 0.5, 0.5, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	// Lights
	// Add a glumby light
	ambient := []float32{0.5, 0.5, 0.5, 1}
	diffuse := []float32{1, 0, 1, 1}
	ambient2 := []float32{0.5, 0.5, 0.5, 1}
	diffuse2 := []float32{1, 1, 1, 1}
	lightPosition := []float32{-10, 10, 10, 0}
	lightPosition2 := []float32{20, -20, 20, 0}
	light = glumby.NewLightSource(glumby.Light0, ambient, diffuse, lightPosition)
	light2 = glumby.NewLightSource(glumby.Light1, ambient2, diffuse2, lightPosition2)
	light.Off()
	light2.On()

	// Camera
	cammy = glumby.NewCamera(glumby.Rect{-2, 2, -1.3, 1.3}, 0.1, 1000, true)

	// mesh
	meshy = buildColorMesh(1, 1, 1, 1)

	distance_z = -60

}

func OnDestroyWindow(w *glumby.Window) {
	log.Println("Destroy called")
}

func buildColorMesh(r, g, b, a float32) *glumby.Mesh {
	m := glumby.NewMesh(gl.QUADS)

	m.Normal3f(0, 0, 1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, -1, 1)

	m.Normal3f(0, 0, 1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, -1, 1)

	m.Normal3f(0, 0, 1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, 1, 1)

	m.Normal3f(0, 0, 1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, 1, 1)

	m.Normal3f(0, 0, -1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, -1, -1)

	m.Normal3f(0, 0, -1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, 1, -1)

	m.Normal3f(0, 0, -1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, 1, -1)

	m.Normal3f(0, 0, -1)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, -1, -1)

	m.Normal3f(0, 1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, 1, -1)

	m.Normal3f(0, 1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, 1, 1)

	m.Normal3f(0, 1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, 1, 1)

	m.Normal3f(0, 1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, 1, -1)

	m.Normal3f(0, -1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, -1, -1)

	m.Normal3f(0, -1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, -1, -1)

	m.Normal3f(0, -1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, -1, 1)

	m.Normal3f(0, -1, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, -1, 1)

	m.Normal3f(1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, -1, -1)

	m.Normal3f(1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, 1, -1)

	m.Normal3f(1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, 1, 1)

	m.Normal3f(1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(1, -1, 1)

	m.Normal3f(-1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, -1, -1)

	m.Normal3f(-1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, -1, 1)

	m.Normal3f(-1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, 1, 1)

	m.Normal3f(-1, 0, 0)
	m.Color4f(r, g, b, a)
	m.Vertex3f(-1, 1, -1)

	return m
}

func buildMesh() *glumby.Mesh {
	m := glumby.NewMesh(gl.QUADS)

	m.Normal3f(0, 0, 1)
	m.TexCoord2f(0, 0)
	m.Vertex3f(-1, -1, 1)
	m.TexCoord2f(1, 0)
	m.Vertex3f(1, -1, 1)
	m.TexCoord2f(1, 1)
	m.Vertex3f(1, 1, 1)
	m.TexCoord2f(0, 1)
	m.Vertex3f(-1, 1, 1)
	m.Normal3f(0, 0, -1)
	m.TexCoord2f(1, 0)
	m.Vertex3f(-1, -1, -1)
	m.TexCoord2f(1, 1)
	m.Vertex3f(-1, 1, -1)
	m.TexCoord2f(0, 1)
	m.Vertex3f(1, 1, -1)
	m.TexCoord2f(0, 0)
	m.Vertex3f(1, -1, -1)
	m.Normal3f(0, 1, 0)
	m.TexCoord2f(0, 1)
	m.Vertex3f(-1, 1, -1)
	m.TexCoord2f(0, 0)
	m.Vertex3f(-1, 1, 1)
	m.TexCoord2f(1, 0)
	m.Vertex3f(1, 1, 1)
	m.TexCoord2f(1, 1)
	m.Vertex3f(1, 1, -1)
	m.Normal3f(0, -1, 0)
	m.TexCoord2f(1, 1)
	m.Vertex3f(-1, -1, -1)
	m.TexCoord2f(0, 1)
	m.Vertex3f(1, -1, -1)
	m.TexCoord2f(0, 0)
	m.Vertex3f(1, -1, 1)
	m.TexCoord2f(1, 0)
	m.Vertex3f(-1, -1, 1)
	m.Normal3f(1, 0, 0)
	m.TexCoord2f(1, 0)
	m.Vertex3f(1, -1, -1)
	m.TexCoord2f(1, 1)
	m.Vertex3f(1, 1, -1)
	m.TexCoord2f(0, 1)
	m.Vertex3f(1, 1, 1)
	m.TexCoord2f(0, 0)
	m.Vertex3f(1, -1, 1)
	m.Normal3f(-1, 0, 0)
	m.TexCoord2f(0, 0)
	m.Vertex3f(-1, -1, -1)
	m.TexCoord2f(1, 0)
	m.Vertex3f(-1, -1, 1)
	m.TexCoord2f(1, 1)
	m.Vertex3f(-1, 1, 1)
	m.TexCoord2f(0, 1)
	m.Vertex3f(-1, 1, -1)

	return m
}

func OnRenderWindow(w *glumby.Window) {

	//gl.CullFace(gl.BACK)

	gl.ClearColor(0, 0, 0, 0.0)
	gl.ClearDepth(1)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	cammy.SetPos(0, 0, 10)
	cammy.RotateX(0.05)
	cammy.Apply()

	rotationX += 0.5
	rotationY += 0.5
	//t2.Bind()
	t2.Unbind()
	gl.Color4f(1, 1, 1, 1)

	now := time.Now().UnixNano()
	glumby.MeshBuffer_Begin(gl.QUADS)
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			if (x+y)%2 == 0 {
				meshy.SetColor(1, 0, 0, 1)
			} else {
				meshy.SetColor(0, 0, 1, 1)
			}
			// draw cube
			//if (x+y)%2 == 1 {
			meshy.DrawWithMeshBuffer(float32(2*x), float32(2*y), 0)
			//}
		}
	}
	glumby.MeshBuffer_End()
	after := time.Now().UnixNano()
	log.Printf("Cubez drawn in %v ms\n", ((after - now) / 1000000))

	//w.SwapBuffers()
}

func OnEventWindow(w *glumby.Window) {
	log.Println("Event called")
}

func OnKeyEvent(w *glumby.Window, ch glumby.Key, mod glumby.ModifierKey, state glumby.Action) {
	log.Printf("Keyevent: ch = %d, mod = %d, state = %d\n", ch, mod, state)
}

func OnCharEvent(w *glumby.Window, ch rune) {
	log.Printf("Keyevent: ch = %d\n", ch)

	w.SetTitle(fmt.Sprintf("You have pressed key [%d]", ch))
}

func OnMouseMoveEvent(w *glumby.Window, x, y float64) {
	log.Printf("Mouse at (%f, %f)\n", x, y)
}

func OnMouseButtonEvent(w *glumby.Window, button glumby.MouseButton, action glumby.Action, mod glumby.ModifierKey) {
	log.Printf("Mouse button %d, action = %d, mod = %d\n", button, action, mod)
}

func OnScrollEvent(w *glumby.Window, xoff, yoff float64) {
	log.Printf("Scroll x = %f, y = %f\n", xoff, yoff)
	distance_z += float32(yoff)
}
