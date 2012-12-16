package main

// import "fmt"
import "log"

import "github.com/go-gl/gl"
import "github.com/go-gl/glu"
import "github.com/go-gl/glfw"

func main() {

	if err := glfw.Init(); err != nil {
		log.Fatal(err.Error())
	}

	if err := glfw.OpenWindow(800, 600, 8, 8, 8, 8, 32, 0, glfw.Windowed); err != nil {
		glfw.Terminate()
		log.Fatal(err.Error())
	}

	glfw.SetWindowTitle("Landscapes")
	glfw.SetSwapInterval(1)

	m := GenerateMap(160, 160, 16)
	m.BuildVertices()

	gl.Enable(gl.LIGHT0)
	gl.Enable(gl.LIGHTING)
	gl.Lightfv(gl.LIGHT0, gl.POSITION, []float32{0, 1, 0.2, 0})
	gl.Lightfv(gl.LIGHT0, gl.AMBIENT, []float32{0.0, 0.0, 0.0, 1})
	gl.Lightfv(gl.LIGHT0, gl.DIFFUSE, []float32{0.75, 0.75, 0.75, 1})
	gl.Lightfv(gl.LIGHT0, gl.SPECULAR, []float32{1, 1, 1, 1})

	gl.ShadeModel(gl.SMOOTH)
	gl.ClearColor(0.1, 0.05, 0.0, 1.0)

	far := 4096.0
	fov := 60.0

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	glu.Perspective(fov, 800.0/600, 1.0, far)
	gl.MatrixMode(gl.MODELVIEW)

	rot := float32(0.0)

	for glfw.WindowParam(glfw.Opened) == 1 {

		rot += 0.125

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.LoadIdentity()

		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(gl.LEQUAL)

		cx := float32(m.width*m.gridSize) * 0.5
		cy := float32(m.maxHeight) * 0.15
		cz := float32(m.depth*m.gridSize) * 0.5

		gl.Translatef(0, 0, -2000)
		gl.Rotatef(30, 1, 0, 0)
		gl.Rotatef(rot, 0, 1, 0)
		gl.Translatef(-cx, -cy, -cz)

		m.Draw()
		glfw.SwapBuffers()
	}

	glfw.Terminate()

}
